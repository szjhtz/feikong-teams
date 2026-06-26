package main

import (
	"context"
	modelproviders "fkteams/internal/adapters/model/providers"
	clicommands "fkteams/internal/adapters/transport/cli/commands"
	agents "fkteams/internal/app/agent/catalog"
	apptools "fkteams/internal/app/tools"
	bootstrapruntimes "fkteams/internal/bootstrap/runtimes"
	bootstraptools "fkteams/internal/bootstrap/tools"
	runtimeport "fkteams/internal/ports/runtime"
	modelregistry "fkteams/internal/runtime/model"
	"log"
	"os"

	"github.com/pterm/pterm"
)

func init() {
	log.SetFlags(log.LstdFlags | log.Lmicroseconds | log.Llongfile)
}

func main() {
	runtimeDefaults, err := bootstrapruntimes.NewDefaults()
	if err != nil {
		pterm.Error.Println(err)
		os.Exit(1)
	}
	toolRegistry, err := bootstraptools.RegisterDefaults()
	if err != nil {
		pterm.Error.Println(err)
		os.Exit(1)
	}
	ctx := runtimeport.WithEngine(context.Background(), runtimeDefaults.Engine)
	ctx = runtimeport.WithInterruptRuntime(ctx, runtimeDefaults.Interrupt)
	ctx = modelregistry.WithRegistry(ctx, runtimeDefaults.ModelRegistry)
	ctx = modelproviders.WithRegistry(ctx, runtimeDefaults.ModelProviderRegistry)
	ctx = apptools.WithRegistry(ctx, toolRegistry)
	ctx = agents.WithRegistry(ctx, agents.NewRegistry())
	if err := clicommands.Root().Run(ctx, os.Args); err != nil {
		pterm.Error.Println(err)
	}
}
