package config

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSanitizeConfig_DatabasePasswords(t *testing.T) {
	cfg := &Config{
		Persistence: Persistence{
			DefaultStore: "test",
			DataStores: map[string]DataStore{
				"cassandra-store": {
					Cassandra: &Cassandra{
						Hosts:    "localhost",
						Port:     9042,
						User:     "cassandra_user",
						Password: "super_secret_cassandra_password",
						Keyspace: "temporal",
					},
				},
				"sql-store": {
					SQL: &SQL{
						User:         "sql_user",
						Password:     "super_secret_sql_password",
						PluginName:   "postgres",
						DatabaseName: "temporal",
						ConnectAddr:  "localhost:5432",
						ConnectProtocol: "tcp",
					},
				},
			},
		},
	}

	sanitized, err := SanitizeConfig(cfg)
	require.NoError(t, err)

	// Verify passwords are masked
	assert.NotContains(t, sanitized, "super_secret_cassandra_password")
	assert.NotContains(t, sanitized, "super_secret_sql_password")
	assert.Contains(t, sanitized, "******")

	// Verify usernames are NOT masked
	assert.Contains(t, sanitized, "cassandra_user")
	assert.Contains(t, sanitized, "sql_user")
}

func TestSanitizeConfig_TLSPrivateKeys(t *testing.T) {
	cfg := &Config{
		Global: Global{
			TLS: RootTLS{
				Frontend: GroupTLS{
					Server: ServerTLS{
						CertData: "-----BEGIN CERTIFICATE-----\nMIICert...\n-----END CERTIFICATE-----",
						KeyData:  "-----BEGIN PRIVATE KEY-----\nMIIPrivateKey...\n-----END PRIVATE KEY-----",
					},
				},
				SystemWorker: WorkerTLS{
					CertData: "-----BEGIN CERTIFICATE-----\nWorkerCert...\n-----END CERTIFICATE-----",
					KeyData:  "-----BEGIN PRIVATE KEY-----\nWorkerKey...\n-----END PRIVATE KEY-----",
				},
			},
		},
	}

	sanitized, err := SanitizeConfig(cfg)
	require.NoError(t, err)

	// Verify private keys are masked
	assert.NotContains(t, sanitized, "MIIPrivateKey")
	assert.NotContains(t, sanitized, "WorkerKey")
	assert.Contains(t, sanitized, "******")

	// Verify certificates are also masked (optional, but good practice)
	assert.NotContains(t, sanitized, "MIICert")
	assert.NotContains(t, sanitized, "WorkerCert")
}

func TestSanitizeConfig_SQLConnectAttributes(t *testing.T) {
	cfg := &Config{
		Persistence: Persistence{
			DefaultStore: "sql-main",
			DataStores: map[string]DataStore{
				"sql-main": {
					SQL: &SQL{
						User:            "admin",
						Password:        "db_password",
						PluginName:      "postgres",
						DatabaseName:    "temporal",
						ConnectAddr:     "localhost:5432",
						ConnectProtocol: "tcp",
						ConnectAttributes: map[string]string{
							"sslmode":  "require",
							"password": "connection_string_password",
							"apikey":   "secret_api_key",
							"timeout":  "30s",
						},
					},
				},
			},
		},
	}

	sanitized, err := SanitizeConfig(cfg)
	require.NoError(t, err)

	// Verify main password is masked
	assert.NotContains(t, sanitized, "db_password")

	// Verify ConnectAttributes passwords are masked
	assert.NotContains(t, sanitized, "connection_string_password")
	assert.NotContains(t, sanitized, "secret_api_key")

	// Verify non-sensitive attributes are preserved
	assert.Contains(t, sanitized, "sslmode")
	assert.Contains(t, sanitized, "require")
	assert.Contains(t, sanitized, "timeout")
	assert.Contains(t, sanitized, "30s")
}

