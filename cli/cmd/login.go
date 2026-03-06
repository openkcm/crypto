package cmd

import (
	"errors"
	"fmt"

	"github.com/openkcm/krypton/pkg/authn"
	"github.com/openkcm/krypton/pkg/authn/provider"
	"github.com/openkcm/krypton/pkg/authn/store"
	"github.com/spf13/cobra"
)

func loginCmd() *cobra.Command {
	var credentials *authn.Credentials
	var authProvider authn.Provider
	var authStore authn.Store

	cmd := &cobra.Command{
		Use:   "login",
		Short: "Login to Krypton",
		Long:  "Authenticate with the Krypton server to obtain access.",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			// Declare credentials from flags (currently credentials are ignored in the no-op provider).

			// Declare authn provider and store based on flags (currently using no-op and fs implementations by default).
			authProvider = &provider.NoOp{}
			store, err := store.NewFS()
			if err != nil {
				return err
			}
			authStore = store

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			token, err := authStore.Get(ctx)
			if err != nil && !errors.Is(err, authn.ErrTokenNotFound) {
				return err
			}

			if token != nil {
				validationResult, err := authProvider.Validate(ctx, token)
				if err != nil {
					return err
				}

				if validationResult.Status == authn.ValidStatus {
					fmt.Println("Already logged in.")
					return nil
				}
			}

			token, err = authProvider.Verify(ctx, credentials)
			if err != nil {
				return err
			}

			err = authStore.Store(ctx, token)
			if err != nil {
				return err
			}

			fmt.Println("Login successful.")
			return nil
		},
	}

	return cmd
}
