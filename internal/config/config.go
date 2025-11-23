package config

import "github.com/openkcm/common-sdk/pkg/commoncfg"

type Config struct {
	commoncfg.BaseConfig `mapstructure:",squash"`
}
