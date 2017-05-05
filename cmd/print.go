// Copyright © 2017 Alexandre Bourget <alex@bourget.cc>

package cmd

import (
	"log"
	"os"

	"github.com/spf13/cobra"
)

// printCmd represents the print command
var printCmd = &cobra.Command{
	Use:   "print",
	Short: "Print a single secret key to stdout.",
	Long: `Example:

secrets-bridge print a-key

secrets-bridge print -c [BASE64-bridge-conf] hello-secret
`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			log.Fatalln("specify one, and only one, key to print")
		}

		c, err := newClient(bridgeConf)
		if err != nil {
			log.Fatalln(err)
		}

		secret, err := c.GetSecret(args[0])
		if err != nil {
			log.Fatalln("failed fetching secret:", err)
		}

		_, _ = os.Stdout.Write(secret)
	},
}

func init() {
	RootCmd.AddCommand(printCmd)

	printCmd.Flags().StringVarP(&bridgeConf, "bridge-conf", "c", "", "Base64-encoded Bridge `configuration`.")
}
