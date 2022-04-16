package tinys3cli

import (
	"context"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path"
	"path/filepath"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gammazero/workerpool"
)

func doUpload(client *s3.Client, localPath, name, remoteDirPath, bucketName string) error {
	fp, err := os.Open(localPath)
	if err != nil {
		return err
	}

	defer fp.Close()

	key := ""
	if len(remoteDirPath) > 0 {
		key = path.Join(remoteDirPath, name)
	} else {
		key = name
	}
	_, err = client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket: &bucketName, Key: &key, Body: fp})

	if err == nil {
		log.Printf("uploaded %s", localPath)
	}
	return err
}

type S3Uploader struct {
	client *s3.Client
	wp     *workerpool.WorkerPool
}

func NewS3Uploader(client *s3.Client, wp *workerpool.WorkerPool) S3Uploader {
	return S3Uploader{client: client, wp: wp}
}

func (uploader *S3Uploader) Wait() {
	uploader.wp.StopWait()
}

func (uploader *S3Uploader) Submit(localPath, remoteDirPath, bucketName string) error {
	info, err := os.Stat(localPath)
	if err != nil {
		return err
	}

	if info.IsDir() {
		err = filepath.WalkDir(localPath, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				// handle possible path err, just in case...
				return err
			}

			if !d.IsDir() {
				p := path
				uploader.wp.Submit(func() {
					err2 := doUpload(uploader.client, p, p, remoteDirPath, bucketName)
					if err2 != nil {
						fmt.Printf("couldn't upload %s, %s", p, err2)
					}
				})
			}

			return nil
		})
		return err
	} else {
		uploader.wp.Submit(func() {
			err2 := doUpload(uploader.client, localPath, info.Name(), remoteDirPath, bucketName)
			if err2 != nil {
				fmt.Printf("couldn't upload %s, %s", localPath, err2)
			}
		})
	}

	return nil
}
