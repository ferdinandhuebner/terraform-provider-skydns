package main

import (
	"github.com/coreos/etcd/client"
	"github.com/coreos/etcd/pkg/transport"
	"net/http"
	"time"
)

type Config struct {
	Endpoints []string
	Username  string
	Password  string
	CaFile    string
	CertFile  string
	KeyFile   string
}

func getTlsTransport(config *Config) (*http.Transport, error) {
	caFile := config.CaFile
	certFile := config.CertFile
	keyFile := config.KeyFile

	defaultDialTimeout := 30 * time.Second

	tls := transport.TLSInfo{
		CAFile:   caFile,
		CertFile: certFile,
		KeyFile:  keyFile,
	}
	return transport.NewTransport(tls, defaultDialTimeout)
}

func (c *Config) Client() (*client.Client, error) {
	var tr client.CancelableTransport
	if c.CaFile != "" && c.CertFile != "" && c.KeyFile != "" {
		tlsTransport, err := getTlsTransport(c)
		if err != nil {
			return nil, err
		} else {
			tr = tlsTransport
		}
	} else {
		tr = client.DefaultTransport
	}

	var cfg client.Config
	if c.Username != "" && c.Password != "" {
		cfg = client.Config{
			Endpoints: c.Endpoints,
			Transport: tr,
			Username:  c.Username,
			Password:  c.Password,
		}
	} else {
		cfg = client.Config{
			Endpoints: c.Endpoints,
			Transport: tr,
		}
	}

	etcdClient, err := client.New(cfg)
	if err != nil {
		return nil, err
	} else {
		return &etcdClient, nil
	}
}
