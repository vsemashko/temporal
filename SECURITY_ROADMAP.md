# Security Improvements Roadmap

**Project:** Temporal Server Security Enhancements
**Version:** 1.0
**Last Updated:** 2025-11-22

---

## Executive Summary

This roadmap outlines the remaining security improvements for Temporal Server following the completion of Phase 1 (Critical and High Priority fixes). The work is organized into three phases based on priority and impact.

**Current Status:**
- ✅ Phase 1: Critical & High Priority - **COMPLETED**
- 🔄 Phase 2: Medium Priority - **IN PROGRESS (33% COMPLETE)**
- 📋 Phase 3: Low Priority - **PLANNED**

**Latest Milestone:**
- ✅ Phase 2.1: Authentication Rate Limiting - **COMPLETED** (2025-11-22)

---

## Phase 1: Critical & High Priority (COMPLETED) ✅

### Completed Items

1. ✅ **TLS 1.3 Default Configuration**
   - Status: COMPLETE
   - Effort: 1 day
   - Impact: HIGH

2. ✅ **Explicit Cipher Suite Configuration**
   - Status: COMPLETE
   - Effort: 1 day
   - Impact: HIGH

3. ✅ **Noop Authorizer Security Warnings**
   - Status: COMPLETE
   - Effort: 0.5 days
   - Impact: MEDIUM

4. ✅ **Generic Authorization Error Messages**
   - Status: COMPLETE
   - Effort: 1 day
   - Impact: HIGH

5. ✅ **TLS Host Verification Warnings**
   - Status: COMPLETE
   - Effort: 0.5 days
   - Impact: MEDIUM

6. ✅ **Test Credentials Warnings**
   - Status: COMPLETE
   - Effort: 0.25 days
   - Impact: LOW

7. ✅ **HTTP Security Headers**
   - Status: COMPLETE
   - Effort: 0.5 days
   - Impact: MEDIUM

**Total Phase 1 Effort:** 4.75 days
**Completion Date:** 2025-11-22

---

## Phase 2: Medium Priority Improvements

**Started:** 2025-11-22
**Target Completion:** 2025-12-15
**Estimated Effort:** 12-17 days
**Progress:** 1/3 items complete (33%)

### 2.1 Authentication Rate Limiting ✅

**Status:** **COMPLETED** (2025-11-22)
**Priority:** HIGH (within Medium category)
**Actual Effort:** 3 days
**Commits:** 47cc09b, 43cd6d9

**Description:**
Implement rate limiting specifically for authentication attempts to prevent brute force attacks on JWT tokens or certificate-based authentication.

**Implementation Plan:**

1. **Create Authentication Rate Limiter** (1 day)
   ```go
   // File: common/rpc/interceptor/auth_rate_limit.go
   type AuthRateLimitInterceptor struct {
       rateLimiter quotas.RequestRateLimiter
       failureTracker map[string]*authFailureTracker
       mu sync.RWMutex
   }
   ```

2. **Track Failed Authentication Attempts** (1 day)
   - Per source IP address
   - Per user identity
   - Exponential backoff for repeated failures

3. **Add Metrics** (0.5 days)
   - `temporal_auth_failures_total`
   - `temporal_auth_rate_limited_total`
   - `temporal_auth_failure_rate_per_ip`

4. **Configuration** (0.5 days)
   ```yaml
   global:
     authorization:
       rateLimit:
         enabled: true
         maxAttemptsPerMinute: 10
         lockoutDuration: 5m
         trackByIP: true
   ```

5. **Testing** (1-2 days)
   - Unit tests for rate limiting logic
   - Integration tests for authentication failures
   - Load tests to verify performance impact

**Acceptance Criteria:**
- [x] Authentication failures are rate-limited per IP
- [x] Sliding window implemented for failure tracking
- [x] Metrics tracking auth failure rates (5 new metrics)
- [x] Configuration options for rate limits
- [x] Documentation for operators
- [x] 11 comprehensive unit tests
- [x] Integration into frontend service

