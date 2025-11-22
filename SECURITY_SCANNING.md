# Security Scanning Guide

**Last Updated:** 2025-11-22
**Phase:** 3.1 - Security Linting in CI/CD

---

## Overview

Temporal Server uses automated security scanning tools to identify potential security vulnerabilities and code quality issues during development and in CI/CD pipelines.

**Scanning Tools:**
- **gosec** - Static security analysis for Go code
- **govulncheck** - Known vulnerability detection in dependencies

---

## Quick Start

### Run All Security Scans Locally

```bash
make lint-security
```

This runs both gosec and govulncheck.

### Run Individual Scanners

```bash
# Run gosec only
make lint-gosec

# Run govulncheck only
make lint-govulncheck
```

---

## gosec - Static Security Analysis

### What It Does

gosec scans Go source code for common security issues:

- **G101**: Hardcoded credentials (passwords, API keys, tokens)
- **G102**: Bind to all network interfaces
- **G103**: Unsafe use of functions (unsafe, uintptr)
- **G104**: Unhandled errors (complemented by errcheck)
- **G201-G204**: SQL injection vulnerabilities
- **G301-G304**: Insecure file permissions
- **G401-G404**: Weak cryptographic primitives (MD5, DES, etc.)
- **G501-G505**: Weak cryptographic implementations
- **G601**: Implicit memory aliasing in for loops
- And many more...

### Configuration

gosec is configured via `.gosec.yaml`:

```yaml
# Severity threshold (low, medium, high)
severity: medium

# Confidence threshold (low, medium, high)
confidence: medium

# Directories to exclude
exclude-dirs:
  - .bin
  - .stamp
  - vendor

# Exclude generated files
exclude-generated: true

# Rules to exclude
excludes:
  - G104  # Audit errors not checked (use golangci-lint's errcheck instead)
```

### Running Locally

```bash
make lint-gosec
```

Example output:
```
[/home/user/temporal/common/auth/jwt.go:42] - G101 (CWE-798): Potential hardcoded credentials
  41:   func parseToken(token string) {
  42:       secret := "my-secret-key"  // BAD: Hardcoded secret
  43:       // ...
```

### Handling Findings

1. **True Positives**: Fix the security issue immediately
2. **False Positives**: Add a `#nosec` comment with justification

```go
// #nosec G104 -- Error intentionally ignored in test cleanup
defer os.RemoveAll(tempDir)
```

**IMPORTANT**: Always provide a reason when using `#nosec`!

### Integration in CI/CD

gosec runs automatically on every pull request via GitHub Actions:

```yaml
# .github/workflows/linters.yml
security-scan:
  runs-on: ubuntu-latest
  steps:
    - uses: actions/checkout@v4
    - uses: actions/setup-go@v5
    - name: run gosec security scanner
      run: make lint-gosec
```

Pull requests with medium+ severity issues will fail the check.

---

## govulncheck - Vulnerability Scanning

### What It Does

govulncheck detects known vulnerabilities in:
- Direct dependencies
- Transitive dependencies
- Go standard library

