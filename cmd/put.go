/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"log"

	"github.com/gammazero/workerpool"
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
		bucketName, remoteDirPath, err := tinys3cli.ParseS3URI(uriStr)
		if err != nil {
			log.Fatal(err)
		}

		client := tinys3cli.CreateClient()
		n_jobs, err := cmd.Flags().GetInt("jobs")
		if err != nil {
			n_jobs = 4
		}
		wp := workerpool.New(n_jobs)
		uploader := tinys3cli.NewS3Uploader(client, wp)

		for _, localPath := range args[0 : argc-1] {
			uploader.Submit(localPath, remoteDirPath, bucketName)
		}

		uploader.Wait()

	},
}

func init() {
	rootCmd.AddCommand(putCmd)
	putCmd.Flags().IntP("jobs", "j", 4, "max parallel jobs")

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// putCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// putCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