**Deliverables Completed:**
- ✅ `common/rpc/interceptor/auth_rate_limit.go` - Core implementation (285 lines)
- ✅ `common/rpc/interceptor/auth_rate_limit_test.go` - Unit tests (11 test cases)
- ✅ `service/frontend/fx.go` - Frontend integration
- ✅ `common/config/config.go` - Configuration schema
- ✅ `common/metrics/metric_defs.go` - 5 new metrics
- ✅ `SECURITY_OPERATOR_GUIDE.md` - Complete operator documentation

**Dependencies:** None

**Risk:** Low - Isolated feature, easy to disable if issues arise ✅ Mitigated

---

### 2.2 Certificate Pinning for Remote Clusters

**Priority:** MEDIUM
**Effort:** 5-7 days
**Assignee:** TBD

**Description:**
Implement certificate pinning for remote cluster connections to provide additional protection against compromised Certificate Authorities.

**Implementation Plan:**

1. **Design Certificate Pinning Mechanism** (1 day)
   - Support SHA-256 fingerprint pinning
   - Support public key pinning
   - Graceful handling of pinning failures

2. **Configuration Schema** (1 day)
   ```yaml
   global:
     tls:
       remoteClusters:
         cluster1:
           client:
             serverName: "cluster1.example.com"
             pinnedCertificates:
               - fingerprint: "sha256:abc123..."
                 description: "Production cert expires 2026-01"
               - fingerprint: "sha256:def456..."
                 description: "Backup cert"
             strictPinning: true  # Fail if no pins match
   ```

3. **Implementation** (2-3 days)
   ```go
   // File: common/rpc/encryption/cert_pinning.go
   type CertificatePinner interface {
       ValidatePinnedCertificate(cert *x509.Certificate) error
   }
   ```

4. **Metrics and Logging** (1 day)
   - Log when pinned cert is validated
   - Alert when pinning fails
   - Metric for pinning validation rate

5. **Testing and Documentation** (2 days)
   - Unit tests for pinning logic
   - Integration tests with test certificates
   - Operator documentation for pinning setup

**Acceptance Criteria:**
- [ ] Certificate pinning configurable per remote cluster
- [ ] Support SHA-256 fingerprint pinning
- [ ] Graceful degradation if pinning misconfigured
- [ ] Metrics and alerts for pinning failures
- [ ] Documentation with examples

**Dependencies:** None

**Risk:** Medium - Incorrect configuration could break remote cluster connections

---

### 2.3 Secrets Rotation Documentation & Tooling

**Priority:** MEDIUM
**Effort:** 2-3 days
**Assignee:** TBD

**Description:**
Create comprehensive documentation and tooling to support secrets rotation without downtime.

**Implementation Plan:**

1. **Document Manual Rotation Procedures** (1 day)
   - TLS certificate rotation
   - JWT signing key rotation
   - Database password rotation
   - API key rotation

2. **Create Rotation Helper Scripts** (1 day)
   ```bash
   # scripts/rotate-certificates.sh
   # scripts/rotate-jwt-keys.sh
   # scripts/validate-rotation.sh
   ```

3. **Testing Guide** (0.5 days)
   - How to test rotation in staging
   - Validation checklist
   - Rollback procedures

4. **Operational Runbook** (0.5 days)
   - Step-by-step rotation guide
   - Monitoring during rotation
   - Common issues and resolutions

**Deliverables:**
- [ ] `docs/operations/secrets-rotation.md`
- [ ] Helper scripts in `scripts/security/`
- [ ] Runbook for on-call engineers
- [ ] Validation checklist

**Dependencies:** None

**Risk:** Low - Documentation only

---

### 2.4 Enhanced Runtime Security Warnings

**Priority:** MEDIUM
**Effort:** 2-3 days
**Assignee:** TBD

**Description:**
Add runtime logging and metrics for insecure configurations detected at startup or runtime.

**Implementation Plan:**

1. **Startup Security Audit** (1 day)
   ```go
   // File: temporal/server/security_audit.go
   func AuditSecurityConfiguration(config *config.Config) []SecurityWarning {
       warnings := []SecurityWarning{}

       if isNoopAuthorizerConfigured(config) {
           warnings = append(warnings, SecurityWarning{
               Severity: CRITICAL,
               Message: "Noop authorizer is active - all requests allowed without auth",
           })
       }

       if isTLSHostVerificationDisabled(config) {
           warnings = append(warnings, SecurityWarning{
               Severity: HIGH,
               Message: "TLS host verification is disabled - MITM attacks possible",
           })
       }

       return warnings
   }
   ```

