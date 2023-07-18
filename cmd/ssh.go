/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"github.com/atotto/clipboard"
	"github.com/quexten/goldwarden/client"
	"github.com/quexten/goldwarden/ipc"
	"github.com/spf13/cobra"
)

var sshCmd = &cobra.Command{
	Use:   "ssh",
	Short: "Commands for managing SSH keys",
	Long:  `Commands for managing SSH keys.`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

// runCmd represents the run command
var sshAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Runs a command with environment variables from your vault",
	Long: `Runs a command with environment variables from your vault.
	The variables are stored as a secure note. Consult the documentation for more information.`,
	Run: func(cmd *cobra.Command, args []string) {
		name, _ := cmd.Flags().GetString("name")
		copyToClipboard, _ := cmd.Flags().GetBool("clipboard")

		result, err := client.SendToAgent(ipc.CreateSSHKeyRequest{
			Name: name,
		})
		if err != nil {
			println("Error: " + err.Error())
			println("Is the daemon running?")
			return
		}

		switch result.(type) {
		case ipc.CreateSSHKeyResponse:
			response := result.(ipc.CreateSSHKeyResponse)
			fmt.Println(response.Digest)

			if copyToClipboard {
				clipboard.WriteAll(string(response.Digest))
			}
			break
		case ipc.ActionResponse:
			println("Error: " + result.(ipc.ActionResponse).Message)
			return
		}
	},
}

var listSSHCmd = &cobra.Command{
	Use:   "list",
	Short: "Lists all SSH keys in your vault",
	Long:  `Lists all SSH keys in your vault.`,
	Run: func(cmd *cobra.Command, args []string) {
		result, err := client.SendToAgent(ipc.GetSSHKeysRequest{})
		if err != nil {
			println("Error: " + err.Error())
			println("Is the daemon running?")
			return
		}

		switch result.(type) {
		case ipc.GetSSHKeysResponse:
			response := result.(ipc.GetSSHKeysResponse)
			for _, key := range response.Keys {
				fmt.Println(key)
			}
			break
		case ipc.ActionResponse:
			println("Error: " + result.(ipc.ActionResponse).Message)
			return
		}
	},
}

func init() {
	rootCmd.AddCommand(sshCmd)
	sshCmd.AddCommand(sshAddCmd)
	sshAddCmd.PersistentFlags().String("name", "", "")
	sshAddCmd.MarkFlagRequired("name")
	sshAddCmd.PersistentFlags().Bool("clipboard", false, "Copy the public key to the clipboard")
	sshCmd.AddCommand(listSSHCmd)
}
