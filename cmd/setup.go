package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/quexten/goldwarden/agent/systemauth"
	"github.com/quexten/goldwarden/browserbiometrics"
	"github.com/spf13/cobra"
)

func setupPolkit() {
	file, err := os.OpenFile("/tmp/goldwarden-policy", os.O_CREATE, 0644)
	if err != nil {
		panic(err)
	}
	file.WriteString(systemauth.POLICY)
	file.Close()

	command := exec.Command("pkexec", "mv", "/tmp/goldwarden-policy", "/usr/share/polkit-1/actions/com.quexten.goldwarden.policy")
	err = command.Run()
	if err != nil {
		panic(err)
	}

	os.Remove("/tmp/goldwarden-policy")
}

var polkitCmd = &cobra.Command{
	Use:   "polkit",
	Short: "Sets up polkit",
	Long:  "Sets up polkit",
	Run: func(cmd *cobra.Command, args []string) {
		setupPolkit()
	},
}

const SYSTEMD_SERVICE = `[Unit]
Description="Goldwarden daemon"

[Service]
ExecStart=BINARY_PATH daemonize

[Install]
WantedBy=default.target`

func setupSystemd() {
	file, err := os.OpenFile("/tmp/goldwarden.service", os.O_CREATE, 0644)
	if err != nil {
		panic(err)
	}

	path, err := os.Executable()
	if err != nil {
		panic(err)
	}

	file.WriteString(strings.ReplaceAll(SYSTEMD_SERVICE, "BINARY_PATH", path))
	file.Close()

	command := exec.Command("pkexec", "mv", "/tmp/goldwarden.service", "/etc/systemd/system/goldwarden.service")
	err = command.Run()
	if err != nil {
		panic(err)
	}

	os.Remove("/tmp/goldwarden.service")
}

var systemdCmd = &cobra.Command{
	Use:   "systemd",
	Short: "Sets up systemd autostart",
	Long:  "Sets up systemd autostart",
	Run: func(cmd *cobra.Command, args []string) {
		setupSystemd()
	},
}

var browserbiometricsCmd = &cobra.Command{
	Use:   "browserbiometrics",
	Short: "Sets up browser biometrics",
	Long:  "Sets up browser biometrics",
	Run: func(cmd *cobra.Command, args []string) {
		err := browserbiometrics.DetectAndInstallBrowsers()
		if err != nil {
			fmt.Println("Error: " + err.Error())
		} else {
			fmt.Println("Done.")
		}
	},
}

var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Sets up Goldwarden integrations",
	Long:  "Sets up Goldwarden integrations",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

func init() {
	rootCmd.AddCommand(setupCmd)
	setupCmd.AddCommand(polkitCmd)
	setupCmd.AddCommand(systemdCmd)
	setupCmd.AddCommand(browserbiometricsCmd)
}
