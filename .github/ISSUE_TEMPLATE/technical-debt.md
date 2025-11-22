---
name: Technical Debt
about: Track technical debt items for resolution
title: '[Tech Debt] '
labels: ['technical-debt', 'maintenance']
assignees: ''
---

## Technical Debt Information

**Location:** `file/path.go:line`
**Type:** [ ] TODO | [ ] FIXME | [ ] XXX | [ ] HACK
**Category:** [ ] 🔴 Critical | [ ] 🟡 High | [ ] 🟢 Medium | [ ] 🔵 Low
**Component:** [e.g., service/matching, service/history, common/persistence]

---

## Current State

**TODO Comment:**
```go
// TODO: [Original comment from code]
```

**Context:**
[Why was this TODO added? What problem does it address?]

**Impact:**
[What is the impact of NOT addressing this?]
- [ ] Blocks new features
- [ ] Security risk
- [ ] Performance impact
- [ ] Maintainability issue
- [ ] Code quality
- [ ] Documentation

---

## Proposed Solution

**Approach:**
[How should this be resolved?]

**Estimated Effort:**
[ ] 1-2 hours | [ ] 1 day | [ ] 2-3 days | [ ] 1 week | [ ] >1 week

**Prerequisites:**
- [ ] Prerequisite 1
- [ ] Prerequisite 2

---

## Implementation Plan

### Tasks
- [ ] Task 1: [Description]
- [ ] Task 2: [Description]
- [ ] Task 3: [Description]

### Testing
- [ ] Unit tests added/updated
- [ ] Integration tests added/updated
- [ ] Manual testing completed

### Documentation
- [ ] Code comments updated
- [ ] User-facing docs updated (if applicable)
- [ ] Architecture docs updated (if applicable)

---

## Priority Justification

**Why this priority?**
[Explain the categorization: Critical/High/Medium/Low]

**Dependencies:**
[List any blockers or dependencies]
- Blocks: #___, #___
- Blocked by: #___, #___
- Related to: #___, #___

---

## Acceptance Criteria

**Definition of Done:**
- [ ] TODO comment removed from code
- [ ] Solution implemented and tested
- [ ] PR merged to main
- [ ] Documentation updated
- [ ] No regressions introduced

**Verification:**
[How will we verify this is properly resolved?]

---

## Sprint Planning

**Target Sprint:** [Sprint name/number]
**Assigned to:** @username
**Story Points:** [1/2/3/5/8/13]

**Related Epic:** #___
**Milestone:** [Q1 2026 Tech Debt Reduction]

---

## Additional Context

**References:**
- TECHNICAL_DEBT_AUDIT.md: [Line number]
- Related PRs: #___
- Related discussions: [Links]

**Notes:**
[Any additional context or considerations]