2. **Runtime Metrics** (1 day)
   - `temporal_security_noop_auth_active` (gauge, 0 or 1)
   - `temporal_security_tls_verification_disabled` (gauge, 0 or 1)
   - `temporal_security_warnings_total` (counter by severity)

3. **Startup Warnings** (0.5 days)
   - Log CRITICAL warnings in RED/bold
   - Emit warnings to stderr
   - Optional: Fail startup if critical warnings present

4. **Configuration Option** (0.5 days)
   ```yaml
   global:
     security:
       auditOnStartup: true
       failOnCriticalWarnings: false  # Set to true in production
   ```

**Acceptance Criteria:**
- [ ] Security audit runs at server startup
- [ ] Critical warnings logged prominently
- [ ] Metrics track insecure configurations
- [ ] Optional fail-fast mode for production

**Dependencies:** None

**Risk:** Low - Non-breaking change

---

### 2.5 TLS Configuration Validation

**Priority:** LOW-MEDIUM
**Effort:** 1-2 days
**Assignee:** TBD

**Description:**
Add comprehensive validation of TLS configurations at startup to catch misconfigurations early.

**Implementation Plan:**

1. **Enhanced Validation** (1 day)
   - Verify certificates are valid PEM format
   - Check certificate and key match
   - Validate certificate expiration
   - Verify CA certificates can validate server certs
   - Check for weak key sizes (< 2048 bits)

2. **Helpful Error Messages** (0.5 days)
   - Clear error messages for common issues
   - Suggestions for resolution
   - Link to documentation

3. **Testing** (0.5 days)
   - Test with invalid certificates
   - Test with expired certificates
   - Test with mismatched cert/key pairs

**Acceptance Criteria:**
- [ ] Invalid TLS configurations fail startup
- [ ] Clear error messages guide operators
- [ ] Validation covers common misconfigurations

**Dependencies:** None

**Risk:** Low - Validation only, no functional changes

---

## Phase 3: Low Priority Enhancements

**Target Start:** 2026-01-01
**Target Completion:** 2026-02-15
**Estimated Effort:** 10-15 days

### 3.1 Security Linting in CI/CD

**Priority:** LOW
**Effort:** 2-3 days

**Description:**
Integrate security scanning tools into the CI/CD pipeline to automatically detect vulnerabilities.

**Tools to Integrate:**
1. **gosec** - Go security scanner
2. **govulncheck** - Go vulnerability scanner
3. **trivy** - Container image scanner
4. **dependabot** - Dependency update automation

**Implementation Plan:**

1. **Add gosec to CI** (0.5 days)
   ```yaml
   # .github/workflows/security.yml
   - name: Run gosec
     run: gosec -fmt sarif -out gosec.sarif ./...
   ```

2. **Add govulncheck** (0.5 days)
   ```yaml
   - name: Check for vulnerabilities
     run: govulncheck ./...
   ```

3. **Add trivy scanning** (0.5 days)
   ```yaml
   - name: Scan Docker images
     run: trivy image --severity HIGH,CRITICAL temporal:latest
   ```

4. **Configure Dependabot** (0.5 days)
   ```yaml
   # .github/dependabot.yml
   version: 2
   updates:
     - package-ecosystem: "gomod"
       directory: "/"
       schedule:
         interval: "weekly"
   ```

5. **Dashboard and Reporting** (1 day)
   - Security scan results in GitHub Security tab
   - Slack notifications for critical issues
   - Weekly security report

**Acceptance Criteria:**
- [ ] Security scans run on every PR
- [ ] Critical vulnerabilities block merges
- [ ] Weekly security reports generated
- [ ] Automated dependency updates

---

### 3.2 Comprehensive Audit Logging

**Priority:** LOW
**Effort:** 4-5 days

**Description:**
Implement structured audit logging for all authorization decisions and security-relevant events.

**Implementation Plan:**

1. **Audit Log Structure** (1 day)
   ```go
   type AuditEvent struct {
       Timestamp   time.Time
       EventType   string // "auth.success", "auth.failure", "config.change"
       UserID      string
       SourceIP    string
       Namespace   string
       Operation   string
       Result      string // "allowed", "denied"
       Reason      string
       Metadata    map[string]string
   }
   ```

