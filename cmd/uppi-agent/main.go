package main

import (
	"log"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/janyksteenbeek/uppi-server-agent/internal/config"
	"github.com/janyksteenbeek/uppi-server-agent/internal/reporter"
	"github.com/janyksteenbeek/uppi-server-agent/internal/updater"
)

func main() {
	var cfg config.Config

	rootCmd := &cobra.Command{
		Use:   "uppi-agent [token]",
		Short: "Uppi Server Monitoring Agent",
		Long:  `A daemon for monitoring server metrics and reporting to Uppi monitoring service.`,
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			token := args[0]
			parts := strings.SplitN(token, ":", 2)
			if len(parts) != 2 {
				log.Fatal("Token must be in format {serverId}:{secret}")
			}

			cfg.ServerId = parts[0]
			cfg.Secret = parts[1]

			if cfg.ServerId == "" {
				log.Fatal("Server ID cannot be empty")
			}
			if cfg.Secret == "" {
				log.Fatal("Secret cannot be empty")
			}

			runDaemon(cfg)
		},
	}

	rootCmd.Flags().StringVar(&cfg.Instance, "instance", config.DefaultInstance, "Instance URL")
	rootCmd.Flags().BoolVar(&cfg.SkipUpdates, "skip-updates", false, "Skip automatic updates")
	rootCmd.Flags().IntVar(&cfg.IntervalMinutes, "interval-minutes", 1, "Reporting interval in minutes")

	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

func runDaemon(cfg config.Config) {
	log.Printf("Starting Uppi Agent v%s", config.Version)
	log.Printf("Instance: %s", cfg.Instance)
	log.Printf("Interval: %d minutes", cfg.IntervalMinutes)
	log.Printf("Skip Updates: %v", cfg.SkipUpdates)

	// Check for updates unless skipped
	if !cfg.SkipUpdates {
		updater.CheckForUpdates()
	}

	// Send initial metrics
	if err := reporter.SendMetrics(cfg); err != nil {
		log.Printf("Failed to send initial metrics: %v", err)
	} else {
		log.Println("Initial metrics sent successfully")
	}

	// Start monitoring loop
	ticker := time.NewTicker(time.Duration(cfg.IntervalMinutes) * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		if err := reporter.SendMetrics(cfg); err != nil {
			log.Printf("Failed to send metrics: %v", err)
		} else {
			log.Println("Metrics sent successfully")
		}
	}
}
