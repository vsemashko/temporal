#!/bin/bash
#
# Technical Debt Audit Script
#
# This script analyzes all TODO/FIXME/XXX/HACK comments in the codebase
# and generates a comprehensive report for categorization and tracking.
#
# Usage:
#   ./scripts/audit_technical_debt.sh [output_file]
#
# Reference: PROJECT_IMPLEMENTATION_ROADMAP.md (Sprint 1, Week 5)

set -euo pipefail

# Configuration
OUTPUT_FILE="${1:-TECHNICAL_DEBT_AUDIT.md}"
REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
TEMP_DIR=$(mktemp -d)

echo "🔍 Starting Technical Debt Audit..."
echo "📂 Repository: $REPO_ROOT"
echo "📄 Output: $OUTPUT_FILE"
echo ""

cd "$REPO_ROOT"

# Extract all TODOs with context
echo "Extracting TODOs, FIXMEs, XXXs, and HACKs..."
grep -rn "TODO\|FIXME\|XXX\|HACK" --include="*.go" --include="*.md" --exclude-dir=vendor --exclude-dir=.git . > "$TEMP_DIR/all_debt.txt" 2>/dev/null || true

# Count by category
TODO_COUNT=$(grep -c "TODO" "$TEMP_DIR/all_debt.txt" 2>/dev/null || echo "0")
FIXME_COUNT=$(grep -c "FIXME" "$TEMP_DIR/all_debt.txt" 2>/dev/null || echo "0")
XXX_COUNT=$(grep -c "XXX" "$TEMP_DIR/all_debt.txt" 2>/dev/null || echo "0")
HACK_COUNT=$(grep -c "HACK" "$TEMP_DIR/all_debt.txt" 2>/dev/null || echo "0")
TOTAL_COUNT=$((TODO_COUNT + FIXME_COUNT + XXX_COUNT + HACK_COUNT))

echo "📊 Found $TOTAL_COUNT total technical debt items:"
echo "   - TODO: $TODO_COUNT"
echo "   - FIXME: $FIXME_COUNT"
echo "   - XXX: $XXX_COUNT"
echo "   - HACK: $HACK_COUNT"
echo ""

# Generate markdown report
cat > "$OUTPUT_FILE" << 'HEADER'
# Technical Debt Audit Report

**Generated:** $(date)
**Repository:** Temporal Server
**Branch:** $(git branch --show-current)

---

## Executive Summary

HEADER

cat >> "$OUTPUT_FILE" << EOF
This report provides a comprehensive audit of all technical debt markers (TODO, FIXME, XXX, HACK) found in the codebase.

**Total Technical Debt Items:** $TOTAL_COUNT

| Category | Count | Percentage |
|----------|-------|------------|
| TODO | $TODO_COUNT | $(echo "scale=1; $TODO_COUNT * 100 / $TOTAL_COUNT" | bc 2>/dev/null || echo "N/A")% |
| FIXME | $FIXME_COUNT | $(echo "scale=1; $FIXME_COUNT * 100 / $TOTAL_COUNT" | bc 2>/dev/null || echo "N/A")% |
| XXX | $XXX_COUNT | $(echo "scale=1; $XXX_COUNT * 100 / $TOTAL_COUNT" | bc 2>/dev/null || echo "N/A")% |
| HACK | $HACK_COUNT | $(echo "scale=1; $HACK_COUNT * 100 / $TOTAL_COUNT" | bc 2>/dev/null || echo "N/A")% |

**Categorization Guidelines:**
- 🔴 **Critical** - Blocks features, security risk, data loss potential
- 🟡 **High** - Impacts performance, stability, or maintainability
- 🟢 **Medium** - Nice to have, refactoring, code quality
- 🔵 **Low** - Cosmetic, future enhancement, documentation

**Target Reduction:**
- Critical: 0 (immediate action required)
- High: <20 (resolve in Q1 2026)
- Medium: <50 (ongoing reduction)
- Low: Track but defer

---

## Debt by Component

