package common

import (
	"github.com/openkcm/common-sdk/pkg/commoncfg"
	"github.com/samber/oops"

	"github.com/openkcm/crypto-edge/internal/config"
)

//nolint:mnd
var defaultConfig = map[string]any{}

func LoadConfig(buildInfo string) (*config.Config, error) {
	cfg := &config.Config{}

	loader := commoncfg.NewLoader(
		cfg,
		commoncfg.WithDefaults(defaultConfig),
		commoncfg.WithPaths(
			"/etc/encrypto",
			"$HOME/.encrypto",
			".",
		),
	)

	err := loader.LoadConfig()
	if err != nil {
		return nil, oops.In("main").Wrapf(err, "failed to load config")
	}

	// Update Version
	err = commoncfg.UpdateConfigVersion(&cfg.BaseConfig, buildInfo)
	if err != nil {
		return nil, oops.In("main").
			Wrapf(err, "Failed to update the version configuration")
	}

	return cfg, nil
}
