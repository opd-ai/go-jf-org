package main

import (
	"os"

	"github.com/opd-ai/go-jf-org/cmd"
	"github.com/rs/zerolog/log"
)

const version = "0.1.0-dev"

func main() {
	if err := cmd.Execute(); err != nil {
		log.Error().Err(err).Msg("Command failed")
		os.Exit(1)
	}
}
