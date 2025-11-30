package config

import "github.com/openkcm/common-sdk/pkg/commoncfg"

type Config struct {
	commoncfg.BaseConfig `mapstructure:",squash"`
	KMIPServer           KMIPServer `mapstructure:"kmip" yaml:"kmip" json:"kmip"`
}

type KMIPServer struct {
	Address string          `yaml:"address" json:"address" default:":5696"`
	TLS     *commoncfg.MTLS `yaml:"tls" json:"tls"`
}