It queries the Go vulnerability database (https://vuln.go.dev) for CVEs affecting your dependencies.

### Running Locally

```bash
make lint-govulncheck
```

Example output:
```
govulncheck is an experimental tool. Share feedback at https://go.dev/s/govulncheck-feedback.

Scanning your code and 500 packages across 75 dependent modules for known vulnerabilities...

Vulnerability #1: GO-2024-1234
  Package: golang.org/x/crypto
  Version: v0.14.0
  Fixed in: v0.15.0
  Description: Timing attack in AES-GCM implementation
  More info: https://pkg.go.dev/vuln/GO-2024-1234
```

### Handling Vulnerabilities

#### Critical/High Severity

1. Update dependency immediately:
   ```bash
   go get golang.org/x/crypto@latest
   go mod tidy
   ```

2. Test thoroughly after upgrade

3. Create PR with clear description of vulnerability fixed

#### Medium/Low Severity

1. **Assess Impact**: Does the vulnerability affect code paths we use?
2. **Timeline**: Plan upgrade within sprint (medium) or next quarter (low)
3. **Track**: Create GitHub issue to track remediation

#### No Fix Available

1. **Assess Risk**: Can we mitigate through configuration or code changes?
2. **Document**: Add to known issues in security docs
3. **Monitor**: Set up alerting for when patch is available

### Integration in CI/CD

govulncheck runs on every pull request:

```yaml
# .github/workflows/linters.yml
- name: run govulncheck vulnerability scanner
  run: make lint-govulncheck
```

**Note**: govulncheck may be configured to only fail on high/critical severity to avoid blocking development on low-risk issues.

---

## CI/CD Integration

### GitHub Actions Workflow

Security scanning is integrated into the `linters` workflow:

**File**: `.github/workflows/linters.yml`

```yaml
security-scan:
  runs-on: ubuntu-latest
  steps:
    - uses: actions/checkout@v4
      with:
        fetch-depth: 0

    - uses: actions/setup-go@v5
      with:
        go-version-file: 'go.mod'
        check-latest: true

    - name: run gosec security scanner
      run: make lint-gosec

    - name: run govulncheck vulnerability scanner
      run: make lint-govulncheck
```

### When Scans Run

- ✅ Every pull request
- ✅ Merges to main branch
- ✅ Scheduled nightly scans (recommended, not yet implemented)

### Failing Checks

If security scans fail:

1. **Review the findings** in the GitHub Actions log
2. **Assess severity** and impact
3. **Fix or justify** each finding
4. **Push fixes** to the PR branch
5. **Re-run checks** automatically

---

## Best Practices

### For Developers

1. **Run locally before pushing**:
   ```bash
   make lint-security
   ```

2. **Never commit secrets**:
   - Use environment variables
   - Use secret management systems (Vault, AWS Secrets Manager)
   - Never hardcode API keys, passwords, or tokens

3. **Keep dependencies updated**:
   ```bash
   go get -u ./...
   go mod tidy
   ```

4. **Use strong cryptography**:
   - Never use MD5, SHA1 for security purposes
   - Use bcrypt/scrypt/argon2 for password hashing
   - Use AES-256-GCM or ChaCha20-Poly1305 for encryption

5. **Handle errors properly**:
   - Check all errors, especially security-critical ones
   - Use golangci-lint's errcheck in addition to gosec

### For Security Team

1. **Review security scan results weekly**
2. **Triage new vulnerabilities within 24 hours**
3. **Track remediation in GitHub issues**
4. **Update .gosec.yaml as needed to reduce false positives**
5. **Conduct quarterly reviews of excluded rules**

### For Operators

1. **Monitor vulnerability announcements**: https://vuln.go.dev
2. **Subscribe to Go security mailing list**: https://groups.google.com/g/golang-announce
3. **Plan regular dependency updates** (monthly recommended)

---

## Troubleshooting

### gosec Reports False Positive

Add `#nosec` with justification:

```go
// #nosec G304 -- File path is validated and sanitized before use
func readConfig(path string) error {
    data, err := os.ReadFile(path)
    // ...
}
```

### govulncheck Fails to Download Database

Check network connectivity:
```bash
curl -I https://vuln.go.dev
```

Set proxy if needed:
```bash
export HTTPS_PROXY=http://proxy.example.com:8080
make lint-govulncheck
```

### Too Many gosec Findings

1. Focus on high-severity issues first
2. Update .gosec.yaml to increase severity threshold temporarily:
   ```yaml
   severity: high  # Only report high severity
   ```
3. Create remediation plan for medium/low severity issues

### Dependency Has Vulnerability But No Update Available

1. Check if vulnerability affects your usage:
   ```bash
   govulncheck -show verbose ./...
   ```

2. If not reachable from your code, document and monitor

3. If reachable, consider:
   - Finding alternative dependency
   - Forking and patching locally
   - Implementing workaround in application code

---

## Metrics and Monitoring

### Track Security Scan Results

Recommended metrics to track:

- **Scan Success Rate**: % of PRs passing security scans
- **Mean Time to Remediate**: Time from vulnerability detection to fix
- **Vulnerability Count by Severity**: Track over time
- **False Positive Rate**: % of findings marked as false positives

### Alerting

Set up alerts for:

- Critical vulnerabilities detected (Slack/email notification)
- Repeated security scan failures on main branch
- New CVEs affecting dependencies (GitHub Dependabot)

---

## Additional Resources

- **gosec Documentation**: https://github.com/securego/gosec
- **gosec Rules Reference**: https://github.com/securego/gosec#available-rules
- **Go Vulnerability Database**: https://vuln.go.dev
- **govulncheck Documentation**: https://pkg.go.dev/golang.org/x/vuln/cmd/govulncheck
- **OWASP Go Security Cheat Sheet**: https://cheatsheetseries.owasp.org/cheatsheets/Go_SCP_Cheat_Sheet.html

---

## Version History

| Date | Version | Changes |
|------|---------|---------|
| 2025-11-22 | 1.0 | Initial implementation (Phase 3.1) |

---

**Maintained By:** Security Team
**Questions:** security@temporal.io
