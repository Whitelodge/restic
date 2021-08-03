package main

import (
	"github.com/spf13/cobra"
)

var cmdApiServer = &cobra.Command{
	Use:   "server",
	Short: "Start the api server",
	Long: `
The "server" command will start the api server to access restic functionality remotely

EXIT STATUS
===========

Exit status is 0 if the command was successful, and non-zero if there was any error.
`,
	DisableAutoGenTag: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runApiServer(cmd, globalOptions, args)
	},
}

func init() {
	cmdRoot.AddCommand(cmdApiServer)
}

func runApiServer(cmd *cobra.Command, opts GlobalOptions, args []string) error {
	return startServer("10173")
}
