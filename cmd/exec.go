// Copyright © 2017 Alexandre Bourget <alex@bourget.cc>

package cmd

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/abourget/secrets-bridge/pkg/agentfwd"
	"github.com/spf13/cobra"
)

// execCmd represents the exec command
var execCmd = &cobra.Command{
	Use:   "exec",
	Short: "Execute a command, with SSH-Agent available and secrets injected in env vars.",
	Long: `Example:

Run my-command.sh with the ENVVAR corresponding to the value of 'key'. my-command.sh can also invoke SSH-Agent through ssh.

    secrets-bridge -c $BRIDGE_CONF exec -e ENVVAR=key -- ./my-command.sh

This invocation will use the bridge config at ~/.bridge-conf by default:

    secrets-bridge exec ssh secured-host
`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			log.Fatalln("no command to execute specified, see examples with --help")
		}

		c, err := newClient(bridgeConf)
		if err != nil {
			log.Fatalln(err)
		}

		log.Println("secrets-bridge: server connection successful")

		if !disableSSHAgentForwarding {
			go func() {
				log.Println("secrets-bridge: Setting up SSH-Agent forwarder...")
				err := agentfwd.ListenAndServeSSHAgentForwarder(c.SSHAgentWebsocketURL(), "https://localhost", c.ClientTLSConfig())
				if err != nil {
					log.Fatalln("couldn't setup SSH-Agent forwarder:", err)
				}
			}()
			os.Setenv("SSH_AUTH_SOCK", agentfwd.UnixSocket)
			time.Sleep(25 * time.Millisecond)
		}

		env := os.Environ()

		for _, envVarKey := range envVars {
			varParts := strings.Split(envVarKey, "=")
			if len(varParts) != 2 {
				log.Fatalf("'-e' env var %q spec malformed\n", varParts[0])
			}

			// Fetch the `key`..
			secret, err := c.GetSecretString(varParts[1])
			if err != nil {
				log.Fatalln("failed fetching secret:", err)
			}

			env = append(env, fmt.Sprintf("%s=%s", varParts[0], secret))
		}

		program := args[0]
		var arguments []string
		if len(args) > 1 {
			arguments = args[1:]
		}
		subprocess := exec.Command(program, arguments...)
		subprocess.Env = env
		subprocess.Stdin = os.Stdin
		subprocess.Stdout = os.Stdout
		subprocess.Stderr = os.Stderr

		log.Printf("secrets-bridge: Calling subprocess with: %q\n", args)
		subErr := subprocess.Run()
		log.Printf("secrets-bridge: Call returned, exited with: %s\n", subErr)

		if !disableSSHAgentForwarding {
			os.Remove(agentfwd.UnixSocket)
		}

		// now reflect the subprocess error...
		if subErr != nil {
			os.Exit(199)
		}
	},
}

var disableSSHAgentForwarding bool
var envVars []string

func init() {
	RootCmd.AddCommand(execCmd)

	execCmd.Flags().StringVarP(&bridgeConf, "bridge-conf", "c", "", "Base64-encoded Bridge `configuration`.")
	execCmd.Flags().BoolVarP(&disableSSHAgentForwarding, "no-ssh-agent", "A", false, "Disable SSH-Agent relay. Otherwise sets SSH_AUTH_SOCK in the subprocess call and listens for incoming Agent calls.")
	execCmd.Flags().StringSliceVarP(&envVars, "env", "e", []string{}, "Inject secret as environment variable. Use the `ENV_VAR=secret-key` syntax.")
}
