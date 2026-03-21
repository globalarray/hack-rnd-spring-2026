package secure

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"os"
)

var ErrCannotAppendRootCA = errors.New("cannot append root CA")

type TLSMode uint8

const (
	ClientMode TLSMode = iota
	ServerMode
)

func loadTLSBase(certFile, keyFile string) (*tls.Config, error) {
	const op string = "security.loadTLSBase"

	cert, err := tls.LoadX509KeyPair(certFile, keyFile)

	if err != nil {
		return nil, fmt.Errorf("%s: %v", op, err)
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:   tls.VersionTLS12,
	}

	return tlsConfig, nil
}

func loadCAPool(caFile string) (*x509.CertPool, error) {
	const op string = "security.loadCAPool"

	caCert, err := os.ReadFile(caFile)

	if err != nil {
		return nil, fmt.Errorf("%s: %v", op, err)
	}

	caCertPool := x509.NewCertPool()

	if !caCertPool.AppendCertsFromPEM(caCert) {
		return nil, fmt.Errorf("%s: %w", op, ErrCannotAppendRootCA)
	}

	return caCertPool, nil
}

func LoadTLSConfig(certFile, keyFile, caFile string, mode TLSMode) (*tls.Config, error) {
	const op string = "security.LoadTLSConfig"

	tlsCfg, err := loadTLSBase(certFile, keyFile)

	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	caPool, err := loadCAPool(caFile)

	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	switch mode {
	case ServerMode:
		tlsCfg.ClientAuth = tls.RequireAndVerifyClientCert
		tlsCfg.ClientCAs = caPool
	case ClientMode:
		tlsCfg.RootCAs = caPool
	}

	return tlsCfg, nil
}
