package runner

import (
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/filebrowser/filebrowser/v2/settings"
)

func GlobExpand(args []string, cwd string) ([]string, error) {
	var expandedArgs []string

	for _, arg := range args {

		//Sets the path from the current working directory to the argument supplied.
		arg_path := "." + cwd + arg
		//Expands the globs in the filesystem.
		matches, err := filepath.Glob(arg_path)

		if err != nil {
			return nil, err
		}
		//No match means the argument is fine so just appended to the end.
		if len(matches) == 0 {
			expandedArgs = append(expandedArgs, arg)
		} else {
			for _, match := range matches {
				if err != nil {
					return nil, err
				}

				//now we need to remove what was appended by cwd
				if err != nil {
					return nil, err
				}

				// The match path is longer than the cwd path
				splitCwdLength := len(strings.Split(cwd, "/")) - 2 //Split includes the whitspaces here, so the expected length is two less.
				matchSplitLength := len(strings.Split(match, "/"))

				//if the split diff is more than one, then add an extra number to help with slicing the array.
				//If this step does not exist, then an empty string will be returned on join, if dir depth is 1
				if matchSplitLength-splitCwdLength > 2 {
					splitCwdLength++
				}
				newMatch := strings.Join(strings.Split(match, "/")[splitCwdLength:], "/")
				//Add to the end of the args
				expandedArgs = append(expandedArgs, newMatch)
			}
		}
	}

	return expandedArgs, nil

}

// ParseCommand parses the command taking in account if the current
// instance uses a shell to run the commands or just calls the binary
// directly.
func ParseCommand(s *settings.Settings, raw string, cwd string) ([]string, error) {
	var command []string

	if len(s.Shell) == 0 {
		cmd, args, err := SplitCommandAndArgs(raw)
		if err != nil {
			return nil, err
		}

		// TODO General regex file expansions for commands like ls as well.
		args, err = GlobExpand(args, cwd)

		if err != nil {
			return nil, err
		}

		_, err = exec.LookPath(cmd)
		if err != nil {
			return nil, err
		}

		command = append(command, cmd)
		command = append(command, args...)
	} else {
		command = append(s.Shell, raw) //nolint:gocritic
	}

	return command, nil
}