2. **Audit Logger Implementation** (1 day)
   - Structured logging (JSON format)
   - Separate audit log file
   - Log rotation support
   - SIEM integration support

3. **Audit Events** (2 days)
   - Authentication success/failure
   - Authorization decisions
   - Configuration changes
   - Certificate rotations
   - Namespace operations
   - Admin operations

4. **Query and Analysis Tools** (1 day)
   - CLI tool to query audit logs
   - Example queries for common investigations
   - Integration with Elasticsearch

**Acceptance Criteria:**
- [ ] All authorization decisions logged
- [ ] Structured JSON format
- [ ] Queryable audit trail
- [ ] SIEM integration examples

---

### 3.3 Configuration Sanitization

**Priority:** LOW
**Effort:** 1-2 days

**Description:**
Automatically sanitize sensitive fields when logging configuration to prevent password leakage.

**Implementation:**

```go
// File: common/config/sanitizer.go
func SanitizeConfig(cfg *Config) *Config {
    sanitized := *cfg

    // Sanitize database passwords
    for name, store := range sanitized.Persistence.DataStores {
        if store.SQL != nil {
            store.SQL.ConnectAttributes["password"] = "***REDACTED***"
        }
    }

    // Sanitize TLS private keys
    if sanitized.Global.TLS.Frontend.Server.KeyData != "" {
        sanitized.Global.TLS.Frontend.Server.KeyData = "***REDACTED***"
    }

    return &sanitized
}
```

**Acceptance Criteria:**
- [ ] Passwords never logged
- [ ] Private keys never logged
- [ ] API keys never logged
- [ ] Config logging uses sanitized version

---

### 3.4 Context Timeout Enforcement

**Priority:** LOW
**Effort:** 2-3 days

**Description:**
Enforce default context timeouts for all operations to prevent resource exhaustion.

**Implementation:**

```go
// File: common/rpc/interceptor/timeout.go
type TimeoutInterceptor struct {
    defaultTimeout time.Duration
    methodTimeouts map[string]time.Duration
}

func (i *TimeoutInterceptor) Intercept(
    ctx context.Context,
    req interface{},
    info *grpc.UnaryServerInfo,
    handler grpc.UnaryHandler,
) (interface{}, error) {
    timeout := i.getTimeout(info.FullMethod)
    ctx, cancel := context.WithTimeout(ctx, timeout)
    defer cancel()

    return handler(ctx, req)
}
```

**Configuration:**
```yaml
global:
  rpc:
    defaultTimeout: 30s
    methodTimeouts:
      "/temporal.api.workflowservice.v1.WorkflowService/StartWorkflowExecution": 10s
      "/temporal.api.workflowservice.v1.WorkflowService/GetWorkflowExecutionHistory": 60s
```

**Acceptance Criteria:**
- [ ] Default timeouts for all operations
- [ ] Configurable per-method timeouts
- [ ] Metrics for timeout occurrences

---

### 3.5 Dependency Management Automation

**Priority:** LOW
**Effort:** 1-2 days

**Description:**
Automate dependency updates and security vulnerability scanning.

**Implementation:**

1. **Automated Dependency Updates** (0.5 days)
   - Dependabot configuration
   - Automated PR creation
   - Automated testing of updates

2. **Vulnerability Scanning** (0.5 days)
   - Daily vulnerability scans
   - Alert on new vulnerabilities
   - Track remediation progress

3. **Dependency Dashboard** (1 day)
   - Show all dependencies with versions
   - Highlight outdated dependencies
   - Show security vulnerabilities

**Acceptance Criteria:**
- [ ] Weekly dependency update PRs
- [ ] Automated vulnerability scanning
- [ ] Dashboard showing dependency health

---

## Implementation Timeline

### Q4 2025 (Current Quarter)
- ✅ Phase 1: Critical & High Priority - COMPLETED

### Q1 2026
- 🔄 Phase 2: Medium Priority
  - Week 1-2: Authentication Rate Limiting
  - Week 3-4: Certificate Pinning
  - Week 5-6: Secrets Rotation Documentation
  - Week 7-8: Runtime Security Warnings
  - Week 9: TLS Configuration Validation
  - Week 10: Buffer for testing and documentation

