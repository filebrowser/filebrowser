package cmd

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(docsCmd)
	docsCmd.Flags().StringP("path", "p", "./docs", "path to save the docs")
}

func printToc(names []string) {
	for i, name := range names {
		name = strings.TrimSuffix(name, filepath.Ext(name))
		name = strings.Replace(name, "-", " ", -1)
		names[i] = name
	}

	sort.Strings(names)

	toc := ""
	for _, name := range names {
		toc += "* [" + name + "](cli/" + strings.Replace(name, " ", "-", -1) + ".md)\n"
	}

	fmt.Println(toc)
}

var docsCmd = &cobra.Command{
	Use:    "docs",
	Hidden: true,
	Args:   cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		dir := mustGetString(cmd, "path")
		generateDocs(rootCmd, dir)
		names := []string{}

		err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if err != nil || info.IsDir() {
				return err
			}

			if !strings.HasPrefix(info.Name(), "filebrowser") {
				return nil
			}

			names = append(names, info.Name())
			return nil
		})

		checkErr(err)
		printToc(names)
	},
}

func generateDocs(cmd *cobra.Command, dir string) {
	for _, c := range cmd.Commands() {
		if !c.IsAvailableCommand() || c.IsAdditionalHelpTopicCommand() {
			continue
		}

		generateDocs(c, dir)
	}

	basename := strings.Replace(cmd.CommandPath(), " ", "-", -1) + ".md"
	filename := filepath.Join(dir, basename)
	f, err := os.Create(filename)
	checkErr(err)
	defer f.Close()
	generateMarkdown(cmd, f)
}

func generateMarkdown(cmd *cobra.Command, w io.Writer) {
	cmd.InitDefaultHelpCmd()
	cmd.InitDefaultHelpFlag()

	buf := new(bytes.Buffer)
	name := cmd.CommandPath()

	short := cmd.Short
	long := cmd.Long
	if len(long) == 0 {
		long = short
	}

	buf.WriteString("# " + name + "\n\n")
	buf.WriteString(short + "\n\n")
	buf.WriteString("## Synopsis\n\n")
	buf.WriteString(long + "\n\n")

	if cmd.Runnable() {
		buf.WriteString(fmt.Sprintf("```\n%s\n```\n\n", cmd.UseLine()))
	}

	if len(cmd.Example) > 0 {
		buf.WriteString("## Examples\n\n")
		buf.WriteString(fmt.Sprintf("```\n%s\n```\n\n", cmd.Example))
	}

	printOptions(buf, cmd, name)
	_, err := buf.WriteTo(w)
	checkErr(err)
}

func printOptions(buf *bytes.Buffer, cmd *cobra.Command, name string) {
	flags := cmd.NonInheritedFlags()
	flags.SetOutput(buf)
	if flags.HasAvailableFlags() {
		buf.WriteString("## Options\n\n```\n")
		flags.PrintDefaults()
		buf.WriteString("```\n\n")
	}

	parentFlags := cmd.InheritedFlags()
	parentFlags.SetOutput(buf)
	if parentFlags.HasAvailableFlags() {
		buf.WriteString("## Options inherited from parent commands\n\n```\n")
		parentFlags.PrintDefaults()
		buf.WriteString("```\n")
	}
}
