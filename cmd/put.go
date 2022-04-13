/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"log"
	"sync"

	tinys3cli "github.com/lucidfrontier45/tinys3cli/pkg"
	"github.com/spf13/cobra"
)

// putCmd represents the put command
var putCmd = &cobra.Command{
	Use:   "put localfile1 [localfile2] ... S3URI",
	Short: "upload file or directory",
	Run: func(cmd *cobra.Command, args []string) {
		argc := len(args)
		uriStr := args[argc-1]
		client := tinys3cli.CreateClient()
		bucketName, remoteDirPath, err := tinys3cli.ParseS3URI(uriStr)
		if err != nil {
			log.Fatal(err)
		}

		var wg sync.WaitGroup
		for _, localPath := range args[0 : argc-1] {
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
	},
}

func init() {
	rootCmd.AddCommand(putCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// putCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// putCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