EOF

# Analyze by directory
echo "Analyzing by component..."
echo "### Top 10 Components by Debt Count" >> "$OUTPUT_FILE"
echo "" >> "$OUTPUT_FILE"

grep "TODO\|FIXME\|XXX\|HACK" "$TEMP_DIR/all_debt.txt" 2>/dev/null | \
  awk -F: '{print $1}' | \
  xargs -I {} dirname {} | \
  sort | uniq -c | sort -rn | head -10 | \
  awk '{print "- `" $2 "`: " $1 " items"}' >> "$OUTPUT_FILE" 2>/dev/null || echo "No data" >> "$OUTPUT_FILE"

echo "" >> "$OUTPUT_FILE"
echo "---" >> "$OUTPUT_FILE"
echo "" >> "$OUTPUT_FILE"

# Detailed listings
echo "## Detailed Listings" >> "$OUTPUT_FILE"
echo "" >> "$OUTPUT_FILE"

# TODO items
echo "### TODO Items ($TODO_COUNT)" >> "$OUTPUT_FILE"
echo "" >> "$OUTPUT_FILE"
echo "Items marked with TODO indicate planned future work or improvements." >> "$OUTPUT_FILE"
echo "" >> "$OUTPUT_FILE"

grep "TODO" "$TEMP_DIR/all_debt.txt" 2>/dev/null | head -50 | while IFS=: read -r file line content; do
  # Clean up the file path
  file_clean=$(echo "$file" | sed 's|^\./||')

  # Extract the TODO comment
  todo_comment=$(echo "$content" | sed 's/.*TODO[: ]*//')

  cat >> "$OUTPUT_FILE" << ITEM

#### \`$file_clean:$line\`
**Comment:** $todo_comment

**Categorization:** [ ] Critical | [ ] High | [ ] Medium | [ ] Low

**Estimated Effort:** [ ] 1-2 hours | [ ] 1 day | [ ] 2-3 days | [ ] 1 week | [ ] >1 week

**GitHub Issue:** #_____

**Notes:**
- Impact:
- Dependencies:
- Assigned to:

---

ITEM
done

# FIXME items
echo "" >> "$OUTPUT_FILE"
echo "### FIXME Items ($FIXME_COUNT)" >> "$OUTPUT_FILE"
echo "" >> "$OUTPUT_FILE"
echo "Items marked with FIXME indicate known bugs or issues requiring fixes." >> "$OUTPUT_FILE"
echo "" >> "$OUTPUT_FILE"

grep "FIXME" "$TEMP_DIR/all_debt.txt" 2>/dev/null | head -30 | while IFS=: read -r file line content; do
  file_clean=$(echo "$file" | sed 's|^\./||')
  fixme_comment=$(echo "$content" | sed 's/.*FIXME[: ]*//')

  cat >> "$OUTPUT_FILE" << ITEM

#### \`$file_clean:$line\`
**Comment:** $fixme_comment

**Categorization:** [ ] Critical | [ ] High | [ ] Medium | [ ] Low

**Estimated Effort:** [ ] 1-2 hours | [ ] 1 day | [ ] 2-3 days | [ ] 1 week | [ ] >1 week

**GitHub Issue:** #_____

**Notes:**
- Impact:
- Root cause:
- Assigned to:

---

ITEM
done

# XXX items
echo "" >> "$OUTPUT_FILE"
echo "### XXX Items ($XXX_COUNT)" >> "$OUTPUT_FILE"
echo "" >> "$OUTPUT_FILE"
echo "Items marked with XXX indicate areas of concern or code that needs attention." >> "$OUTPUT_FILE"
echo "" >> "$OUTPUT_FILE"

grep "XXX" "$TEMP_DIR/all_debt.txt" 2>/dev/null | head -20 | while IFS=: read -r file line content; do
  file_clean=$(echo "$file" | sed 's|^\./||')
  xxx_comment=$(echo "$content" | sed 's/.*XXX[: ]*//')

  cat >> "$OUTPUT_FILE" << ITEM

