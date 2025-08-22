package cmd

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
)

// Version information - set by main.go from build-time variables
var (
	version   = "dev"
	gitCommit = "unknown"
	buildTime = "unknown"
)

// SetVersionInfo sets the version information from main.go
func SetVersionInfo(v, commit, time string) {
	version = v
	gitCommit = commit
	buildTime = time
}

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Display version information",
	Long: `Display detailed version information including:
- Version number (major.git-sha.timestamp format)
- Git commit hash
- Build timestamp
- Go version and platform information`,
	Run: func(cmd *cobra.Command, args []string) {
		showVersion(cmd)
	},
}

func showVersion(cmd *cobra.Command) {
	// Check if we want detailed output
	detailed, _ := cmd.Flags().GetBool("detailed")

	if detailed {
		fmt.Fprintf(cmd.OutOrStdout(), "lxc-go-cli version information:\n")
		fmt.Fprintf(cmd.OutOrStdout(), "  Version:    %s\n", version)
		fmt.Fprintf(cmd.OutOrStdout(), "  Git Commit: %s\n", gitCommit)
		fmt.Fprintf(cmd.OutOrStdout(), "  Build Time: %s\n", buildTime)
		fmt.Fprintf(cmd.OutOrStdout(), "  Go Version: %s\n", runtime.Version())
		fmt.Fprintf(cmd.OutOrStdout(), "  Platform:   %s/%s\n", runtime.GOOS, runtime.GOARCH)
	} else {
		fmt.Fprintf(cmd.OutOrStdout(), "lxc-go-cli %s\n", version)
	}
}

func init() {
	rootCmd.AddCommand(versionCmd)

	// Add detailed flag
	versionCmd.Flags().BoolP("detailed", "d", false, "Show detailed version information")
}
