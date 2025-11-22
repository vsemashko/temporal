package auth

type (
	// TLS describe TLS configuration (for Cassandra, SQL)
	TLS struct {
		Enabled bool `yaml:"enabled"`

		// CertPath and KeyPath are optional depending on server
		// config, but both fields must be omitted to avoid using a
		// client certificate
		CertFile string `yaml:"certFile"`
		KeyFile  string `yaml:"keyFile"`
		CaFile   string `yaml:"caFile"` // optional depending on server config

		// EnableHostVerification controls TLS hostname verification
		// SECURITY WARNING: Disabling host verification (setting to false) exposes connections to
		// man-in-the-middle attacks and should ONLY be done in development/testing environments.
		// When false, this sets InsecureSkipVerify=true which disables certificate hostname validation.
		// See http://golang.org/pkg/crypto/tls/ InSecureSkipVerify for more info.
		// DEFAULT: false (host verification disabled) - STRONGLY recommended to set to true in production
		EnableHostVerification bool `yaml:"enableHostVerification"`

		ServerName string `yaml:"serverName"`

		// Base64 equivalents of the above artifacts.
		// You cannot specify both a Data and a File for the same artifact (e.g. setting CertFile and CertData)
		CertData string `yaml:"certData"`
		KeyData  string `yaml:"keyData"`
		CaData   string `yaml:"caData"` // optional depending on server config
	}
)
