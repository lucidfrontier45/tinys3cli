package tinys3cli

import (
	"context"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func CreateClient() *s3.Client {
	// Load the Shared AWS Configuration (~/.aws/config)
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatal(err)
	}

	// Create an Amazon S3 service client
	return s3.NewFromConfig(cfg)

}

func ParseS3URI(uriStr string) (bucketName string, remotePath string, err error) {
	uri, err := url.Parse(uriStr)

	if err != nil {
		return "", "", err
	}

	if strings.ToLower(uri.Scheme) != "s3" {
		return "", "", fmt.Errorf("invalid scheme %s", uri.Scheme)
	}

	remotePath = ""
	if len(uri.Path) > 0 {
		remotePath = uri.Path[1:]
	}

	return uri.Host, remotePath, nil

}

func ListObjects(client *s3.Client, uriStr string) (*s3.ListObjectsV2Output, error) {
	bucketName, path, err := ParseS3URI(uriStr)

	if err != nil {
		return nil, err
	}

	// Get the first page of results for ListObjectsV2 for a bucket
	return client.ListObjectsV2(context.TODO(), &s3.ListObjectsV2Input{
		Bucket: aws.String(bucketName),
		Prefix: aws.String(path),
	})
}

func doUpload(client *s3.Client, localPath, name, remoteDirPath, bucketName string) error {
	fp, err := os.Open(localPath)
	if err != nil {
		return err
	}

	defer fp.Close()

	key := ""
	if len(remoteDirPath) > 0 {
		key = fmt.Sprintf("%s/%s", remoteDirPath, name)
	} else {
		key = name
	}
	_, err = client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket: &bucketName, Key: &key, Body: fp})
	return err
}

func UploadObjects(client *s3.Client, localPath, remoteDirPath, bucketName string) error {
	info, err := os.Stat(localPath)
	if err != nil {
		return err
	}

	if info.IsDir() {
		var wg sync.WaitGroup
		err = filepath.WalkDir(localPath, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				// handle possible path err, just in case...
				return err
			}

			if !d.IsDir() {
				wg.Add(1)
				go func(p string) {
					err2 := doUpload(client, p, p, remoteDirPath, bucketName)
					if err2 != nil {
						fmt.Printf("couldn't upload %s, %s", p, err2)
					}
					defer wg.Done()
				}(path)
			}

			return nil
		})
		wg.Wait()
		return err
	} else {
		return doUpload(client, localPath, info.Name(), remoteDirPath, bucketName)
	}
}

func doDownload(client *s3.Client, localPath, remotePath, bucketName string) error {
	output, err := client.GetObject(context.TODO(), &s3.GetObjectInput{Bucket: &bucketName, Key: &remotePath})
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
	log.Printf("written %d bytes tp %s", n, localPath)

	return err
}

func DownloadObjects(client *s3.Client, localPath, remotePath, bucketName string, recursive bool) error {
	if recursive {
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
			log.Printf("key=%s size=%d", aws.ToString(object.Key), object.Size)
			dirPath, fileName := path.Split(*object.Key)
			dirPath = path.Join(localPath, dirPath)
			err = os.MkdirAll(dirPath, os.ModePerm)
			if err != nil {
				log.Printf("error on %s, %s", *object.Key, err)
			}
			filePath := path.Join(dirPath, fileName)
			err = doDownload(client, filePath, *object.Key, bucketName)
			if err != nil {
				log.Printf("error on %s, %s", *object.Key, err)
			}
		}

	} else {
		splt := strings.Split(remotePath, "/")
		filename := splt[len(splt)-1]
		var destPath string
		info, err := os.Stat(localPath)
		if err == nil && info.IsDir() {
			destPath = path.Join(localPath, filename)
		} else {
			destPath = localPath
		}
		return doDownload(client, destPath, remotePath, bucketName)
	}

	return nil
}
