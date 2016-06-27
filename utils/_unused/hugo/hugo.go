package hugo

import (
	"log"
	"os"

	"github.com/hacdias/caddy-hugo/config"
	"github.com/hacdias/caddy-hugo/tools/commands"
)

// Run is used to run the static website generator
func Run(c *config.Config.Config, force bool) {
	os.RemoveAll(c.Path + "public")

	// Prevent running if watching is enabled
	if b, pos := stringInSlice("--watch", c.Args); b && !force {
		if len(c.Args) > pos && c.Args[pos+1] != "false" {
			return
		}

		if len(c.Args) == pos+1 {
			return
		}
	}

	if err := commands.Run(c.Hugo, c.Args, c.Path); err != nil {
		log.Panic(err)
	}
}
