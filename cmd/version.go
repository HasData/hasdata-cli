package cmd

import (
	"encoding/json"
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
)

func init() {
	var jsonOut bool
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Print the version, commit, and build date",
		RunE: func(c *cobra.Command, _ []string) error {
			if jsonOut {
				enc := json.NewEncoder(c.OutOrStdout())
				enc.SetIndent("", "  ")
				return enc.Encode(map[string]string{
					"version": versionStr,
					"commit":  commitStr,
					"date":    dateStr,
					"os":      runtime.GOOS,
					"arch":    runtime.GOARCH,
				})
			}
			fmt.Fprintf(c.OutOrStdout(), "hasdata %s (%s) built %s %s/%s\n",
				versionStr, commitStr, dateStr, runtime.GOOS, runtime.GOARCH)
			return nil
		},
	}
	cmd.Flags().BoolVar(&jsonOut, "json", false, "emit JSON")
	rootCmd.AddCommand(cmd)
}
