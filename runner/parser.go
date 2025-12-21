package runner

import (
	"github.com/filebrowser/filebrowser/v2/settings"
)

// ParseCommand parses the command taking in account if the current
// instance uses a shell to run the commands or just calls the binary
// directly.
func ParseCommand(s *settings.Settings, raw string) (command []string, name string, err error) {
	name, args, err := SplitCommandAndArgs(raw)
	if err != nil {
		return
	}

	if len(s.Shell) == 0 || s.Shell[0] == "" {
		command = append(command, name)
		command = append(command, args...)
	} else {
		command = append(command, s.Shell...)
		command = append(command, raw)
	}

	return command, name, nil
}
