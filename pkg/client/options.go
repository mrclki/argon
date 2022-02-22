package client

import (
	"crypto/tls"
)

type Option func(*Options)

type Options struct {
	TLSConfig *tls.Config
}

// Apply calls each option on o in turn
func (o *Options) Apply(options ...Option) {
	for _, option := range options {
		option(o)
	}
}

func WithTLSConfig(config *tls.Config) Option {
	return func(o *Options) {
		o.TLSConfig = config
	}
}
