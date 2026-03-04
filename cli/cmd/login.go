package cmd

import (
	"fmt"

	"github.com/openkcm/krypton/pkg/auth"
	"github.com/openkcm/krypton/pkg/auth/provider"
	"github.com/openkcm/krypton/pkg/auth/store"
	"github.com/spf13/cobra"
)

func loginCmd() *cobra.Command {
	var authenticator []byte
	var authProvider auth.Provider
	var authStore auth.Store

	cmd := &cobra.Command{
		Use:   "login",
		Short: "Login to Krypton",
		Long:  "Authenticate with the Krypton server to obtain access.",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			// Declare authenticator from flags.
			// Declare auth provider and store based on flags (currently using no-op implementations).
			authProvider = &provider.NoOp{}
			authStore = &store.NoOp{}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			assertion, err := authStore.Get()
			if err != nil {
				return err
			}

			validationResult, err := authProvider.Validate(assertion)
			if err != nil {
				return err
			}

			if validationResult == auth.Valid {
				fmt.Println("Already logged in.")
				return nil
			}

			assertion, err = authProvider.Verify(authenticator)
			if err != nil {
				return err
			}

			err = authStore.Store(assertion)
			if err != nil {
				return err
			}

			fmt.Println("Login successful.")
			return nil
		},
	}

	return cmd
}
