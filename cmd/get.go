/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"log"

	tinys3cli "github.com/lucidfrontier45/tinys3cli/pkg"
	"github.com/spf13/cobra"
)

// getCmd represents the get command
var getCmd = &cobra.Command{
	Use:   "get S3URI localPath",
	Short: "download file or directory ",
	Run: func(cmd *cobra.Command, args []string) {
		var uriStr, localPath string
		uriStr = args[0]
		localPath = args[1]

		bucketName, remotePath, err := tinys3cli.ParseS3URI(uriStr)
		if err != nil {
			log.Fatal(err)
		}

		recursive, _ := cmd.Flags().GetBool("recursive")

		n_jobs, err := cmd.Flags().GetInt("jobs")
		if err != nil {
			n_jobs = 4
		}
		downloader := tinys3cli.NewS3Downloader(n_jobs)
		err = downloader.Submit(localPath, remotePath, bucketName, recursive)
		defer downloader.Wait()
		if err != nil {
			log.Fatal(err)
		}

	},
}

func init() {
	rootCmd.AddCommand(getCmd)
	getCmd.Flags().BoolP("recursive", "r", false, "download recursively")
	getCmd.Flags().IntP("jobs", "j", 4, "max parallel jobs")
	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// getCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// getCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