### Q2 2026
- 📋 Phase 3: Low Priority
  - Weeks 1-2: Security Linting in CI/CD
  - Weeks 3-6: Comprehensive Audit Logging
  - Week 7: Configuration Sanitization
  - Weeks 8-9: Context Timeout Enforcement
  - Week 10: Dependency Management Automation
  - Weeks 11-12: Integration testing and documentation

---

## Resource Requirements

### Phase 2 (Medium Priority)
- **Engineers:** 1-2 senior engineers
- **Time:** 2.5 months (10 weeks)
- **Effort:** 12-17 engineering days
- **Budget:** Minimal (existing resources)

### Phase 3 (Low Priority)
- **Engineers:** 1 senior engineer
- **Time:** 1.5 months (6 weeks)
- **Effort:** 10-15 engineering days
- **Budget:** Potential tooling costs (< $1000/month for scanning services)

---

## Success Metrics

### Phase 2 Metrics
- Authentication brute force attempts detected and blocked
- Certificate pinning violations detected (should be zero)
- Successful secrets rotations without downtime
- Zero critical security warnings in production

### Phase 3 Metrics
- Security vulnerabilities detected in CI before merge
- 100% of security events logged to audit trail
- Zero passwords/keys logged in configuration
- All API operations complete within timeout limits
- Dependency vulnerabilities resolved within 7 days

---

## Risk Management

### High Risk Items
1. **Authentication Rate Limiting** - ✅ COMPLETED - Could block legitimate users
   - Mitigation: ✅ Implemented - Configurable thresholds, comprehensive monitoring, easy disable, fail-open design
   - Result: Disabled by default, operators opt-in with appropriate thresholds

2. **Certificate Pinning** - 🔵 NEXT - Could break connections if misconfigured
   - Mitigation: Graceful degradation, clear error messages, testing

### Medium Risk Items
1. **Timeout Enforcement** - Could break long-running operations
   - Mitigation: Per-method configuration, gradual rollout

### Low Risk Items
- All documentation and monitoring improvements are low risk
- Security linting is non-blocking by default
- Audit logging is additive only

---

## Decision Log

| Date | Decision | Rationale |
|------|----------|-----------|
| 2025-11-22 | Prioritize auth rate limiting in Phase 2 | High impact, prevents brute force attacks ✅ COMPLETED |
| 2025-11-22 | Auth rate limiting disabled by default | Opt-in approach prevents disruption to existing deployments |
| 2025-11-22 | Fail-open design for auth rate limiter | If IP cannot be extracted, allow request (availability over security) |
| 2025-11-22 | Defer SIEM integration to Phase 3 | Requires audit logging foundation first |
| 2025-11-22 | Make security linting non-blocking initially | Reduce risk of false positives breaking builds |

---

## Questions and Open Items

1. **Q:** Should we backport Phase 1 fixes to older versions?
   **A:** TBD - Need to assess support policy and customer impact

2. **Q:** Should authentication rate limiting be enabled by default?
   **A:** ✅ RESOLVED - Disabled by default (opt-in). Rationale: Prevents disruption to existing deployments. Operators can enable with appropriate thresholds for their environment. Default when enabled: 10 failures/min, 5min lockout.

3. **Q:** What SIEM systems should we support for audit log integration?
   **A:** TBD - Survey customer requirements (Splunk, ELK, Datadog likely candidates)

---

## Appendix: Deferred Items

Items explicitly deferred or rejected:

1. **Web Application Firewall (WAF)** - Deferred
   - Reason: Most customers use external WAF (Cloudflare, AWS WAF)
   - Better handled at infrastructure layer

2. **Intrusion Detection System (IDS)** - Deferred
   - Reason: Network-level concern, not application-level
   - Better handled by network security team

3. **Automatic Certificate Renewal** - Deferred
   - Reason: Requires integration with various CAs (Let's Encrypt, etc.)
   - Complex implementation, customers have existing solutions

4. **Hardware Security Module (HSM) Integration** - Deferred
   - Reason: Niche requirement, high complexity
   - Can be implemented as custom plugin if needed

---

**Roadmap Owner:** Security Team
**Last Reviewed:** 2025-11-22
**Next Review:** 2026-01-01
**Status:** Active
