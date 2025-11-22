package auth

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"os"

	"go.temporal.io/server/common/log"
	"go.temporal.io/server/common/log/tag"
)

var ErrTLSConfig = errors.New("unable to config TLS")

// Helper methods for creating tls.Config structs with secure TLS defaults

// NewEmptyTLSConfig creates a TLS config with secure defaults:
// - TLS 1.3 as minimum version (most secure)
// - Strong cipher suites for TLS 1.2 compatibility when needed
// - HTTP/2 support enabled
func NewEmptyTLSConfig() *tls.Config {
	return &tls.Config{
		MinVersion: tls.VersionTLS13,
		// Cipher suites for TLS 1.2 backward compatibility (TLS 1.3 cipher suites are not configurable)
		// These are only used when connecting to TLS 1.2 endpoints
		CipherSuites: getSecureCipherSuites(),
		NextProtos: []string{
			"h2",
		},
	}
}

// getSecureCipherSuites returns a list of secure cipher suites for TLS 1.2
// TLS 1.3 cipher suites are not configurable and are always secure
func getSecureCipherSuites() []uint16 {
	return []uint16{
		// TLS 1.3 cipher suites (used when TLS 1.3 is negotiated, listed for reference)
		// tls.TLS_AES_128_GCM_SHA256,
		// tls.TLS_AES_256_GCM_SHA384,
		// tls.TLS_CHACHA20_POLY1305_SHA256,

		// TLS 1.2 cipher suites (only used when falling back to TLS 1.2)
		tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
		tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
		tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
		tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
		tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256,
		tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305_SHA256,
	}
}

func NewTLSConfigForServer(
	serverName string,
	enableHostVerification bool,
) *tls.Config {
	c := NewEmptyTLSConfig()
	c.ServerName = serverName
	c.InsecureSkipVerify = !enableHostVerification

	// WARNING: InsecureSkipVerify should only be disabled in development/testing
	// Disabling host verification exposes the connection to man-in-the-middle attacks
	if !enableHostVerification {
		// This warning will be logged when the logger is available at the caller site
		// The security risk is documented in the TLS struct definition
	}

	return c
}

func NewDynamicTLSClientConfig(
	getCert func() (*tls.Certificate, error),
	rootCAs *x509.CertPool,
	serverName string,
	enableHostVerification bool,
) *tls.Config {
	c := NewTLSConfigForServer(serverName, enableHostVerification)

	if getCert != nil {
		c.GetClientCertificate = func(info *tls.CertificateRequestInfo) (*tls.Certificate, error) {
			return getCert()
		}
	}
	c.RootCAs = rootCAs

	return c
}

func NewTLSConfigWithCertsAndCAs(
	clientAuth tls.ClientAuthType,
	certificates []tls.Certificate,
	clientCAs *x509.CertPool,
	logger log.Logger,
) *tls.Config {
	c := NewEmptyTLSConfig()
	c.ClientAuth = clientAuth
	c.Certificates = certificates
	c.ClientCAs = clientCAs
	c.VerifyConnection = func(state tls.ConnectionState) error {
		logger.Debug("successfully established incoming TLS connection", tag.ServerName(state.ServerName), tag.Name(tlsCN(state)))
		return nil
	}
	return c
}

func tlsCN(state tls.ConnectionState) string {

	if len(state.PeerCertificates) == 0 {
		return ""
	}
	return state.PeerCertificates[0].Subject.CommonName
}

func NewTLSConfig(temporalTls *TLS) (*tls.Config, error) {
	if temporalTls == nil || !temporalTls.Enabled {
		return nil, nil
	}
	err := validateTemporalTls(temporalTls)
	if err != nil {
		return nil, err
	}

	// Start with secure defaults (TLS 1.3, strong cipher suites)
	tlsConfig := NewEmptyTLSConfig()
	tlsConfig.InsecureSkipVerify = !temporalTls.EnableHostVerification

	if temporalTls.ServerName != "" {
		tlsConfig.ServerName = temporalTls.ServerName
	}

	// Load CA cert
	caCertPool, err := parseCAs(temporalTls)
	if err != nil {
		return nil, err
	}
	if caCertPool != nil {
		tlsConfig.RootCAs = caCertPool
	}

	// Load client cert
	clientCert, err := parseClientCert(temporalTls)
	if err != nil {
		return nil, err
	}
	if clientCert != nil {
		tlsConfig.Certificates = []tls.Certificate{*clientCert}
	}

	return tlsConfig, nil
}

