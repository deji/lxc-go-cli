/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package main

import "github.com/deji/lxc-go-cli/cmd"

// Version information set at build time via ldflags
var (
	Version   = "dev"     // Will be overridden at build time
	GitCommit = "unknown" // Will be overridden at build time
	BuildTime = "unknown" // Will be overridden at build time
)

func main() {
	// Set version info for the CLI
	cmd.SetVersionInfo(Version, GitCommit, BuildTime)
	cmd.Execute()
}
