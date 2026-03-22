package shared

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

type Config struct {
	Address        string
	Insecure       bool
	CACertPath     string
	ClientCertPath string
	ClientKeyPath  string
	ServerName     string
}

func Dial(ctx context.Context, cfg Config) (*grpc.ClientConn, error) {
	if strings.TrimSpace(cfg.Address) == "" {
		return nil, fmt.Errorf("grpc address is required")
	}

	var dialOptions []grpc.DialOption
	if cfg.Insecure {
		dialOptions = append(dialOptions, grpc.WithTransportCredentials(insecure.NewCredentials()))
	} else {
		tlsConfig, err := loadTLSConfig(cfg)
		if err != nil {
			return nil, err
		}
		dialOptions = append(dialOptions, grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)))
	}

	dialOptions = append(dialOptions, grpc.WithBlock())

	conn, err := grpc.DialContext(ctx, cfg.Address, dialOptions...)
	if err != nil {
		return nil, fmt.Errorf("grpc dial %s: %w", cfg.Address, err)
	}

	return conn, nil
}

func loadTLSConfig(cfg Config) (*tls.Config, error) {
	if cfg.CACertPath == "" || cfg.ClientCertPath == "" || cfg.ClientKeyPath == "" {
		return nil, fmt.Errorf("mTLS requires CA cert, client cert and client key")
	}

	caCert, err := os.ReadFile(cfg.CACertPath)
	if err != nil {
		return nil, fmt.Errorf("read CA cert: %w", err)
	}

	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(caCert) {
		return nil, fmt.Errorf("append CA cert: invalid PEM")
	}

	clientCert, err := tls.LoadX509KeyPair(cfg.ClientCertPath, cfg.ClientKeyPath)
	if err != nil {
		return nil, fmt.Errorf("load client certificate: %w", err)
	}

	return &tls.Config{
		MinVersion:   tls.VersionTLS12,
		RootCAs:      certPool,
		Certificates: []tls.Certificate{clientCert},
		ServerName:   cfg.ServerName,
	}, nil
}