#### \`$file_clean:$line\`
**Comment:** $xxx_comment

**Categorization:** [ ] Critical | [ ] High | [ ] Medium | [ ] Low

**Estimated Effort:** [ ] 1-2 hours | [ ] 1 day | [ ] 2-3 days | [ ] 1 week | [ ] >1 week

**GitHub Issue:** #_____

---

ITEM
done

# HACK items
echo "" >> "$OUTPUT_FILE"
echo "### HACK Items ($HACK_COUNT)" >> "$OUTPUT_FILE"
echo "" >> "$OUTPUT_FILE"
echo "Items marked with HACK indicate temporary workarounds that should be replaced with proper solutions." >> "$OUTPUT_FILE"
echo "" >> "$OUTPUT_FILE"

grep "HACK" "$TEMP_DIR/all_debt.txt" 2>/dev/null | head -20 | while IFS=: read -r file line content; do
  file_clean=$(echo "$file" | sed 's|^\./||')
  hack_comment=$(echo "$content" | sed 's/.*HACK[: ]*//')

  cat >> "$OUTPUT_FILE" << ITEM

#### \`$file_clean:$line\`
**Comment:** $hack_comment

**Categorization:** [ ] Critical | [ ] High | [ ] Medium | [ ] Low

**Estimated Effort:** [ ] 1-2 hours | [ ] 1 day | [ ] 2-3 days | [ ] 1 week | [ ] >1 week

**GitHub Issue:** #_____

**Notes:**
- Proper solution:
- Blocking factors:

---

ITEM
done

# Priority items
echo "" >> "$OUTPUT_FILE"
echo "## High-Priority Items (Examples)" >> "$OUTPUT_FILE"
echo "" >> "$OUTPUT_FILE"
echo "Based on preliminary analysis, these items appear to be high-priority:" >> "$OUTPUT_FILE"
echo "" >> "$OUTPUT_FILE"

cat >> "$OUTPUT_FILE" << 'PRIORITY'
### 1. Circular Dependency in cluster_metadata_loader
**File:** `temporal/cluster_metadata_loader.go:15`
**Comment:** "TODO: move this to the [cluster] package. It is here temporarily to avoid a circular dependency."
**Category:** 🟡 High
**Effort:** 2-3 days
**Impact:** Cleaner architecture, easier maintenance
**Action:** Refactor to resolve circular dependency

### 2. Replace context.TODO() Calls
**Files:** Multiple locations (elasticsearch/tasks.go, temporal/fx_test.go, etc.)
**Comment:** Using context.TODO() instead of proper context propagation
**Category:** 🟡 High
**Effort:** 1 week (multiple files)
**Impact:** Proper timeout handling, better context cancellation
**Action:** Replace all context.TODO() with proper parent context or timeout context

### 3. Workflow Task Type TODOs
**Files:** api/enums/v1/workflow_task_type.pb.go and related
**Comment:** Various workflow task handling improvements
**Category:** 🟢 Medium
**Effort:** Variable
**Impact:** Improved workflow task handling
**Action:** Review and implement workflow improvements

