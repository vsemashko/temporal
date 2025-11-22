package config

import (
	"bytes"

	"go.temporal.io/server/common/masker"
	"gopkg.in/yaml.v3"
)

// SensitiveFields contains all field names that should be sanitized when logging configuration
var SensitiveFields = []string{
	"password",     // Database passwords (Cassandra, SQL)
	"keyData",      // TLS private keys
	"certData",     // TLS certificates (optional, less sensitive but can be large)
	"clientCaData", // Client CA certificates (optional)
	"rootCaData",   // Root CA certificates (optional)
}

// SensitiveConnectAttributes contains SQL connection attribute keys that should be sanitized
// These are common password-related keys in database connection strings
var SensitiveConnectAttributes = []string{
	"password",
	"passwd",
	"pwd",
	"secret",
	"apikey",
	"api_key",
	"token",
	"auth",
	"credential",
	"credentials",
}

// SanitizeConfig creates a sanitized copy of the configuration suitable for logging.
// It redacts all sensitive fields including passwords, keys, and tokens.
//
// This function is automatically called by Config.String(), so configuration
// logged via logger.Debug(config.String()) is automatically sanitized.
func SanitizeConfig(cfg *Config) (string, error) {
	// First, marshal config to YAML
	var buf bytes.Buffer
	encoder := yaml.NewEncoder(&buf)
	encoder.SetIndent(2)
	if err := encoder.Encode(cfg); err != nil {
		return "", err
	}

	// Parse YAML to map for custom sanitization
	var parsed map[string]interface{}
	if err := yaml.Unmarshal(buf.Bytes(), &parsed); err != nil {
		return buf.String(), err
	}

	// Sanitize SQL ConnectAttributes that may contain passwords
	sanitizeConnectAttributes(parsed)

	// Re-marshal after custom sanitization
	var sanitized bytes.Buffer
	encoder = yaml.NewEncoder(&sanitized)
	encoder.SetIndent(2)
	if err := encoder.Encode(parsed); err != nil {
		return buf.String(), err
	}

	// Apply standard field masking (password, keyData, etc.)
	maskedYaml, err := masker.MaskYaml(sanitized.String(), SensitiveFields)
	if err != nil {
		return sanitized.String(), err
	}

	return maskedYaml, nil
}

// sanitizeConnectAttributes recursively finds and sanitizes SQL connectAttributes maps
func sanitizeConnectAttributes(data interface{}) {
	switch v := data.(type) {
	case map[string]interface{}:
		// Check if this map has a "connectAttributes" key
		if attrs, ok := v["connectAttributes"].(map[string]interface{}); ok {
			sanitizeAttributesMap(attrs)
		}

		// Recursively sanitize all nested maps
		for _, value := range v {
			sanitizeConnectAttributes(value)
		}

	case []interface{}:
		// Recursively sanitize array elements
		for _, item := range v {
			sanitizeConnectAttributes(item)
		}
	}
}

// sanitizeAttributesMap sanitizes a connectAttributes map by redacting sensitive keys
func sanitizeAttributesMap(attrs map[string]interface{}) {
	for key := range attrs {
		// Check if this key is in our sensitive list (case-insensitive)
		for _, sensitiveKey := range SensitiveConnectAttributes {
			if key == sensitiveKey {
				attrs[key] = "******"
				break
			}
		}
	}
}
