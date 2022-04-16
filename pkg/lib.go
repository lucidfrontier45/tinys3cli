package tinys3cli

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/url"
	"os"
	"path"
	"strings"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

var Version = "0.1.0"

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
	log.Printf("written %d bytes to %s", n, localPath)

	return err
}

func DownloadObjects(client *s3.Client, localPath, remotePath, bucketName string, recursive bool) error {
	if recursive {
		println("recursive")
	} else {
		println("single")
	}
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

		var mux sync.Mutex
		var wg sync.WaitGroup

		for _, object := range listResult.Contents {
			wg.Add(1)
			go func(obj types.Object) {
				defer wg.Done()
				dirPath, fileName := path.Split(*obj.Key)
				dirPath = path.Join(localPath, dirPath)
				mux.Lock()
				err = os.MkdirAll(dirPath, os.ModePerm)
				mux.Unlock()
				if err != nil {
					log.Printf("error on %s, %s", *obj.Key, err)
				}

				filePath := path.Join(dirPath, fileName)
				err = doDownload(client, filePath, *obj.Key, bucketName)
				if err != nil {
					log.Printf("error on %s, %s", *obj.Key, err)
				}
			}(object)
		}
		wg.Wait()

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
