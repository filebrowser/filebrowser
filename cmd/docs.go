package cmd

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

func init() {
	rootCmd.AddCommand(docsCmd)
	docsCmd.Flags().StringP("path", "p", "./docs", "path to save the docs")
}

func removeAll (dir string) error {
	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if strings.HasPrefix(info.Name(), "filebrowser") {
			return os.Remove(path)
		}

		return nil
	})
}

func printToc (names []string) {
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

		err := removeAll(dir)
		checkErr(err)
		err = doc.GenMarkdownTree(rootCmd, dir)
		checkErr(err)
		
		names := []string{}

		err = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if err != nil || info.IsDir() {
				return err
			}

			if !strings.HasPrefix(info.Name(), "filebrowser") {
				return nil
			}

			name := strings.Replace(info.Name(), "_", "-", -1)
			names = append(names, name)
			newPath := filepath.Join(dir, name)
			err = os.Rename(path, newPath)
			if err != nil {
				return err
			}
			path = newPath

			fd, err := os.OpenFile(path, os.O_RDONLY, os.ModePerm)
			if err != nil {
				return err
			}
			defer fd.Close()

			content := ""
			sc := bufio.NewScanner(fd)
			for sc.Scan() {
				txt := sc.Text()
				if txt == "### SEE ALSO" {
					break
				}
				content += txt + "\n"
			}
			if err := sc.Err(); err != nil {
				return err
			}

			content = strings.TrimSpace(content) + "\n"
			return ioutil.WriteFile(path, []byte(content), 077)
		})

		checkErr(err)
		printToc(names)
	},
}
