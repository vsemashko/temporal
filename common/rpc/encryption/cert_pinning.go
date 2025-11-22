package encryption

import (
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"encoding/hex"
	"fmt"
	"strings"

	"go.temporal.io/server/common/log"
	"go.temporal.io/server/common/log/tag"
	"go.temporal.io/server/common/metrics"
)

// CertificatePinner validates TLS certificates against pinned fingerprints
type CertificatePinner struct {
	// pins maps cluster names to their pinned certificate fingerprints
	pins map[string][]string
	// strictMode determines behavior when no pins match
	strictMode bool
	logger     log.Logger
	metrics    metrics.Handler
}

// PinConfig represents certificate pinning configuration for a single cluster
type PinConfig struct {
	// Fingerprints is a list of SHA-256 fingerprints of pinned certificates
	// Format: "sha256:abc123..." or just "abc123..."
	Fingerprints []string
	// Description provides context about the pinned certificate
	Description string
	// StrictPinning controls whether to fail if no pins match
	// If false, pinning acts as additional validation but doesn't block connections
	StrictPinning bool
}

// NewCertificatePinner creates a new certificate pinner
func NewCertificatePinner(
	pins map[string]PinConfig,
	logger log.Logger,
	metricsHandler metrics.Handler,
) *CertificatePinner {
	// Normalize all fingerprints (remove "sha256:" prefix, convert to lowercase)
	normalizedPins := make(map[string][]string)
	for cluster, config := range pins {
		normalized := make([]string, 0, len(config.Fingerprints))
		for _, fp := range config.Fingerprints {
			normalized = append(normalized, normalizeFingerprint(fp))
		}
		normalizedPins[cluster] = normalized
	}

	return &CertificatePinner{
		pins:       normalizedPins,
		strictMode: false, // Can be overridden per-cluster
		logger:     logger,
		metrics:    metricsHandler,
	}
}

// ValidatePinnedCertificate validates a certificate against pinned fingerprints
// Returns nil if validation passes, error otherwise
func (p *CertificatePinner) ValidatePinnedCertificate(
	clusterName string,
	cert *x509.Certificate,
	strictMode bool,
) error {
	// Get pinned fingerprints for this cluster
	pinnedFingerprints, hasPins := p.pins[clusterName]

	// If no pins configured for this cluster, allow connection
	if !hasPins || len(pinnedFingerprints) == 0 {
		p.logger.Debug("No certificate pins configured for cluster",
			tag.NewStringTag("cluster", clusterName))
		return nil
	}

	// Calculate fingerprint of the presented certificate
	certFingerprint := calculateFingerprint(cert)

	// Check if certificate fingerprint matches any pinned fingerprint
	for _, pinnedFP := range pinnedFingerprints {
		if certFingerprint == pinnedFP {
			p.logger.Debug("Certificate pin validation successful",
				tag.NewStringTag("cluster", clusterName),
				tag.NewStringTag("fingerprint", certFingerprint[:16]+"..."))

			p.metrics.Counter(metrics.CertPinValidationSuccess.Name()).Record(1,
				metrics.StringTag("cluster", clusterName))

			return nil
		}
	}

	// No matching pin found
	p.logger.Warn("Certificate pin validation failed - fingerprint mismatch",
		tag.NewStringTag("cluster", clusterName),
		tag.NewStringTag("cert_fingerprint", certFingerprint[:16]+"..."),
		tag.NewStringTag("subject", cert.Subject.String()))

	p.metrics.Counter(metrics.CertPinValidationFailure.Name()).Record(1,
		metrics.StringTag("cluster", clusterName))

	if strictMode {
		return fmt.Errorf(
			"certificate pin validation failed for cluster %s: certificate fingerprint %s does not match any pinned fingerprints",
			clusterName,
			certFingerprint[:16]+"...",
		)
	}

	// In non-strict mode, log warning but allow connection
	p.logger.Warn("Certificate pin mismatch in non-strict mode - allowing connection",
		tag.NewStringTag("cluster", clusterName))

	return nil
}

