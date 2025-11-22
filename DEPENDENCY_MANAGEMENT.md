# Dependency Management & Vulnerability Response

**Last Updated:** 2025-11-22
**Phase:** 3.5 - Dependency Scanning Automation

---

## Overview

Temporal Server uses automated dependency management to:
- Keep dependencies up-to-date with latest security patches
- Detect known vulnerabilities in dependencies
- Reduce manual maintenance burden
- Improve overall security posture

**Tools:**
- **Dependabot** - Automated dependency updates and security alerts
- **govulncheck** - Go vulnerability scanner (see SECURITY_SCANNING.md)
- **GitHub Security Advisories** - Vulnerability database

---

## Quick Reference

### For Developers

```bash
# Check for outdated dependencies
go list -u -m all

# Update all dependencies to latest minor/patch
go get -u ./...
go mod tidy

# Update specific dependency
go get github.com/example/package@latest
go mod tidy

# Check for vulnerabilities
make lint-govulncheck
```

### For Maintainers

**Dependabot PR Review Checklist:**
- [ ] Read the changelog and release notes
- [ ] Check for breaking changes
- [ ] Review security impact
- [ ] Verify CI/CD passes
- [ ] Test locally if major version bump
- [ ] Merge or request changes

---

## Dependabot Configuration

Dependabot is configured via `.github/dependabot.yml`:

```yaml
version: 2
updates:
  - package-ecosystem: "gomod"
    directory: "/"
    schedule:
      interval: "weekly"
      day: "monday"
      time: "09:00"
```

### Update Schedule

| Ecosystem | Frequency | Day | Time (PST) |
|-----------|-----------|-----|------------|
| Go modules | Weekly | Monday | 09:00 |
| GitHub Actions | Weekly | Monday | 09:00 |

### Grouping Strategy

**Security Updates:**
- Grouped separately for priority review
- Auto-created immediately when vulnerability detected
- Labeled `dependencies`, `security`

**Minor & Patch Updates:**
- Grouped together weekly
- Reduces PR noise
- Labeled `dependencies`, `automated`

**Major Version Updates:**
- Created individually
- Requires careful review for breaking changes
- Labeled `dependencies`, `major-version`

---

## Vulnerability Response Workflow

### 1. Detection

Vulnerabilities are detected through:
- **Dependabot Security Alerts** - GitHub automatically scans dependencies
- **govulncheck** - CI/CD pipeline on every PR
- **Manual Reports** - Team members or security researchers

### 2. Triage (Within 24 Hours for Critical/High)

**Severity Classification:**

| Severity | CVSS Score | Response Time | Action |
|----------|------------|---------------|--------|
| **Critical** | 9.0-10.0 | < 4 hours | Emergency patch |
| **High** | 7.0-8.9 | < 24 hours | Priority patch |
| **Medium** | 4.0-6.9 | < 1 week | Scheduled patch |
| **Low** | 0.1-3.9 | < 1 month | Bundled update |

**Triage Questions:**
1. Does the vulnerability affect code paths we use?
2. Is the vulnerability reachable from external input?
3. What is the potential impact (confidentiality, integrity, availability)?
4. Is a patch/fix available?
5. Are there workarounds?

### 3. Assessment

**Check Exploitability:**

```bash
# Use govulncheck to see if vulnerability is reachable
govulncheck -show verbose ./...
```

Example output:
```
Vulnerability #1: GO-2024-1234
  Your code is affected by GO-2024-1234.

  Call stacks in your code:
    main.go:42:10: example.VulnerableFunc calls vulnerable.Function
```

If **NOT reachable**: Lower priority, document, and schedule for next update cycle.

If **REACHABLE**: Follow remediation workflow immediately.

### 4. Remediation

#### Option A: Update Dependency (Preferred)

```bash
# Update to patched version
go get github.com/vulnerable/package@v1.2.3
go mod tidy

# Verify vulnerability is fixed
make lint-govulncheck

# Run full test suite
make test

# Create PR with clear description
git checkout -b fix/cve-2024-1234
git add go.mod go.sum
git commit -m "fix(deps): update vulnerable/package to v1.2.3 (CVE-2024-1234)"
git push origin fix/cve-2024-1234
```

