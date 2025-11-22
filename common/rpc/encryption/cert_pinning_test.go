package encryption

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.temporal.io/server/common/log"
	"go.temporal.io/server/common/metrics"
)

type (
	certPinningTestSuite struct {
		suite.Suite
		*require.Assertions

		logger         log.Logger
		metricsHandler metrics.Handler
		testCert1      *x509.Certificate
		testCert2      *x509.Certificate
		fingerprint1   string
		fingerprint2   string
	}
)

func TestCertPinningTestSuite(t *testing.T) {
	suite.Run(t, &certPinningTestSuite{})
}

func (s *certPinningTestSuite) SetupTest() {
	s.Assertions = require.New(s.T())
	s.logger = log.NewTestLogger()
	s.metricsHandler = metrics.NoopMetricsHandler

	// Generate test certificates
	s.testCert1 = s.generateTestCertificate("test1.example.com")
	s.testCert2 = s.generateTestCertificate("test2.example.com")

	// Calculate fingerprints
	s.fingerprint1 = calculateFingerprint(s.testCert1)
	s.fingerprint2 = calculateFingerprint(s.testCert2)
}

func (s *certPinningTestSuite) TestNoPinsConfigured() {
	pinner := NewCertificatePinner(nil, s.logger, s.metricsHandler)

	// Should allow any certificate when no pins configured
	err := pinner.ValidatePinnedCertificate("cluster1", s.testCert1, true)
	s.NoError(err)
}

func (s *certPinningTestSuite) TestValidPinMatch() {
	pins := map[string]PinConfig{
		"cluster1": {
			Fingerprints:  []string{s.fingerprint1},
			StrictPinning: true,
		},
	}

	pinner := NewCertificatePinner(pins, s.logger, s.metricsHandler)

	// Should succeed with matching fingerprint
	err := pinner.ValidatePinnedCertificate("cluster1", s.testCert1, true)
	s.NoError(err)
}

func (s *certPinningTestSuite) TestPinMismatchStrict() {
	pins := map[string]PinConfig{
		"cluster1": {
			Fingerprints:  []string{s.fingerprint1},
			StrictPinning: true,
		},
	}

	pinner := NewCertificatePinner(pins, s.logger, s.metricsHandler)

	// Should fail with non-matching fingerprint in strict mode
	err := pinner.ValidatePinnedCertificate("cluster1", s.testCert2, true)
	s.Error(err)
	s.Contains(err.Error(), "certificate pin validation failed")
}

func (s *certPinningTestSuite) TestPinMismatchNonStrict() {
	pins := map[string]PinConfig{
		"cluster1": {
			Fingerprints:  []string{s.fingerprint1},
			StrictPinning: false,
		},
	}

	pinner := NewCertificatePinner(pins, s.logger, s.metricsHandler)

	// Should allow with non-matching fingerprint in non-strict mode
	err := pinner.ValidatePinnedCertificate("cluster1", s.testCert2, false)
	s.NoError(err)
}

func (s *certPinningTestSuite) TestMultiplePins() {
	pins := map[string]PinConfig{
		"cluster1": {
			Fingerprints:  []string{s.fingerprint1, s.fingerprint2},
			StrictPinning: true,
		},
	}

	pinner := NewCertificatePinner(pins, s.logger, s.metricsHandler)

	// Should succeed with first fingerprint
	err := pinner.ValidatePinnedCertificate("cluster1", s.testCert1, true)
	s.NoError(err)

	// Should succeed with second fingerprint
	err = pinner.ValidatePinnedCertificate("cluster1", s.testCert2, true)
	s.NoError(err)
}

func (s *certPinningTestSuite) TestFingerprintNormalization() {
	pins := map[string]PinConfig{
		"cluster1": {
			// Test various fingerprint formats
			Fingerprints: []string{
				"sha256:" + s.fingerprint1,                       // with prefix
				s.fingerprint2,                                   // without prefix
				FormatFingerprint(s.fingerprint1),                // with colons
				normalizeFingerprint("SHA256:" + s.fingerprint2), // uppercase prefix
			},
			StrictPinning: true,
		},
	}

	pinner := NewCertificatePinner(pins, s.logger, s.metricsHandler)

	// All formats should be normalized and work
	err := pinner.ValidatePinnedCertificate("cluster1", s.testCert1, true)
	s.NoError(err)

	err = pinner.ValidatePinnedCertificate("cluster1", s.testCert2, true)
	s.NoError(err)
}

