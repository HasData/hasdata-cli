package cmd

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"time"

	"github.com/creativeprojects/go-selfupdate"
	"github.com/hasdata-com/hasdata-cli/internal/config"
	"github.com/spf13/cobra"
)

const (
	githubOwner = "hasdata-com"
	githubRepo  = "hasdata-cli"
)

func init() {
	var check bool
	var force bool
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Check for updates and install the latest release",
		RunE: func(c *cobra.Command, _ []string) error {
			ctx := c.Context()
			latest, found, err := selfupdate.DetectLatest(ctx, selfupdate.ParseSlug(githubOwner+"/"+githubRepo))
			if err != nil {
				return &networkError{err: fmt.Errorf("check latest: %w", err)}
			}
			if !found {
				fmt.Fprintln(c.OutOrStdout(), "No releases found.")
				return nil
			}
			current := versionStr
			if check {
				fmt.Fprintf(c.OutOrStdout(), "current: %s\nlatest:  %s\n", current, latest.Version())
				return nil
			}
			if !force && latest.LessOrEqual(current) {
				fmt.Fprintf(c.OutOrStdout(), "Already up to date (%s).\n", current)
				return nil
			}
			exe, err := os.Executable()
			if err != nil {
				return err
			}
			fmt.Fprintf(c.OutOrStdout(), "Updating %s → %s (%s/%s)...\n", current, latest.Version(), runtime.GOOS, runtime.GOARCH)
			if err := selfupdate.UpdateTo(ctx, latest.AssetURL, latest.AssetName, exe); err != nil {
				return &networkError{err: fmt.Errorf("update: %w", err)}
			}
			fmt.Fprintf(c.OutOrStdout(), "Updated to %s.\n", latest.Version())

			// Persist the latest known version so the notifier stays quiet.
			path, _ := config.Path()
			if cfg, err := config.Load(path); err == nil {
				cfg.LatestKnownVer = latest.Version()
				cfg.LastUpdateCheck = time.Now()
				_ = config.Save(path, cfg)
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&check, "check", false, "only report whether a newer version is available")
	cmd.Flags().BoolVar(&force, "force", false, "reinstall even if the current version is up to date")
	rootCmd.AddCommand(cmd)
}

// maybeNotifyUpdate prints a one-line notice to stderr if a newer version was
// detected in a prior check. Does NOT perform a network call here — the actual
// polling happens at most once per 24h inside a goroutine started from Client().
func maybeNotifyUpdate(stderr interface {
	Write(p []byte) (int, error)
}) {
	if loadedCfg == nil || !config.ShouldCheckUpdates(loadedCfg) {
		return
	}
	if loadedCfg.LatestKnownVer == "" || versionStr == "dev" {
		// Schedule a background refresh but don't block.
		go refreshLatestKnownVersion()
		return
	}
	if loadedCfg.LatestKnownVer == versionStr {
		return
	}
	// best-effort semver: simple string compare is enough for the notice.
	if loadedCfg.LatestKnownVer > versionStr {
		fmt.Fprintf(stderr.(interface {
			Write([]byte) (int, error)
		}), "hasdata: new version %s available (current %s) — run 'hasdata update'\n",
			loadedCfg.LatestKnownVer, versionStr)
	}
	if time.Since(loadedCfg.LastUpdateCheck) > 24*time.Hour {
		go refreshLatestKnownVersion()
	}
}

func refreshLatestKnownVersion() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	latest, found, err := selfupdate.DetectLatest(ctx, selfupdate.ParseSlug(githubOwner+"/"+githubRepo))
	if err != nil || !found {
		return
	}
	path, err := config.Path()
	if err != nil {
		return
	}
	cfg, err := config.Load(path)
	if err != nil {
		return
	}
	cfg.LatestKnownVer = latest.Version()
	cfg.LastUpdateCheck = time.Now()
	_ = config.Save(path, cfg)
}