### 4. Replication TODOs
**Files:** service/history/replication/*.go
**Comment:** Replication task batching, metrics, error handling
**Category:** 🟡 High
**Effort:** 1-2 weeks
**Impact:** Better replication performance and reliability
**Action:** Implement replication improvements

### 5. Test Coverage TODOs
**Files:** Various test files
**Comment:** Missing test cases, edge case testing
**Category:** 🟢 Medium
**Effort:** Ongoing
**Impact:** Better test coverage, fewer bugs
**Action:** Add missing test coverage incrementally

PRIORITY

# Action plan
cat >> "$OUTPUT_FILE" << 'FOOTER'

---

## Action Plan

### Phase 1: Categorization (Week 5)
- [ ] Review this report with engineering team
- [ ] Categorize each item as Critical/High/Medium/Low
- [ ] Estimate effort for each item
- [ ] Identify dependencies and blockers
- [ ] Create GitHub issues for all Critical and High items

### Phase 2: Critical & High Resolution (Week 6)
- [ ] Resolve all Critical items (target: 0)
- [ ] Resolve top 20 High items (target: <20 remaining)
- [ ] Document any blockers or deferred items
- [ ] Update this report with progress

### Phase 3: Ongoing Reduction (Q1 2026)
- [ ] Allocate 20% of sprint capacity to technical debt
- [ ] Target: <100 total items by end of Q1
- [ ] Prevent new debt: Code review for new TODOs
- [ ] Monthly progress tracking

---

## Metrics

**Baseline (Today):**
- Total: $(TOTAL_COUNT)
- Critical: TBD (requires categorization)
- High: TBD
- Medium: TBD
- Low: TBD

**Target (End of Q1 2026):**
- Total: <100
- Critical: 0
- High: <20
- Medium: <50
- Low: <30

**Progress Tracking:**
- Update this report monthly
- Track resolution rate in sprint retrospectives
- Celebrate debt reduction milestones

---

**Next Steps:**
1. Review this report in team meeting
2. Begin categorization using the checklist format above
3. Create GitHub issues using provided template
4. Schedule Sprint 1 (February) for debt reduction
5. Set up automated tracking (GitHub Projects)

**Report Generated:** $(date)
**Script:** scripts/audit_technical_debt.sh
**Reference:** PROJECT_IMPLEMENTATION_ROADMAP.md

FOOTER

# Create summary CSV for easy filtering
echo "Creating CSV summary..."
cat > "${OUTPUT_FILE%.md}.csv" << 'CSV_HEADER'
File,Line,Type,Comment,Category,Effort,Issue
CSV_HEADER

grep "TODO\|FIXME\|XXX\|HACK" "$TEMP_DIR/all_debt.txt" 2>/dev/null | while IFS=: read -r file line content; do
  file_clean=$(echo "$file" | sed 's|^\./||' | sed 's/,/;/g')

  # Detect type
  if echo "$content" | grep -q "TODO"; then
    type="TODO"
  elif echo "$content" | grep -q "FIXME"; then
    type="FIXME"
  elif echo "$content" | grep -q "XXX"; then
    type="XXX"
  elif echo "$content" | grep -q "HACK"; then
    type="HACK"
  else
    type="UNKNOWN"
  fi

  # Extract comment
  comment=$(echo "$content" | sed 's/.*TODO[: ]*//' | sed 's/.*FIXME[: ]*//' | sed 's/.*XXX[: ]*//' | sed 's/.*HACK[: ]*//' | sed 's/,/;/g' | head -c 100)

  echo "$file_clean,$line,$type,\"$comment\",TBD,TBD,TBD" >> "${OUTPUT_FILE%.md}.csv"
done

# Cleanup
rm -rf "$TEMP_DIR"

echo ""
echo "✅ Technical Debt Audit Complete!"
echo ""
echo "📄 Reports generated:"
echo "   - Markdown: $OUTPUT_FILE"
echo "   - CSV: ${OUTPUT_FILE%.md}.csv"
echo ""
echo "📊 Summary:"
echo "   - Total items: $TOTAL_COUNT"
echo "   - Target reduction: $(echo "scale=0; $TOTAL_COUNT - 100" | bc 2>/dev/null || echo "N/A") items (to reach <100)"
echo "   - Percentage reduction needed: $(echo "scale=1; ($TOTAL_COUNT - 100) * 100 / $TOTAL_COUNT" | bc 2>/dev/null || echo "N/A")%"
echo ""
echo "🎯 Next Steps:"
echo "   1. Review $OUTPUT_FILE with team"
echo "   2. Categorize items as Critical/High/Medium/Low"
echo "   3. Create GitHub issues for Critical/High items"
echo "   4. Schedule Sprint 1 for debt reduction (February)"
echo ""
echo "Reference: PROJECT_IMPLEMENTATION_ROADMAP.md (Sprint 1, Week 5-6)"
