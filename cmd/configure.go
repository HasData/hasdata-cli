package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/hasdata-com/hasdata-cli/internal/config"
	"github.com/spf13/cobra"
)

func init() {
	var apiKey string
	var endpoint string
	var nonInteractive bool
	cmd := &cobra.Command{
		Use:   "configure",
		Short: "Configure the CLI (stores API key in ~/.hasdata/config.yaml)",
		RunE: func(c *cobra.Command, _ []string) error {
			path, err := config.Path()
			if err != nil {
				return err
			}
			existing, err := config.Load(path)
			if err != nil {
				return err
			}
			if !nonInteractive && apiKey == "" {
				fmt.Fprintf(c.OutOrStderr(), "API key [%s]: ", maskKey(existing.APIKey))
				r := bufio.NewReader(os.Stdin)
				line, _ := r.ReadString('\n')
				line = strings.TrimSpace(line)
				if line != "" {
					apiKey = line
				}
			}
			if apiKey != "" {
				existing.APIKey = apiKey
			}
			if endpoint != "" {
				existing.Endpoint = endpoint
			}
			if err := config.Save(path, existing); err != nil {
				return err
			}
			fmt.Fprintf(c.OutOrStdout(), "Wrote %s\n", path)
			return nil
		},
	}
	cmd.Flags().StringVar(&apiKey, "api-key", "", "API key")
	cmd.Flags().StringVar(&endpoint, "endpoint", "", "override API endpoint")
	cmd.Flags().BoolVar(&nonInteractive, "non-interactive", false, "do not prompt for missing values")
	rootCmd.AddCommand(cmd)
}

func maskKey(k string) string {
	if k == "" {
		return "not set"
	}
	if len(k) <= 8 {
		return "****"
	}
	return k[:4] + "…" + k[len(k)-4:]
}
