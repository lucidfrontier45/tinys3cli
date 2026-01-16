/*
Copyright © 2022 Du Shiqiao <lucidfrontier.45@gmail.com>
*/
package cmd

import (
	"fmt"

	tinys3cli "github.com/lucidfrontier45/tinys3cli/pkg"
	"github.com/spf13/cobra"
)

// putCmd represents the put command
var putCmd = &cobra.Command{
	Use:   "put localPath S3URI",
	Short: "upload file or directory",
	RunE: func(cmd *cobra.Command, args []string) error {
		argc := len(args)
		if argc < 1 {
			return fmt.Errorf("no source files specified")
		}

		uriStr := args[argc-1]
		bucketName, remoteDirPath, err := tinys3cli.ParseS3URI(uriStr)
		if err != nil {
			return fmt.Errorf("invalid S3 URI: %w", err)
		}

		n_jobs, err := cmd.Flags().GetInt("jobs")
		if err != nil {
			n_jobs = 4
		}
		uploader, err := tinys3cli.NewS3Uploader(n_jobs)
		if err != nil {
			return fmt.Errorf("failed to create uploader: %w", err)
		}

		for _, localPath := range args[0 : argc-1] {
			err = uploader.Submit(localPath, remoteDirPath, bucketName)
			if err != nil {
				return fmt.Errorf("failed to submit upload: %w", err)
			}
		}

		uploader.Wait()

		if uploader.GetLastErr() != nil {
			return fmt.Errorf("upload failed: %w", uploader.GetLastErr())
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(putCmd)
	putCmd.Flags().IntP("jobs", "j", 4, "max parallel jobs")

	// Here you will define your flags and configuration settings.
}
