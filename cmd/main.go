package main

import (
	"flag"
	"log"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/lucidfrontier45/tinys3cli"
)

func main() {
	flag.Parse()
	args := flag.Args()
	argc := len(args)
	client := tinys3cli.CreateClient()

	switch args[0] {
	case "ls":
		uriStr := args[1]
		output, err := tinys3cli.ListObjects(client, uriStr)
		if err != nil {
			log.Fatal(err)
		}

		log.Println("first page results:")
		for _, object := range output.Contents {
			log.Printf("key=%s size=%d", aws.ToString(object.Key), object.Size)
		}
	case "get":
		var uriStr, localPath string
		var recursive bool

		if args[1] == "-r" {
			uriStr = args[2]
			localPath = args[3]
			recursive = true
		} else {
			uriStr = args[1]
			localPath = args[2]
			recursive = false
		}

		bucketName, remotePath, err := tinys3cli.ParseS3URI(uriStr)
		if err != nil {
			log.Fatal(err)
		}

		err = tinys3cli.DownloadObjects(client, localPath, remotePath, bucketName, recursive)
		if err != nil {
			log.Fatal(err)
		}
	case "put":
		uriStr := args[argc-1]
		bucketName, remoteDirPath, err := tinys3cli.ParseS3URI(uriStr)
		if err != nil {
			log.Fatal(err)
		}

		var wg sync.WaitGroup
		for _, localPath := range args[1 : argc-1] {
			wg.Add(1)
			go func(p string) {
				defer wg.Done()
				err = tinys3cli.UploadObjects(client, p, remoteDirPath, bucketName)
				if err != nil {
					log.Printf("couldn't upload %s, %s", p, err)
				} else {
					log.Printf("uploaded %s to %s", p, uriStr)
				}
			}(localPath)
		}
		wg.Wait()
	default:
		log.Fatalf("invalid command %s", args[0])
	}

}
