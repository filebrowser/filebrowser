package cmd

import (
	"bytes"
	"fmt"
	"os"
	"path"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

func init() {
	rootCmd.AddCommand(docsCmd)
	docsCmd.Flags().String("out", "www/docs/cli", "directory to write the docs to")
}

var docsCmd = &cobra.Command{
	Use:    "docs",
	Hidden: true,
	Args:   cobra.NoArgs,
	RunE: func(cmd *cobra.Command, _ []string) error {
		outputDir, err := cmd.Flags().GetString("out")
		if err != nil {
			return err
		}

		tempDir, err := os.MkdirTemp(os.TempDir(), "filebrowser-docs-")
		if err != nil {
			return err
		}
		defer os.RemoveAll(tempDir)

		rootCmd.Root().DisableAutoGenTag = true

		err = doc.GenMarkdownTreeCustom(cmd.Root(), tempDir, func(_ string) string {
			return ""
		}, func(s string) string {
			return s
		})
		if err != nil {
			return err
		}

		entries, err := os.ReadDir(tempDir)
		if err != nil {
			return err
		}

		headerRegex := regexp.MustCompile(`(?m)^(##)(.*)$`)
		linkRegex := regexp.MustCompile(`\(filebrowser(.*)\.md\)`)

		fmt.Println("Generated Documents:")

		for _, entry := range entries {
			srcPath := path.Join(tempDir, entry.Name())
			dstPath := path.Join(outputDir, strings.ReplaceAll(entry.Name(), "_", "-"))

			data, err := os.ReadFile(srcPath)
			if err != nil {
				return err
			}

			data = headerRegex.ReplaceAll(data, []byte("#$2"))
			data = linkRegex.ReplaceAllFunc(data, func(b []byte) []byte {
				return bytes.ReplaceAll(b, []byte("_"), []byte("-"))
			})
			data = bytes.ReplaceAll(data, []byte("## SEE ALSO"), []byte("## See Also"))

			err = os.WriteFile(dstPath, data, 0666)
			if err != nil {
				return err
			}

			fmt.Println("- " + dstPath)
		}

		return nil
	},
}