func (s *certPinningTestSuite) TestGetPinnedClustersCount() {
	pins := map[string]PinConfig{
		"cluster1": {Fingerprints: []string{s.fingerprint1}},
		"cluster2": {Fingerprints: []string{s.fingerprint2}},
	}

	pinner := NewCertificatePinner(pins, s.logger, s.metricsHandler)
	s.Equal(2, pinner.GetPinnedClustersCount())
}

func (s *certPinningTestSuite) TestGetPinnedFingerprintsForCluster() {
	pins := map[string]PinConfig{
		"cluster1": {Fingerprints: []string{s.fingerprint1, s.fingerprint2}},
	}

	pinner := NewCertificatePinner(pins, s.logger, s.metricsHandler)

	fps := pinner.GetPinnedFingerprintsForCluster("cluster1")
	s.Len(fps, 2)
	s.Contains(fps, s.fingerprint1)
	s.Contains(fps, s.fingerprint2)

	// Non-existent cluster
	fps = pinner.GetPinnedFingerprintsForCluster("cluster-nonexistent")
	s.Nil(fps)
}

func (s *certPinningTestSuite) TestVerifyPeerCertificateCallback() {
	pins := map[string]PinConfig{
		"cluster1": {
			Fingerprints:  []string{s.fingerprint1},
			StrictPinning: true,
		},
	}

	pinner := NewCertificatePinner(pins, s.logger, s.metricsHandler)

	// Create VerifyPeerCertificate callback
	verifyFunc := pinner.CreateVerifyPeerCertificate("cluster1", true)

	// Create verified chains (simulating successful TLS verification)
	verifiedChains := [][]*x509.Certificate{
		{s.testCert1}, // Leaf certificate is first in chain
	}

	// Should succeed with matching certificate
	err := verifyFunc(nil, verifiedChains)
	s.NoError(err)

	// Should fail with non-matching certificate
	verifiedChains[0][0] = s.testCert2
	err = verifyFunc(nil, verifiedChains)
	s.Error(err)
}

func (s *certPinningTestSuite) TestVerifyPeerCertificateNoChains() {
	pins := map[string]PinConfig{
		"cluster1": {Fingerprints: []string{s.fingerprint1}},
	}

	pinner := NewCertificatePinner(pins, s.logger, s.metricsHandler)
	verifyFunc := pinner.CreateVerifyPeerCertificate("cluster1", true)

	// Should fail with no verified chains
	err := verifyFunc(nil, nil)
	s.Error(err)
	s.Contains(err.Error(), "no verified certificate chains")
}

func (s *certPinningTestSuite) TestValidateFingerprint() {
	// Valid fingerprint (64 hex characters)
	err := ValidateFingerprint(s.fingerprint1)
	s.NoError(err)

	// Valid with prefix
	err = ValidateFingerprint("sha256:" + s.fingerprint1)
	s.NoError(err)

	// Invalid - too short
	err = ValidateFingerprint("abc123")
	s.Error(err)
	s.Contains(err.Error(), "invalid fingerprint length")

	// Invalid - non-hex characters
	err = ValidateFingerprint("zzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzz")
	s.Error(err)
	s.Contains(err.Error(), "non-hex character")
}

func (s *certPinningTestSuite) TestFormatFingerprint() {
	formatted := FormatFingerprint(s.fingerprint1)

	// Should start with SHA256:
	s.True(len(formatted) > 7)
	s.Equal("SHA256:", formatted[:7])

	// Should contain colons
	s.Contains(formatted, ":")

	// Should be uppercase hex
	for i := 7; i < len(formatted); i++ {
		ch := formatted[i]
		if ch != ':' {
			s.True((ch >= '0' && ch <= '9') || (ch >= 'A' && ch <= 'F'))
		}
	}
}

