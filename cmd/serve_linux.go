package cmd

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	daemon "github.com/sevlyar/go-daemon"
	"github.com/spf13/cobra"
)

func serveDaemonized(cmd *cobra.Command, args []string) {
	if daemonize == "" {
		serve(cmd, args)
		return
	}

	ctx := daemon.Context{
		LogFileName: daemonize,
	}
	child, err := ctx.Reborn()
	if err != nil {
		log.Fatalln("Error setting up daemon:", err)
	}

	if child == nil {
		log.Printf("Will daemonize upon successful socket listening. pid=%d\n", os.Getpid())

		serve(cmd, args)
		// WARN: with all those `Fatalln` in `serve`, it's possible
		// `ctx.Release()` won't get called.
		ctx.Release()
		return
	}

	log.Printf("Setting up secrets-bridge daemon, pid=%d\n", child.Pid)

	// Parent, waiting for Listen signal, then exits.
	childReady := make(chan os.Signal, 5)
	signal.Notify(childReady, syscall.SIGUSR1)
	select {
	case <-childReady:
		log.Println("Setup successfully.")
		return
	case <-time.After(5 * time.Second):
		log.Fatalln("Failed. Timed out")
	}
}

func detachFromParent() {
	if daemonize != "" {
		time.Sleep(100 * time.Millisecond)
		proc, err := os.FindProcess(os.Getppid())
		if err != nil {
			log.Fatalln("Couldn't find parent process id:", err)
		}

		log.Println("Daemonizing")

		if err := proc.Signal(syscall.SIGUSR1); err != nil {
			log.Fatalln("couldn't send SIGUSR1 signal to parent:", err)
		}
	}
}
