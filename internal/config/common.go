package config

import (
	"github.com/openkcm/common-sdk/pkg/commoncfg"
	"github.com/samber/oops"
)

//nolint:mnd
var defaultConfig = map[string]any{}

func LoadConfig(buildInfo string, paths ...string) (*Config, error) {
	cfg := &Config{}

	loader := commoncfg.NewLoader(
		cfg,
		commoncfg.WithDefaults(defaultConfig),
		commoncfg.WithPaths(paths...),
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
