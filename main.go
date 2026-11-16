// Command girc is a simple terminal IRC client.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"girc/app"
	"girc/config"
)

func main() {
	configPath := flag.String("config", "config.yaml", "path to the configuration file")
	flag.Parse()

	if err := run(*configPath); err != nil {
		fmt.Fprintln(os.Stderr, "girc:", err)
		os.Exit(1)
	}
}

func run(configPath string) error {
	cfg, err := config.Load(configPath)
	if err != nil {
		return err
	}
	return app.Run(context.Background(), cfg)
}
