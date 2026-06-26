package main

import (
	"context"
	clicommands "fkteams/internal/adapters/transport/cli/commands"
	bootstrapruntimes "fkteams/internal/bootstrap/runtimes"
	bootstraptools "fkteams/internal/bootstrap/tools"
	"log"
	"os"

	"github.com/pterm/pterm"
)

func init() {
	log.SetFlags(log.LstdFlags | log.Lmicroseconds | log.Llongfile)
}

func main() {
	if err := bootstrapruntimes.RegisterDefaults(); err != nil {
		pterm.Error.Println(err)
		os.Exit(1)
	}
	if err := bootstraptools.RegisterDefaults(); err != nil {
		pterm.Error.Println(err)
		os.Exit(1)
	}
	if err := clicommands.Root().Run(context.Background(), os.Args); err != nil {
		pterm.Error.Println(err)
	}
}
