package module

import (
	"context"

	"github.com/spf13/cobra"
)

type EmbeddedModule interface {
	Name() string
	Init(cfg any, serveCmd *cobra.Command) error
	RunServe(ctxStartup, ctxShutdown context.Context, shutdown func()) error
}
