package tinys3cli

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"strings"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gammazero/workerpool"
)

func doDownload(client *s3.Client, localPath, remotePath, bucketName string, versionId string) error {
	var output *s3.GetObjectOutput
	var err error
	if versionId != "" {
		output, err = client.GetObject(context.TODO(), &s3.GetObjectInput{Bucket: &bucketName, Key: &remotePath, VersionId: &versionId})
	} else {
		output, err = client.GetObject(context.TODO(), &s3.GetObjectInput{Bucket: &bucketName, Key: &remotePath})
	}
	if err != nil {
		return err
	}
	defer output.Body.Close()

	fp, err := os.Create(localPath)
	if err != nil {
		return err
	}
	defer fp.Close()

	n, err := io.Copy(fp, output.Body)
	log.Printf("written %d bytes to %s", n, localPath)

	return err
}

type S3Downloader struct {
	client    *s3.Client
	wp        *workerpool.WorkerPool
	mux       sync.Mutex
	lasterror error
}

func NewS3Downloader(n_jobs int) S3Downloader {
	return S3Downloader{client: CreateClient(), wp: workerpool.New(n_jobs), mux: sync.Mutex{}}
}

func (downloader *S3Downloader) GetLastErr() error {
	return downloader.lasterror
}

func (downloader *S3Downloader) SetLastErr(err error) {
	downloader.mux.Lock()
	defer downloader.mux.Unlock()
	downloader.lasterror = err
}

func (downloader *S3Downloader) Wait() {
	downloader.wp.StopWait()
}

func (downloader *S3Downloader) Submit(localPath, remotePath, bucketName string, recursive bool, versionId string) error {
	client := downloader.client
	wp := downloader.wp

	// strip final slash
	remotePath = strings.TrimSuffix(remotePath, "/")

	if recursive {
		splt := strings.Split(remotePath, "/")
		n := len(splt)
		remoteDirPrefix := ""
		if n > 0 {
			remoteDirPrefix = strings.Join(splt[:n-1], "/")
		}
		prefixLen := len(remoteDirPrefix)

		info, err := os.Stat(localPath)
		if err == nil && !info.IsDir() {
			return fmt.Errorf("cannot make directory, %s is a file", localPath)
		}

		listResult, err := client.ListObjectsV2(context.TODO(), &s3.ListObjectsV2Input{
			Bucket: aws.String(bucketName),
			Prefix: aws.String(remotePath),
		})
		if err != nil {
			return err
		}

		for _, object := range listResult.Contents {
			obj := object
			// skip directory
			if strings.HasSuffix(*obj.Key, "/") {
				continue
			}
			wp.Submit(func() {
				dirPath, fileName := path.Split(*obj.Key)
				dirPath = path.Join(localPath, dirPath[prefixLen:])
				downloader.mux.Lock()
				err = os.MkdirAll(dirPath, os.ModePerm)
				downloader.mux.Unlock()
				if err != nil {
					log.Printf("error on %s, %s", *obj.Key, err)
					downloader.SetLastErr(err)
					return
				}

				filePath := path.Join(dirPath, fileName)
				err = doDownload(client, filePath, *obj.Key, bucketName, "")
				if err != nil {
					log.Printf("error on %s, %s", *obj.Key, err)
					downloader.SetLastErr(err)
				}
			})
		}

	} else {
		_, filename := path.Split(remotePath)
		var destPath string
		info, err := os.Stat(localPath)
		if err == nil && info.IsDir() {
			destPath = path.Join(localPath, filename)
		} else {
			destPath = localPath
		}

		wp.Submit(func() {
			err := doDownload(client, destPath, remotePath, bucketName, versionId)
			if err != nil {
				log.Printf("error on %s, %s", remotePath, err)
				downloader.SetLastErr(err)
			}
		})
	}

	return nil
}
