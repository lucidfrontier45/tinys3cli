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

	"github.com/aws/aws-sdk-go-v2/service/s3"
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

type WalkError struct {
	Errors []error
}

func (e *WalkError) Error() string {
	if len(e.Errors) == 0 {
		return ""
	}
	return fmt.Sprintf("%d errors occurred during walk", len(e.Errors))
}

func calcLocalDirPrefix(localPath string) (string, int) {
	localPath = strings.TrimSuffix(localPath, "/")
	splt := strings.Split(localPath, "/")
	n := len(splt)
	if n > 0 {
		prefix := strings.Join(splt[:n-1], "/")
		return prefix, len(prefix)
	}
	return "", 0
}

// Uploader handles S3 upload operations.
type Uploader struct {
	*baseWorker
}

// NewUploader creates a new S3 uploader with the specified number of jobs.
func NewUploader(n_jobs int) (Uploader, error) {
	bw, err := newBaseWorker(n_jobs)
	if err != nil {
		return Uploader{}, err
	}
	return Uploader{baseWorker: bw}, nil
}

// Submit queues an upload job for the given local path to the S3 bucket.
func (uploader *Uploader) Submit(localPath, remoteDirPath, bucketName string) error {
	info, err := os.Stat(localPath)
	if err != nil {
		return err
	}

	client := uploader.client
	wp := uploader.wp

	// strip final slash
	localPath = strings.TrimSuffix(localPath, "/")

	if info.IsDir() {
		_, prefixLen := calcLocalDirPrefix(localPath)

		walkErr := &WalkError{Errors: make([]error, 0)}

		walkDirErr := filepath.WalkDir(
			localPath,
			func(path string, d fs.DirEntry, walkErrIn error) error {
				if walkErrIn != nil {
					walkErr.Errors = append(walkErr.Errors, walkErrIn)
					return nil
				}

				if !d.IsDir() {
					path := path
					wp.Submit(func() {
						err2 := doUpload(client, path, path[prefixLen:], remoteDirPath, bucketName)
						if err2 != nil {
							log.Printf("couldn't upload %s, %s", path, err2)
							uploader.SetLastErr(err2)
						}
					})
				}

				return nil
			},
		)
		if walkDirErr != nil {
			return walkErr
		}
	} else {
		wp.Submit(func() {
			err2 := doUpload(client, localPath, info.Name(), remoteDirPath, bucketName)
			if err2 != nil {
				log.Printf("couldn't upload %s, %s", localPath, err2)
				uploader.SetLastErr(err2)
			}
		})
	}

	return nil
}
