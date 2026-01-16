/*
Copyright © 2022 Du Shiqiao <lucidfrontier.45@gmail.com>
*/
package tinys3cli

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gammazero/workerpool"
)

func doUpload(client *s3.Client, localPath, name, remoteDirPath, bucketName string) error {
	fp, err := os.Open(localPath)
	if err != nil {
		return err
	}

	defer func() {
		if cerr := fp.Close(); cerr != nil {
			err = errors.Join(err, cerr)
		}
	}()

	key := ""
	if len(remoteDirPath) > 0 {
		key = path.Join(remoteDirPath, name)
	} else {
		key = name
	}
	_, err = client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket: &bucketName, Key: &key, Body: fp,
	})

	if err == nil {
		log.Printf("uploaded %s", localPath)
	}
	return err
}

// S3Uploader handles S3 upload operations.
type S3Uploader struct {
	client    *s3.Client
	wp        *workerpool.WorkerPool
	mux       sync.Mutex
	lasterror error
}

// NewS3Uploader creates a new S3 uploader with the specified number of jobs.
func NewS3Uploader(n_jobs int) (S3Uploader, error) {
	client, err := CreateClient()
	if err != nil {
		return S3Uploader{}, err
	}
	return S3Uploader{client: client, wp: workerpool.New(n_jobs), mux: sync.Mutex{}}, nil
}

// GetLastErr returns the last error encountered during upload.
func (uploader *S3Uploader) GetLastErr() error {
	return uploader.lasterror
}

// SetLastErr sets the last error.
func (uploader *S3Uploader) SetLastErr(err error) {
	uploader.mux.Lock()
	defer uploader.mux.Unlock()
	uploader.lasterror = err
}

// Wait blocks until all upload jobs are complete.
func (uploader *S3Uploader) Wait() {
	uploader.wp.StopWait()
}

// Submit queues an upload job for the given local path to the S3 bucket.
func (uploader *S3Uploader) Submit(localPath, remoteDirPath, bucketName string) error {
	info, err := os.Stat(localPath)
	if err != nil {
		return err
	}

	client := uploader.client
	wp := uploader.wp

	// strip final slash
	localPath = strings.TrimSuffix(localPath, "/")

	if info.IsDir() {
		splt := strings.Split(localPath, "/")
		n := len(splt)
		localDirPrefix := ""
		if n > 0 {
			localDirPrefix = strings.Join(splt[:n-1], "/")
		}
		prefixLen := len(localDirPrefix)

		err = filepath.WalkDir(localPath, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				// handle possible path err, just in case...
				return err
			}

			if !d.IsDir() {
				path := path
				wp.Submit(func() {
					err2 := doUpload(client, path, path[prefixLen:], remoteDirPath, bucketName)
					if err2 != nil {
						fmt.Printf("couldn't upload %s, %s", path, err2)
						uploader.SetLastErr(err2)
					}
				})
			}

			return nil
		})
		return err
	} else {
		wp.Submit(func() {
			err2 := doUpload(client, localPath, info.Name(), remoteDirPath, bucketName)
			if err2 != nil {
				fmt.Printf("couldn't upload %s, %s", localPath, err2)
				uploader.SetLastErr(err2)
			}
		})
	}

	return nil
}