func (s *certPinningTestSuite) TestGetCertificateFingerprint() {
	fp := GetCertificateFingerprint(s.testCert1)

	// Should be formatted
	s.True(len(fp) > 0)
	s.Equal("SHA256:", fp[:7])

	// Should match the calculated fingerprint when normalized
	normalized := normalizeFingerprint(fp)
	s.Equal(s.fingerprint1, normalized)
}

// Helper function to generate test certificates
func (s *certPinningTestSuite) generateTestCertificate(hostname string) *x509.Certificate {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	s.NoError(err)

	template := &x509.Certificate{
		SerialNumber: big.NewInt(time.Now().Unix()),
		Subject: pkix.Name{
			CommonName:   hostname,
			Organization: []string{"Test Org"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(365 * 24 * time.Hour),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		DNSNames:              []string{hostname},
	}

	certDER, err := x509.CreateCertificate(rand.Reader, template, template, &privateKey.PublicKey, privateKey)
	s.NoError(err)

	cert, err := x509.ParseCertificate(certDER)
	s.NoError(err)

	return cert
}

func (s *certPinningTestSuite) TestMultipleClustersDifferentPins() {
	pins := map[string]PinConfig{
		"cluster1": {
			Fingerprints:  []string{s.fingerprint1},
			StrictPinning: true,
		},
		"cluster2": {
			Fingerprints:  []string{s.fingerprint2},
			StrictPinning: true,
		},
	}

	pinner := NewCertificatePinner(pins, s.logger, s.metricsHandler)

	// cluster1 should accept cert1 but reject cert2
	err := pinner.ValidatePinnedCertificate("cluster1", s.testCert1, true)
	s.NoError(err)
	err = pinner.ValidatePinnedCertificate("cluster1", s.testCert2, true)
	s.Error(err)

	// cluster2 should accept cert2 but reject cert1
	err = pinner.ValidatePinnedCertificate("cluster2", s.testCert2, true)
	s.NoError(err)
	err = pinner.ValidatePinnedCertificate("cluster2", s.testCert1, true)
	s.Error(err)
}

// Test helper to generate a certificate with a specific key
func generateCertificate(t *testing.T, cn string) (*tls.Certificate, *x509.Certificate) {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)

	template := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			CommonName: cn,
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(365 * 24 * time.Hour),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	certDER, err := x509.CreateCertificate(rand.Reader, template, template, &privateKey.PublicKey, privateKey)
	require.NoError(t, err)

	cert, err := x509.ParseCertificate(certDER)
	require.NoError(t, err)

	// Create TLS certificate
	privateKeyBytes, err := x509.MarshalECPrivateKey(privateKey)
	require.NoError(t, err)

	privateKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "EC PRIVATE KEY",
		Bytes: privateKeyBytes,
	})

	certPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certDER,
	})

	tlsCert, err := tls.X509KeyPair(certPEM, privateKeyPEM)
	require.NoError(t, err)

	return &tlsCert, cert
}

func TestFingerprintCalculation(t *testing.T) {
	_, cert := generateCertificate(t, "test.example.com")

	fp1 := calculateFingerprint(cert)
	fp2 := calculateFingerprint(cert)

	// Fingerprints should be consistent
	require.Equal(t, fp1, fp2)

	// Should be 64 hex characters (SHA-256 = 32 bytes = 64 hex chars)
	require.Len(t, fp1, 64)

	// Should be valid hex
	require.NoError(t, ValidateFingerprint(fp1))
}

func TestNormalizeFingerprintFormats(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "lowercase with sha256 prefix",
			input:    "sha256:abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890",
			expected: "abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890",
		},
		{
			name:     "uppercase with SHA256 prefix",
			input:    "SHA256:ABCDEF1234567890ABCDEF1234567890ABCDEF1234567890ABCDEF1234567890",
			expected: "abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890",
		},
		{
			name:     "with colons",
			input:    "AB:CD:EF:12:34:56:78:90:AB:CD:EF:12:34:56:78:90:AB:CD:EF:12:34:56:78:90:AB:CD:EF:12:34:56:78:90",
			expected: "abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890",
		},
		{
			name:     "plain hex",
			input:    "abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890",
			expected: "abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := normalizeFingerprint(tc.input)
			require.Equal(t, tc.expected, result)
		})
	}
}
