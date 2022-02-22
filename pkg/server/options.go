package server

import (
	"crypto/tls"
)

type Option func(*Options)

type Options struct {
	Id             string
	Addr           string
	Port           int
	TLSConfig      *tls.Config
	StoragePath    string
	PrometheusAddr string
	PrometheusPort int
}

// Apply calls each option on o in turn
func (o *Options) Apply(options ...Option) {
	for _, option := range options {
		option(o)
	}
}

func WithId(id string) Option {
	return func(o *Options) {
		o.Id = id
	}
}

func WithAddr(addr string) Option {
	return func(o *Options) {
		o.Addr = addr
	}
}

func WithPort(port int) Option {
	return func(o *Options) {
		o.Port = port
	}
}

func WithTLSConfig(config *tls.Config) Option {
	return func(o *Options) {
		o.TLSConfig = config
	}
}

func WithStoragePath(path string) Option {
	return func(o *Options) {
		o.StoragePath = path
	}
}

func WithPrometheusAddr(addr string) Option {
	return func(o *Options) {
		o.PrometheusAddr = addr
	}
}

func WithPrometheusPort(port int) Option {
	return func(o *Options) {
		o.PrometheusPort = port
	}
}
