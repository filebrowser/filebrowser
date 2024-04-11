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
	"github.com/spf13/pflag"
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
	Run: func(cmd *cobra.Command, _ []string) {
		dir := mustGetString(cmd.Flags(), "path")
		generateDocs(rootCmd, dir)
		names := []string{}

		err := filepath.Walk(dir, func(_ string, info os.FileInfo, err error) error {
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
	if long == "" {
		long = short
	}

	buf.WriteString("---\ndescription: " + short + "\n---\n\n")
	buf.WriteString("# " + name + "\n\n")
	buf.WriteString("## Synopsis\n\n")
	buf.WriteString(long + "\n\n")

	if cmd.Runnable() {
		_, _ = fmt.Fprintf(buf, "```\n%s\n```\n\n", cmd.UseLine())
	}

	if cmd.Example != "" {
		buf.WriteString("## Examples\n\n")
		_, _ = fmt.Fprintf(buf, "```\n%s\n```\n\n", cmd.Example)
	}

	printOptions(buf, cmd)
	_, err := buf.WriteTo(w)
	checkErr(err)
}

func generateFlagsTable(fs *pflag.FlagSet, buf io.StringWriter) {
	_, _ = buf.WriteString("| Name | Shorthand | Usage |\n")
	_, _ = buf.WriteString("|------|-----------|-------|\n")

	fs.VisitAll(func(f *pflag.Flag) {
		_, _ = buf.WriteString("|" + f.Name + "|" + f.Shorthand + "|" + f.Usage + "|\n")
	})
}

func printOptions(buf *bytes.Buffer, cmd *cobra.Command) {
	flags := cmd.NonInheritedFlags()
	flags.SetOutput(buf)
	if flags.HasAvailableFlags() {
		buf.WriteString("## Options\n\n")
		generateFlagsTable(flags, buf)
		buf.WriteString("\n")
	}

	parentFlags := cmd.InheritedFlags()
	parentFlags.SetOutput(buf)
	if parentFlags.HasAvailableFlags() {
		buf.WriteString("### Inherited\n\n")
		generateFlagsTable(parentFlags, buf)
		buf.WriteString("\n")
	}
}
