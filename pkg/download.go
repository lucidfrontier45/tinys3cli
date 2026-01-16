/*
Copyright © 2022 Du Shiqiao <lucidfrontier.45@gmail.com>
*/
package tinys3cli

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"strings"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func doDownload(
	client *s3.Client,
	localPath, remotePath, bucketName string,
	versionId string,
) error {
	var output *s3.GetObjectOutput
	var err error
	if versionId != "" {
		output, err = client.GetObject(
			context.TODO(),
			&s3.GetObjectInput{Bucket: &bucketName, Key: &remotePath, VersionId: &versionId},
		)
	} else {
		output, err = client.GetObject(
			context.TODO(),
			&s3.GetObjectInput{Bucket: &bucketName, Key: &remotePath},
		)
	}
	if err != nil {
		return err
	}
	defer func() {
		if cerr := output.Body.Close(); cerr != nil {
			err = errors.Join(err, cerr)
		}
	}()

	fp, err := os.Create(localPath)
	if err != nil {
		return err
	}
	defer func() {
		if cerr := fp.Close(); cerr != nil {
			err = errors.Join(err, cerr)
		}
	}()

	n, err := io.Copy(fp, output.Body)
	log.Printf("written %d bytes to %s", n, localPath)

	return err
}

var dirOnceMap sync.Map

func calcRemoteDirPrefix(remotePath string) (string, int) {
	remotePath = strings.TrimSuffix(remotePath, "/")
	splt := strings.Split(remotePath, "/")
	n := len(splt)
	if n > 0 {
		prefix := strings.Join(splt[:n-1], "/")
		return prefix, len(prefix)
	}
	return "", 0
}

func ensureDir(dirPath string) error {
	type onceValue struct {
		once sync.Once
		err  error
	}
	val, _ := dirOnceMap.LoadOrStore(dirPath, &onceValue{})
	ov := val.(*onceValue)
	ov.once.Do(func() {
		ov.err = os.MkdirAll(dirPath, os.ModePerm)
	})
	return ov.err
}

// Downloader handles S3 download operations.
type Downloader struct {
	*baseWorker
}

// NewDownloader creates a new S3 downloader with the specified number of jobs.
func NewDownloader(n_jobs int) (Downloader, error) {
	bw, err := newBaseWorker(n_jobs)
	if err != nil {
		return Downloader{}, err
	}
	return Downloader{baseWorker: bw}, nil
}

// Submit queues a download job for the given S3 path to the local path.
func (downloader *Downloader) Submit(
	localPath, remotePath, bucketName string,
	recursive bool,
	versionId string,
) error {
	client := downloader.client
	wp := downloader.wp

	// strip final slash
	remotePath = strings.TrimSuffix(remotePath, "/")

	if recursive {
		_, prefixLen := calcRemoteDirPrefix(remotePath)

		info, err := os.Stat(localPath)
		if err == nil && !info.IsDir() {
			return fmt.Errorf("cannot make directory, %s is a file", localPath)
		}

		if err := ValidatePath(localPath, ""); err != nil {
			return fmt.Errorf("invalid local path: %w", err)
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

			relPath := (*obj.Key)[prefixLen:]
			if err := ValidatePath(relPath, localPath); err != nil {
				return fmt.Errorf("invalid remote path %q: %w", *obj.Key, err)
			}

			wp.Submit(func() {
				dirPath, fileName := path.Split(*obj.Key)
				dirPath = path.Join(localPath, dirPath[prefixLen:])
				err = ensureDir(dirPath)
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
		if err := ValidatePath(filename, ""); err != nil {
			return fmt.Errorf("invalid remote path: %w", err)
		}

		var destPath string
		info, err := os.Stat(localPath)
		if err == nil && info.IsDir() {
			destPath = path.Join(localPath, filename)
		} else {
			destPath = localPath
		}

		if err := ValidatePath(destPath, ""); err != nil {
			return fmt.Errorf("invalid destination path: %w", err)
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
