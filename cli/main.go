package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/unknown420x/Roblox-Group-Finder/internal"
)

var version = "dev"

func main() {
	cfg, err := internal.LoadConfig()
	if err != nil {
		fatal(err)
	}

	showVersion := flag.Bool(
		"version",
		false,
		"Show version",
	)

	command := "scan"
	if len(os.Args) > 1 && !isFlag(os.Args[1]) {
		command = os.Args[1]
		os.Args = os.Args[1:]
	}
	switch command {
	case "version":
		fmt.Println(version)
		return
	case "reset":
		if err := internal.ResetConfig(); err != nil {
			fatal(err)
		}
		if err := internal.ResetState(); err != nil {
			fatal(err)
		}
		fmt.Println("Saved configuration and state reset.")
		return
	case "config":
		configureMode(&cfg)
		return
	case "scan":
	default:
		printUsage()
		os.Exit(2)
	}
	fs := flag.NewFlagSet("scan", flag.ExitOnError)
	workers := fs.Int("workers", cfg.Workers, "Concurrent workers")
	rps := fs.Int("rps", cfg.RPS, "Maximum requests per second")
	batch := fs.Int("batch-size", cfg.BatchSize, "IDs per request")
	minID := fs.Int("min-id", cfg.MinID, "Minimum group ID")
	maxID := fs.Int("max-id", cfg.MaxID, "Maximum group ID")
	timeout := fs.String("timeout", cfg.Timeout, "HTTP timeout")
	webhook := fs.String("webhook", cfg.WebhookURL, "Discord webhook URL")
	unique := fs.Bool("unique", cfg.Unique, "Do not repeat IDs")
	fs.Parse(os.Args[1:])
	cfg.Workers = *workers
	cfg.RPS = *rps
	cfg.BatchSize = *batch
	cfg.MinID = *minID
	cfg.MaxID = *maxID
	cfg.Timeout = *timeout
	cfg.WebhookURL = *webhook
	cfg.Unique = *unique
	if *showVersion {
		fmt.Printf("Roblox Group Finder by: Samulxz v. %s\n", version)
		return
	}
	if err := cfg.Validate(); err != nil {
		fatal(err)
	}
	if err := internal.SaveConfig(cfg); err != nil {
		fatal(err)
	}
	timeoutDuration, err := cfg.TimeoutDuration()
	if err != nil {
		fatal(err)
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	client := internal.NewRobloxClient(timeoutDuration)
	var hook *internal.WebhookClient
	if cfg.WebhookURL != "" {
		hook = internal.NewWebhookClient(cfg.WebhookURL, timeoutDuration)
	}
	scanner := internal.NewScanner(client, hook, cfg)
	fmt.Printf("Group Finder by: Samulxz v. %s\nWorkers: %d | RPS: %d | Batch: %d\nRange: %d-%d | Unique: %t\n\n", version, cfg.Workers, cfg.RPS, cfg.BatchSize, cfg.MinID, cfg.MaxID, cfg.Unique)
	go statsLoop(ctx, scanner)
	scanner.Run(ctx)
	fmt.Println("\nStopped.")
}

func isFlag(s string) bool { return len(s) > 1 && s[0] == '-' }
func printUsage()          { fmt.Println("Usage: groupfinder [scan|config|reset|version] [flags]") }
func configureMode(cfg *internal.Config) {
	fmt.Println("Configuration (press Enter to keep current value)")
	cfg.Workers = askInt("Workers", cfg.Workers)
	cfg.RPS = askInt("Requests/sec", cfg.RPS)
	cfg.BatchSize = askInt("Batch size", cfg.BatchSize)
	cfg.MinID = askInt("Minimum group ID", cfg.MinID)
	cfg.MaxID = askInt("Maximum group ID", cfg.MaxID)
	cfg.Timeout = askString("Timeout", cfg.Timeout)
	cfg.WebhookURL = askString("Webhook", cfg.WebhookURL)
	if err := cfg.Validate(); err != nil {
		fatal(err)
	}
	if err := internal.SaveConfig(*cfg); err != nil {
		fatal(err)
	}
	fmt.Println("Configuration saved.")
}
func askInt(label string, current int) int {
	var v string
	fmt.Printf("%s [%d]: ", label, current)
	fmt.Scanln(&v)
	if v == "" {
		return current
	}
	var n int
	if _, err := fmt.Sscan(v, &n); err != nil {
		return current
	}
	return n
}
func askString(label, current string) string {
	display := current
	if display == "" {
		display = "not configured"
	}
	fmt.Printf("%s [%s]: ", label, display)
	var v string
	fmt.Scanln(&v)
	if v == "" {
		return current
	}
	return v
}
func statsLoop(ctx context.Context, s *internal.Scanner) {
	start := time.Now()
	t := time.NewTicker(time.Second)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			st := s.Snapshot()
			elapsed := time.Since(start).Seconds()
			rate := float64(st.Requests)
			if elapsed > 0 {
				rate /= elapsed
			}
			fmt.Printf("\r\033[2KRequests: %-7d | Groups: %-8d | RPS: %5.2f | Hits: %-5d | 429: %-5d | Errors: %-5d", st.Requests, st.Checked, rate, st.Hits, st.RateLimited, st.Errors)
		}
	}
}
func fatal(err error) { fmt.Fprintln(os.Stderr, "error:", err); os.Exit(1) }