// CreateVerifyPeerCertificate creates a VerifyPeerCertificate function for tls.Config
// This function is called during TLS handshake to validate the peer's certificate
func (p *CertificatePinner) CreateVerifyPeerCertificate(
	clusterName string,
	strictMode bool,
) func(rawCerts [][]byte, verifiedChains [][]*x509.Certificate) error {
	return func(rawCerts [][]byte, verifiedChains [][]*x509.Certificate) error {
		// If no verified chains, let standard verification fail
		if len(verifiedChains) == 0 {
			return fmt.Errorf("no verified certificate chains")
		}

		// Validate the leaf certificate (first cert in the first verified chain)
		if len(verifiedChains[0]) == 0 {
			return fmt.Errorf("empty certificate chain")
		}

		leafCert := verifiedChains[0][0]
		return p.ValidatePinnedCertificate(clusterName, leafCert, strictMode)
	}
}

// GetPinnedClustersCount returns the number of clusters with pinning configured
func (p *CertificatePinner) GetPinnedClustersCount() int {
	return len(p.pins)
}

// GetPinnedFingerprintsForCluster returns the pinned fingerprints for a cluster
func (p *CertificatePinner) GetPinnedFingerprintsForCluster(clusterName string) []string {
	if fps, ok := p.pins[clusterName]; ok {
		// Return a copy to prevent modification
		result := make([]string, len(fps))
		copy(result, fps)
		return result
	}
	return nil
}

// Helper functions

// calculateFingerprint computes SHA-256 fingerprint of a certificate
func calculateFingerprint(cert *x509.Certificate) string {
	hash := sha256.Sum256(cert.Raw)
	return hex.EncodeToString(hash[:])
}

// normalizeFingerprint removes "sha256:" prefix and converts to lowercase
func normalizeFingerprint(fp string) string {
	// Remove "sha256:" prefix if present
	normalized := strings.TrimPrefix(fp, "sha256:")
	normalized = strings.TrimPrefix(normalized, "SHA256:")

	// Remove any colons or spaces
	normalized = strings.ReplaceAll(normalized, ":", "")
	normalized = strings.ReplaceAll(normalized, " ", "")

	// Convert to lowercase
	return strings.ToLower(normalized)
}

// FormatFingerprint formats a fingerprint in standard format (SHA256:AB:CD:EF:...)
func FormatFingerprint(fp string) string {
	normalized := normalizeFingerprint(fp)

	// Insert colons every 2 characters
	var formatted strings.Builder
	formatted.WriteString("SHA256:")

	for i := 0; i < len(normalized); i += 2 {
		if i > 0 {
			formatted.WriteString(":")
		}
		if i+2 <= len(normalized) {
			formatted.WriteString(strings.ToUpper(normalized[i : i+2]))
		} else {
			formatted.WriteString(strings.ToUpper(normalized[i:]))
		}
	}

	return formatted.String()
}

// GetCertificateFingerprint is a utility function to get the fingerprint of a certificate
// Useful for operators to obtain fingerprints for configuration
func GetCertificateFingerprint(cert *x509.Certificate) string {
	return FormatFingerprint(calculateFingerprint(cert))
}

// ValidateFingerprint checks if a fingerprint string is valid
func ValidateFingerprint(fp string) error {
	normalized := normalizeFingerprint(fp)

	// SHA-256 produces 32 bytes = 64 hex characters
	if len(normalized) != 64 {
		return fmt.Errorf("invalid fingerprint length: expected 64 hex characters, got %d", len(normalized))
	}

	// Check if all characters are valid hex
	for _, ch := range normalized {
		if !((ch >= '0' && ch <= '9') || (ch >= 'a' && ch <= 'f')) {
			return fmt.Errorf("invalid fingerprint: contains non-hex character '%c'", ch)
		}
	}

	return nil
}
