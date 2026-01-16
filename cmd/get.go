/*
Copyright © 2022 Du Shiqiao <lucidfrontier.45@gmail.com>
*/
package cmd

import (
	"fmt"

	tinys3cli "github.com/lucidfrontier45/tinys3cli/pkg"
	"github.com/spf13/cobra"
)

// getCmd represents the get command
var getCmd = &cobra.Command{
	Use:   "get S3URI localPath",
	Short: "download file or directory ",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 2 {
			return fmt.Errorf("requires S3 URI and local path arguments")
		}
		var uriStr, localPath string
		uriStr = args[0]
		localPath = args[1]

		bucketName, remotePath, err := tinys3cli.ParseS3URI(uriStr)
		if err != nil {
			return fmt.Errorf("invalid S3 URI: %w", err)
		}

		recursive, _ := cmd.Flags().GetBool("recursive")

		n_jobs, err := cmd.Flags().GetInt("jobs")
		if err != nil {
			n_jobs = 4
		}

		downloader, err := tinys3cli.NewDownloader(n_jobs)
		if err != nil {
			return fmt.Errorf("failed to create downloader: %w", err)
		}

		versionId, err := cmd.Flags().GetString("version-id")
		if err != nil {
			versionId = ""
		}
		if versionId != "" && recursive {
			return fmt.Errorf("version ID cannot be specified when downloading recursively")
		}

		err = downloader.Submit(localPath, remotePath, bucketName, recursive, versionId)
		if err != nil {
			return fmt.Errorf("failed to submit download: %w", err)
		}
		downloader.Wait()
		if downloader.GetLastErr() != nil {
			return fmt.Errorf("download failed: %w", downloader.GetLastErr())
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(getCmd)
	getCmd.Flags().BoolP("recursive", "r", false, "download recursively")
	getCmd.Flags().IntP("jobs", "j", 4, "max parallel jobs")
	getCmd.Flags().StringP("version-id", "v", "", "file version ID")
	// Here you will define your flags and configuration settings.
}
