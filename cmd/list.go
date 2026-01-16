/*
Copyright © 2022 Du Shiqiao <lucidfrontier.45@gmail.com>
*/
package cmd

import (
	"fmt"

	tinys3cli "github.com/lucidfrontier45/tinys3cli/pkg"
	"github.com/spf13/cobra"
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list S3URI",
	Short: "list objects in S3 bucket",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return fmt.Errorf("requires S3 URI argument")
		}
		client, err := tinys3cli.CreateClient()
		if err != nil {
			return fmt.Errorf("failed to create client: %w", err)
		}
		output, err := tinys3cli.ListObjects(client, args[0])
		if err != nil {
			return fmt.Errorf("failed to list objects: %w", err)
		}

		fmt.Println("first page results:")
		for _, object := range output.Contents {
			fmt.Printf("key=%s size=%d\n", *object.Key, object.Size)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(listCmd)

	// Here you will define your flags and configuration settings.
}
