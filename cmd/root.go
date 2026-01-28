package cmd

import "github.com/spf13/cobra"

var (
	RootCmd = &cobra.Command{
		Use:   "krypton",
		Short: "A secure key management and kryptongraphic operations service.",
		Long: `Crypto is a modular key management service designed to securely generate, store,
	and use kryptongraphic keys for applications and microservices.
	
	It provides support for encryption, decryption, signing, hashing, secure key
	storage, certificate-based operations, and protocol-specific interfaces such as
	KMIP. Crypto also integrates with observability tooling, supports multiple auth
	methods, and runs as a collection of embedded services that can be enabled or
	disabled at runtime.
	
	Use the available subcommands to start services, run migrations, inspect system
	information, or integrate Crypto into your infrastructure workflows.`,
	}
)
