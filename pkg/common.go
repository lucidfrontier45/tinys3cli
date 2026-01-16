/*
Copyright © 2022 Du Shiqiao <lucidfrontier.45@gmail.com>
*/
package tinys3cli

import (
	"context"
	"fmt"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// Version holds the application version.
var Version = "0.3.3"

// CreateClient creates an S3 client using the default AWS configuration.
// It loads credentials and region from ~/.aws/config and ~/.aws/credentials.
func CreateClient() (*s3.Client, error) {
	// Load the Shared AWS Configuration (~/.aws/config)
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return nil, err
	}

	// Create an Amazon S3 service client
	return s3.NewFromConfig(cfg), nil
}

// ParseS3URI parses an S3 URI string and returns the bucket name and remote path.
// The URI format must be s3://bucket/path/to/object.
func ParseS3URI(uriStr string) (bucketName, remotePath string, err error) {
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

	if err := ValidatePath(remotePath, "", false); err != nil {
		return "", "", fmt.Errorf("invalid remote path: %w", err)
	}

	return uri.Host, remotePath, nil
}

// ListObjects lists objects in an S3 bucket with the given prefix.
// Returns the first page of results (up to 1000 objects).
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

// ValidatePath checks if a path is safe and stays within the base directory.
// It rejects paths containing directory traversal sequences (../ or ..\)
// and ensures the resolved path does not escape the base directory.
// When noLocalPathCheck is true, all validations are skipped for local paths.
func ValidatePath(path, baseDir string, noLocalPathCheck bool) error {
	// Skip all validations if local path checking is disabled
	if noLocalPathCheck {
		return nil
	}
	path = filepath.ToSlash(path)
	baseDir = filepath.ToSlash(baseDir)

	if strings.Contains(path, "../") || strings.Contains(path, "..\\") {
		return fmt.Errorf("path %q contains directory traversal sequence", path)
	}

	if strings.HasPrefix(path, "..") {
		return fmt.Errorf("path %q starts with directory traversal", path)
	}

	if baseDir != "" {
		absBase, err := filepath.Abs(baseDir)
		if err != nil {
			return fmt.Errorf("failed to resolve base directory: %w", err)
		}
		absPath := filepath.Join(absBase, path)
		relPath, err := filepath.Rel(absBase, absPath)
		if err != nil {
			return fmt.Errorf("path validation failed: %w", err)
		}
		if strings.HasPrefix(relPath, "..") {
			return fmt.Errorf("path %q escapes base directory %q", path, baseDir)
		}
	}

	return nil
}
