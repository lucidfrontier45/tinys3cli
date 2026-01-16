/*
Copyright © 2022 Du Shiqiao <lucidfrontier.45@gmail.com>
*/
package cmd

import (
	"fmt"
	"log"

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

		n_jobs, _ := cmd.Flags().GetInt("jobs")
		if n_jobs <= 0 {
			envJobs := tinys3cli.GetWorkerCountFromEnv()
			if envJobs > 0 {
				n_jobs = envJobs
			} else {
				n_jobs = tinys3cli.GetDefaultWorkerCount()
			}
		}
		clamped, warning := tinys3cli.ValidateWorkerCount(n_jobs)
		if warning != "" {
			log.Printf(warning, n_jobs, tinys3cli.GetMaxWorkerCount(), clamped)
			n_jobs = clamped
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

		noLocalPathCheck, _ := cmd.Flags().GetBool("no-local-path-check")

		err = downloader.Submit(
			localPath,
			remotePath,
			bucketName,
			recursive,
			versionId,
			noLocalPathCheck,
		)
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
	getCmd.Flags().Bool("no-local-path-check", false, "disable local path validation")
	// Here you will define your flags and configuration settings.
}
