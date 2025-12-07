package cmd

import (
	"os"
	"time"

	"github.com/opd-ai/go-jf-org/internal/config"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var (
	cfgFile string
	cfg     *config.Config
	verbose bool
)

// rootCmd represents the base command
var rootCmd = &cobra.Command{
	Use:   "go-jf-org",
	Short: "A tool to organize media files for Jellyfin server",
	Long: `go-jf-org organizes disorganized media files (movies, TV shows, music, books) 
into a clean, Jellyfin-compatible directory structure.

It extracts metadata from filenames and files, enriches it with external APIs,
and safely moves files without ever deleting anything.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// Set up logging
		zerolog.TimeFieldFormat = time.RFC3339
		if verbose {
			zerolog.SetGlobalLevel(zerolog.DebugLevel)
			log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
		} else {
			zerolog.SetGlobalLevel(zerolog.InfoLevel)
			log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
		}

		// Load configuration
		var err error
		cfg, err = config.Load(cfgFile)
		if err != nil {
			log.Warn().Err(err).Msg("Failed to load config, using defaults")
			cfg = config.DefaultConfig()
		}
	},
}

// Execute runs the root command
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.go-jf-org/config.yaml)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
}