**PR Description Template:**
```markdown
## Security Fix: CVE-2024-1234

**Vulnerability:** [Description]
**Severity:** High (CVSS 8.5)
**Affected Package:** github.com/vulnerable/package v1.2.0
**Fixed Version:** v1.2.3

### Changes
- Updated package from v1.2.0 to v1.2.3

### Testing
- [x] All tests pass
- [x] govulncheck shows no vulnerabilities
- [x] Verified fix addresses CVE-2024-1234

### References
- CVE: https://nvd.nist.gov/vuln/detail/CVE-2024-1234
- Advisory: https://github.com/advisories/GHSA-xxxx-xxxx-xxxx
```

#### Option B: Workaround (If No Fix Available)

1. **Assess Risk:** Can we avoid using the vulnerable code path?
2. **Implement Mitigation:**
   - Input validation
   - Additional access controls
   - Disable vulnerable feature
3. **Document:** Add to known issues, set up monitoring for patch
4. **Create Tracking Issue:**
   ```markdown
   Title: [Security] Track CVE-2024-1234 remediation

   **Status:** Awaiting upstream patch
   **Severity:** High
   **Package:** github.com/vulnerable/package
   **Workaround:** [Description of mitigation]
   **Monitoring:** Check weekly for patch availability
   ```

#### Option C: Replace Dependency

If vulnerability won't be patched or dependency is abandoned:

1. Research alternative packages
2. Evaluate maturity, security, and maintenance
3. Create spike/POC for replacement
4. Plan migration (may require significant effort)
5. Execute replacement with thorough testing

### 5. Verification

```bash
# Verify vulnerability is resolved
make lint-govulncheck

# Run full test suite
make test

# Integration tests
make test-integration

# Manual verification if needed
# ...
```

### 6. Communication

**Internal:**
- Update GitHub issue/PR with resolution
- Notify team in Slack #security channel
- Update security metrics dashboard

**External (if applicable):**
- Security advisory if CVE affects users
- Blog post for significant vulnerabilities
- Update release notes

---

## Handling Dependabot PRs

### Automatic Merge Criteria

Dependabot PRs can be auto-merged if ALL of the following are true:
- ✅ Update is patch or minor version (not major)
- ✅ All CI/CD checks pass
- ✅ No known breaking changes in changelog
- ✅ Security update OR grouped minor/patch update
- ✅ No manual review required label

### Manual Review Required

Review manually if ANY of the following:
- 🔴 Major version bump
- 🔴 CI/CD failures
- 🔴 Known breaking changes in changelog
- 🔴 Core dependency (gRPC, protobuf, database drivers)
- 🔴 Security-critical package (TLS, JWT, crypto)

### Review Process

1. **Read the PR description**
   - Dependabot includes changelog and release notes
   - Check "compatibility score" if provided

2. **Review changes**
   ```bash
   # View dependency diff
   gh pr diff [PR-NUMBER]

   # Check for breaking changes
   go doc -all github.com/package/name@newversion | grep -i "deprecated\|breaking"
   ```

3. **Test locally (for major updates)**
   ```bash
   gh pr checkout [PR-NUMBER]
   make test
   make test-integration
   ```

4. **Approve or request changes**
   ```bash
   # Approve and merge
   gh pr review [PR-NUMBER] --approve
   gh pr merge [PR-NUMBER] --squash

   # Request changes
   gh pr review [PR-NUMBER] --request-changes --body "..."
   ```

---

## Dependency Update Policy

### Regular Updates

**Schedule:**
- **Security Patches:** Immediate (as soon as Dependabot creates PR)
- **Minor/Patch Updates:** Weekly (bundled)
- **Major Updates:** Quarterly (planned with testing)

**Process:**
1. Dependabot creates PR on Monday 09:00 PST
2. Maintainers review within 2 business days
3. CI/CD validates changes
4. Merge if all checks pass
5. Monitor for issues post-merge

### Emergency Updates

For critical security vulnerabilities:

1. **Detection** (within 1 hour of disclosure)
2. **Assessment** (within 2 hours)
3. **Patch Development** (within 4 hours)
4. **Testing** (within 6 hours)
5. **Deployment** (within 8 hours of disclosure)

Emergency update process may skip normal review for time-critical fixes.

---

## Dependency Pinning Strategy

### When to Pin

**DO pin:**
- ✅ Production deployments (use exact versions)
- ✅ Core dependencies with stability requirements
- ✅ When a specific version fixes a critical bug