func TestSanitizeConfig_MultipleDataStores(t *testing.T) {
	cfg := &Config{
		Persistence: Persistence{
			DefaultStore: "default",
			DataStores: map[string]DataStore{
				"default": {
					SQL: &SQL{
						User:     "user1",
						Password: "password1",
						PluginName: "postgres",
						DatabaseName: "db1",
						ConnectAddr: "host1:5432",
						ConnectProtocol: "tcp",
						ConnectAttributes: map[string]string{
							"password": "attr_password1",
						},
					},
				},
				"visibility": {
					SQL: &SQL{
						User:     "user2",
						Password: "password2",
						PluginName: "postgres",
						DatabaseName: "db2",
						ConnectAddr: "host2:5432",
						ConnectProtocol: "tcp",
						ConnectAttributes: map[string]string{
							"password": "attr_password2",
							"token":    "secret_token",
						},
					},
				},
			},
		},
	}

	sanitized, err := SanitizeConfig(cfg)
	require.NoError(t, err)

	// Verify all passwords are masked
	assert.NotContains(t, sanitized, "password1")
	assert.NotContains(t, sanitized, "password2")
	assert.NotContains(t, sanitized, "attr_password1")
	assert.NotContains(t, sanitized, "attr_password2")
	assert.NotContains(t, sanitized, "secret_token")

	// Verify usernames and hosts are preserved
	assert.Contains(t, sanitized, "user1")
	assert.Contains(t, sanitized, "user2")
	assert.Contains(t, sanitized, "host1:5432")
	assert.Contains(t, sanitized, "host2:5432")
}

func TestSanitizeConfig_EmptyConfig(t *testing.T) {
	cfg := &Config{}

	sanitized, err := SanitizeConfig(cfg)
	require.NoError(t, err)
	assert.NotEmpty(t, sanitized)
}

func TestSanitizeConfig_NoSensitiveData(t *testing.T) {
	cfg := &Config{
		Persistence: Persistence{
			DefaultStore:     "default",
			NumHistoryShards: 4,
		},
	}

	sanitized, err := SanitizeConfig(cfg)
	require.NoError(t, err)
	assert.Contains(t, sanitized, "default")
	assert.Contains(t, sanitized, "4")
}

func TestSanitizeConnectAttributes_VariousKeys(t *testing.T) {
	// Test all sensitive keys
	testCases := []struct {
		key       string
		value     string
		shouldMask bool
	}{
		{"password", "secret", true},
		{"passwd", "secret", true},
		{"pwd", "secret", true},
		{"secret", "secret", true},
		{"apikey", "secret", true},
		{"api_key", "secret", true},
		{"token", "secret", true},
		{"auth", "secret", true},
		{"credential", "secret", true},
		{"credentials", "secret", true},
		{"sslmode", "require", false},
		{"timeout", "30s", false},
		{"maxConnections", "100", false},
	}

	for _, tc := range testCases {
		t.Run(tc.key, func(t *testing.T) {
			cfg := &Config{
				Persistence: Persistence{
					DefaultStore: "test",
					DataStores: map[string]DataStore{
						"test": {
							SQL: &SQL{
								User:            "user",
								PluginName:      "postgres",
								DatabaseName:    "db",
								ConnectAddr:     "localhost",
								ConnectProtocol: "tcp",
								ConnectAttributes: map[string]string{
									tc.key: tc.value,
								},
							},
						},
					},
				},
			}

			sanitized, err := SanitizeConfig(cfg)
			require.NoError(t, err)

			if tc.shouldMask {
				assert.NotContains(t, sanitized, tc.value, "Expected %s to be masked", tc.key)
			} else {
				assert.Contains(t, sanitized, tc.value, "Expected %s to be preserved", tc.key)
			}
		})
	}
}

func TestConfigString_UsesSanitization(t *testing.T) {
	cfg := &Config{
		Persistence: Persistence{
			DefaultStore: "test",
			DataStores: map[string]DataStore{
				"test": {
					SQL: &SQL{
						User:            "admin",
						Password:        "super_secret_password",
						PluginName:      "postgres",
						DatabaseName:    "temporal",
						ConnectAddr:     "localhost:5432",
						ConnectProtocol: "tcp",
					},
				},
			},
		},
		Global: Global{
			TLS: RootTLS{
				Frontend: GroupTLS{
					Server: ServerTLS{
						KeyData: "-----BEGIN PRIVATE KEY-----\nSecret\n-----END PRIVATE KEY-----",
					},
				},
			},
		},
	}

	configString := cfg.String()

	// Verify sensitive data is not in the string representation
	assert.NotContains(t, configString, "super_secret_password")
	assert.NotContains(t, configString, "Secret")
	assert.Contains(t, configString, "******")

	// Verify non-sensitive data is preserved
	assert.Contains(t, configString, "admin")
	assert.Contains(t, configString, "temporal")
	assert.Contains(t, configString, "localhost:5432")
}

func TestConfigString_FallbackOnError(t *testing.T) {
	// Test that Config.String() doesn't panic even with unusual configurations
	cfg := &Config{
		Persistence: Persistence{
			DefaultStore: "test",
		},
	}

	configString := cfg.String()
	assert.NotEmpty(t, configString)
	assert.True(t, strings.Contains(configString, "test") || strings.Contains(configString, "persistence"))
}
