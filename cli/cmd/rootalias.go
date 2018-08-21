package cmd

import (
	"log"
	"os"
)

// checkRootAlias compares the first argument provided in the CLI with a list of
// subcmds and aliases. If no match is found, the first alias of rootCmd is added.
func checkRootAlias() {
	l := len(rootCmd.Aliases)
	if l == 0 {
		return
	}
	if l > 1 {
		log.Printf("rootCmd.Aliases should contain a single string. '%s' is used.\n", rootCmd.Aliases[0])
	}
	if len(os.Args) > 1 {
		for _, v := range append(nonRootSubCmds(), []string{"--help", "--version"}...) {
			if os.Args[1] == v {
				return
			}
		}
	}
	os.Args = append([]string{os.Args[0], rootCmd.Aliases[0]}, os.Args[1:]...)
}

// nonRootSubCmds traverses the list of subcommands of rootCmd and returns a string
// slice containing the names and aliases of all the subcmds, except the one defined
// in the Aliases field of rootCmd.
func nonRootSubCmds() (l []string) {
	for _, c := range rootCmd.Commands() {
		isAlias := false
		for _, a := range append(c.Aliases, c.Name()) {
			if a == rootCmd.Aliases[0] {
				isAlias = true
				break
			}
		}
		if !isAlias {
			l = append(l, c.Name())
			l = append(l, c.Aliases...)
		}
	}
	return
}