**DO NOT pin:**
- ❌ Development/test dependencies (use `^` or `~`)
- ❌ Transitive dependencies (let Go modules handle)
- ❌ Without good reason (makes security updates harder)

### Pinning Syntax

```go
// go.mod

require (
    // Exact version (pinned)
    github.com/critical/package v1.2.3

    // Latest compatible (unpinned)
    github.com/dev/tool v1.2.0  // Go modules uses minimum version selection
)
```

---

## Metrics & Monitoring

### Key Metrics

Track these metrics monthly:

| Metric | Target | Current |
|--------|--------|---------|
| Mean Time to Patch (MTTP) | < 7 days | TBD |
| % Dependencies Up-to-Date | > 90% | TBD |
| Critical Vulns Open > 7 Days | 0 | TBD |
| Dependabot PRs Merged | > 80% | TBD |

### Dashboards

**GitHub Security Dashboard:**
- Navigate to: Repository → Security → Dependabot alerts
- View: Open vulnerabilities, dismissed alerts, resolved
- Filter by severity, package ecosystem

**Dependabot Insights:**
- Navigate to: Repository → Insights → Dependency graph → Dependabot
- View: Open PRs, update frequency, merge rate

---

## Troubleshooting

### Dependabot PR Fails CI

**Common causes:**
1. **Breaking API changes**
   - Solution: Update code to use new API
   - Or: Request `@dependabot ignore this major version`

2. **Test failures**
   - Solution: Fix tests to work with new version
   - Or: Investigate if dependency introduced bugs

3. **Dependency conflict**
   - Solution: Update conflicting dependencies together
   - Or: Use `go mod why` to understand dependency chain

### Dependabot Isn't Creating PRs

**Checklist:**
1. Is Dependabot enabled? (Settings → Security → Dependabot)
2. Is `.github/dependabot.yml` valid? (check for YAML syntax errors)
3. Have you hit the `open-pull-requests-limit`?
4. Are updates paused? (check Dependabot settings)

### Vulnerability Shows as "Dismissed"

**Reasons for dismissal:**
1. **No patch available** - Waiting for upstream fix
2. **Not reachable** - Code path not used
3. **Low severity** - Bundled with next update
4. **False positive** - Reported incorrectly

**Review dismissals quarterly** to ensure they're still valid.

---

## Best Practices

### For Developers

1. **Keep dependencies minimal**
   - Only add necessary dependencies
   - Avoid "convenience" libraries for simple tasks
   - Regularly audit `go.mod` for unused dependencies

2. **Update frequently**
   ```bash
   # Weekly dependency check
   go list -u -m all | grep '\['
   ```

3. **Test before merging**
   - Don't blindly merge Dependabot PRs
   - Run tests locally for major updates
   - Check changelogs for breaking changes

4. **Document decisions**
   - If ignoring a Dependabot suggestion, comment why
   - Use `@dependabot ignore` commands in PR comments

### For Maintainers

1. **Triage security alerts daily**
2. **Review Dependabot PRs within 2 business days**
3. **Keep `.github/dependabot.yml` up-to-date**
4. **Monitor dependency health** (use tools like `go mod graph`)
5. **Plan major updates** (don't let them accumulate)

### For Security Team

1. **Weekly vulnerability review**
2. **Track MTTP metrics**
3. **Quarterly dependency audit**
4. **Update security playbooks** based on lessons learned
5. **Coordinate with upstream** on critical vulnerabilities

---

## Additional Resources

- **GitHub Dependabot Docs:** https://docs.github.com/en/code-security/dependabot
- **Go Vulnerability Database:** https://vuln.go.dev
- **CVE Database:** https://cve.mitre.org
- **NIST NVD:** https://nvd.nist.gov
- **CVSS Calculator:** https://www.first.org/cvss/calculator/3.1
- **Go Modules Reference:** https://go.dev/ref/mod

**Internal Resources:**
- SECURITY_SCANNING.md - gosec and govulncheck usage
- SECURITY_OPERATOR_GUIDE.md - Security configuration
- SECURITY_ROADMAP.md - Security improvement roadmap

---

## Version History

| Date | Version | Changes |
|------|---------|---------|
| 2025-11-22 | 1.0 | Initial implementation (Phase 3.5) |

---

**Maintained By:** Security Team & Platform Team
**Questions:** security@temporal.io
