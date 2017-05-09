// +build !linux

package cmd

import (
	"log"

	"github.com/spf13/cobra"
)

func serveDaemonized(cmd *cobra.Command, args []string) {
	if daemonize == "" {
		serve(cmd, args)
		return
	} else {
		log.Fatalln("Daemonization not support on this platform.")
	}
}

func detachFromParent() {}
