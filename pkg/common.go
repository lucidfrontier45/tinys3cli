package tinys3cli

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

var Version = "0.3.3"

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