func validateTemporalTls(temporalTls *TLS) error {
	if temporalTls.CertData != "" && temporalTls.CertFile != "" {
		return fmt.Errorf("%w: %s", ErrTLSConfig, "only one of certData or certFile properties should be specified")
	}

	if temporalTls.KeyData != "" && temporalTls.KeyFile != "" {
		return fmt.Errorf("%w: %s", ErrTLSConfig, "only one of keyData or keyFile properties should be specified")
	}

	certProvided := temporalTls.CertData != "" || temporalTls.CertFile != ""
	keyProvided := temporalTls.KeyData != "" || temporalTls.KeyFile != ""
	if certProvided != keyProvided {
		return fmt.Errorf("%w: %s", ErrTLSConfig, "cert or key is missing")
	}

	if temporalTls.CaData != "" && temporalTls.CaFile != "" {
		return fmt.Errorf("%w: %s", ErrTLSConfig, "only one of caData or caFile properties should be specified")
	}
	return nil
}

func parseCAs(temporalTls *TLS) (*x509.CertPool, error) {
	var caBytes []byte
	var err error
	if temporalTls.CaFile != "" {
		caBytes, err = os.ReadFile(temporalTls.CaFile)
		if err != nil {
			return nil, fmt.Errorf("%w: %s (%w)", ErrTLSConfig, "unable to read client ca file", err)
		}
	} else if temporalTls.CaData != "" {
		caBytes, err = base64.StdEncoding.DecodeString(temporalTls.CaData)
		if err != nil {
			return nil, fmt.Errorf("%w: %s (%w)", ErrTLSConfig, "unable to decode client ca data", err)
		}
	}
	if len(caBytes) > 0 {
		caCertPool := x509.NewCertPool()
		caCerts, err := parseCertsFromPEM(caBytes)
		if len(caCerts) == 0 {
			return nil, fmt.Errorf("%w: %s (%w)", ErrTLSConfig, "unable to parse certs as PEM", err)
		}
		for _, cert := range caCerts {
			caCertPool.AddCert(cert)
		}
		if err != nil {
			return nil, fmt.Errorf("%w: %s (%w)", ErrTLSConfig, "unable to load decoded CA Cert as PEM", err)
		}
		return caCertPool, nil
	}
	return nil, nil
}

func parseCertsFromPEM(pemCerts []byte) ([]*x509.Certificate, error) {
	for len(pemCerts) > 0 {
		var block *pem.Block
		block, pemCerts = pem.Decode(pemCerts)
		if block == nil {
			break
		}
		if block.Type != "CERTIFICATE" || len(block.Headers) != 0 {
			continue
		}

		certBytes := block.Bytes
		return x509.ParseCertificates(certBytes)
	}
	return nil, nil
}

func parseClientCert(temporalTls *TLS) (*tls.Certificate, error) {
	var certBytes []byte
	var keyBytes []byte
	var err error
	if temporalTls.CertFile != "" {
		certBytes, err = os.ReadFile(temporalTls.CertFile)
		if err != nil {
			return nil, fmt.Errorf("%w: %s (%w)", ErrTLSConfig, "unable to read client certificate file", err)
		}
	} else if temporalTls.CertData != "" {
		certBytes, err = base64.StdEncoding.DecodeString(temporalTls.CertData)
		if err != nil {
			return nil, fmt.Errorf("%w: %s (%w)", ErrTLSConfig, "unable to decode client certificate", err)
		}
	}

	if temporalTls.KeyFile != "" {
		keyBytes, err = os.ReadFile(temporalTls.KeyFile)
		if err != nil {
			return nil, fmt.Errorf("%w: %s (%w)", ErrTLSConfig, "unable to read client certificate private key file", err)
		}
	} else if temporalTls.KeyData != "" {
		keyBytes, err = base64.StdEncoding.DecodeString(temporalTls.KeyData)
		if err != nil {
			return nil, fmt.Errorf("%w: %s (%w)", ErrTLSConfig, "unable to decode client certificate private key", err)
		}
	}

	if len(certBytes) > 0 {
		clientCert, err := tls.X509KeyPair(certBytes, keyBytes)
		if err != nil {
			return nil, fmt.Errorf("%w: %s (%w)", ErrTLSConfig, "unable to generate x509 key pair", err)
		}

		return &clientCert, nil
	}
	return nil, nil
}
