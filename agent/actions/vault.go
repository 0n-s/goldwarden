package actions

import (
	"context"
	"fmt"

	"github.com/quexten/goldwarden/agent/bitwarden"
	"github.com/quexten/goldwarden/agent/bitwarden/crypto"
	"github.com/quexten/goldwarden/agent/config"
	"github.com/quexten/goldwarden/agent/sockets"
	"github.com/quexten/goldwarden/agent/systemauth"
	"github.com/quexten/goldwarden/agent/systemauth/pinentry"
	"github.com/quexten/goldwarden/agent/vault"

	"github.com/quexten/goldwarden/ipc/messages"
)

func handleUnlockVault(request messages.IPCMessage, cfg *config.Config, vault *vault.Vault, callingContext *sockets.CallingContext) (response messages.IPCMessage, err error) {
	if !cfg.HasPin() {
		response, err = messages.IPCMessageFromPayload(messages.ActionResponse{
			Success: false,
			Message: "No pin set",
		})
		if err != nil {
			panic(err)
		}

		return
	}

	if !cfg.IsLocked() {
		response, err = messages.IPCMessageFromPayload(messages.ActionResponse{
			Success: true,
			Message: "Unlocked",
		})
		if err != nil {
			panic(err)
		}

		return
	}

	err = cfg.TryUnlock(vault)
	if err != nil {
		response, err = messages.IPCMessageFromPayload(messages.ActionResponse{
			Success: false,
			Message: "wrong pin: " + err.Error(),
		})
		if err != nil {
			panic(err)
		}

		return
	}

	if cfg.IsLoggedIn() {
		token, err := cfg.GetToken()
		if err == nil {
			if token.AccessToken != "" {
				ctx := context.Background()
				bitwarden.RefreshToken(ctx, cfg)
				token, err := cfg.GetToken()
				userSymmkey, err := cfg.GetUserSymmetricKey()
				if err != nil {
					fmt.Println(err)
				}
				safeUserSymmkey, err := crypto.SymmetricEncryptionKeyFromBytes(userSymmkey)
				if err != nil {
					fmt.Println(err)
				}
				err = bitwarden.DoFullSync(context.WithValue(ctx, bitwarden.AuthToken{}, token.AccessToken), vault, cfg, &safeUserSymmkey, true)
				if err != nil {
					fmt.Println(err)
				}
			}
		}
	}

	response, err = messages.IPCMessageFromPayload(messages.ActionResponse{
		Success: true,
	})
	if err != nil {
		panic(err)
	}

	return
}

func handleLockVault(request messages.IPCMessage, cfg *config.Config, vault *vault.Vault, callingContext *sockets.CallingContext) (response messages.IPCMessage, err error) {
	if !cfg.HasPin() {
		response, err = messages.IPCMessageFromPayload(messages.ActionResponse{
			Success: false,
			Message: "No pin set",
		})
		if err != nil {
			panic(err)
		}

		return
	}

	if cfg.IsLocked() {
		response, err = messages.IPCMessageFromPayload(messages.ActionResponse{
			Success: true,
			Message: "Locked",
		})
		if err != nil {
			panic(err)
		}

		return
	}

	cfg.Lock()
	vault.Clear()
	vault.Keyring.Lock()

	response, err = messages.IPCMessageFromPayload(messages.ActionResponse{
		Success: true,
	})
	if err != nil {
		panic(err)
	}

	return
}

func handleWipeVault(request messages.IPCMessage, cfg *config.Config, vault *vault.Vault, callingContext *sockets.CallingContext) (response messages.IPCMessage, err error) {
	cfg.Purge()
	cfg.WriteConfig()
	vault.Clear()

	response, err = messages.IPCMessageFromPayload(messages.ActionResponse{
		Success: true,
	})
	if err != nil {
		panic(err)
	}

	return
}

func handleUpdateVaultPin(request messages.IPCMessage, cfg *config.Config, vault *vault.Vault, callingContext *sockets.CallingContext) (response messages.IPCMessage, err error) {
	pin, err := pinentry.GetPassword("Pin Change", "Enter your desired pin")
	if err != nil {
		response, err = messages.IPCMessageFromPayload(messages.ActionResponse{
			Success: false,
			Message: err.Error(),
		})
		if err != nil {
			return messages.IPCMessage{}, err
		} else {
			return response, nil
		}
	}
	cfg.UpdatePin(pin, true)

	response, err = messages.IPCMessageFromPayload(messages.ActionResponse{
		Success: true,
	})

	return
}

func handlePinStatus(request messages.IPCMessage, cfg *config.Config, vault *vault.Vault, callingContext *sockets.CallingContext) (response messages.IPCMessage, err error) {
	var pinStatus string
	if cfg.HasPin() {
		pinStatus = "enabled"
	} else {
		pinStatus = "disabled"
	}

	response, err = messages.IPCMessageFromPayload(messages.ActionResponse{
		Success: true,
		Message: pinStatus,
	})

	return
}

func init() {
	AgentActionsRegistry.Register(messages.MessageTypeForEmptyPayload(messages.UnlockVaultRequest{}), handleUnlockVault)
	AgentActionsRegistry.Register(messages.MessageTypeForEmptyPayload(messages.LockVaultRequest{}), handleLockVault)
	AgentActionsRegistry.Register(messages.MessageTypeForEmptyPayload(messages.WipeVaultRequest{}), handleWipeVault)
	AgentActionsRegistry.Register(messages.MessageTypeForEmptyPayload(messages.UpdateVaultPINRequest{}), ensureBiometricsAuthorized(systemauth.AccessVault, handleUpdateVaultPin))
	AgentActionsRegistry.Register(messages.MessageTypeForEmptyPayload(messages.GetVaultPINRequest{}), handlePinStatus)
}
