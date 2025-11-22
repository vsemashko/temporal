#!/bin/bash
#
# Deployment Window Calculator
#
# This script helps you find optimal deployment windows based on your constraints.
# It considers blackout periods, team availability, and risk windows.
#
# Usage:
#   ./scripts/calculate_deployment_windows.sh
#
# Or with options:
#   ./scripts/calculate_deployment_windows.sh --start-date 2025-12-01 --scenario standard
#
# Reference: TIMELINE_SCENARIOS.md

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Default values
START_DATE=""
SCENARIO="standard"
OUTPUT_FILE="DEPLOYMENT_WINDOWS_SCHEDULE.md"

# Parse command line arguments
while [[ $# -gt 0 ]]; do
  case $1 in
    --start-date)
      START_DATE="$2"
      shift 2
      ;;
    --scenario)
      SCENARIO="$2"
      shift 2
      ;;
    --output)
      OUTPUT_FILE="$2"
      shift 2
      ;;
    --help)
      echo "Usage: $0 [OPTIONS]"
      echo ""
      echo "Options:"
      echo "  --start-date YYYY-MM-DD    Start date for Week 1 (default: next Monday)"
      echo "  --scenario NAME            Scenario: aggressive|standard|conservative|minimal (default: standard)"
      echo "  --output FILE              Output file (default: DEPLOYMENT_WINDOWS_SCHEDULE.md)"
      echo "  --help                     Show this help message"
      exit 0
      ;;
    *)
      echo "Unknown option: $1"
      echo "Use --help for usage information"
      exit 1
      ;;
  esac
done

echo -e "${BLUE}🗓️  Deployment Window Calculator${NC}"
echo "=================================="
echo ""

# Determine start date (default: next Monday)
if [ -z "$START_DATE" ]; then
  # Calculate next Monday
  case "$(date +%u)" in
    1) START_DATE=$(date -d "+7 days" +%Y-%m-%d) ;;  # If Monday, next week
    *) START_DATE=$(date -d "next Monday" +%Y-%m-%d) ;;
  esac
  echo -e "${YELLOW}No start date specified. Using next Monday: $START_DATE${NC}"
else
  echo -e "${GREEN}Start date: $START_DATE${NC}"
fi

# Validate start date
if ! date -d "$START_DATE" >/dev/null 2>&1; then
  echo -e "${RED}Error: Invalid start date format. Use YYYY-MM-DD${NC}"
  exit 1
fi

# Determine timeline based on scenario
case "$SCENARIO" in
  aggressive)
    TOTAL_WEEKS=4
    TIMELINE_DAYS=30
    echo -e "${GREEN}Scenario: Aggressive (30 days, 4 weeks)${NC}"
    ;;
  standard)
    TOTAL_WEEKS=8
    TIMELINE_DAYS=60
    echo -e "${GREEN}Scenario: Standard (60 days, 8 weeks)${NC}"
    ;;
  conservative)
    TOTAL_WEEKS=12
    TIMELINE_DAYS=90
    echo -e "${GREEN}Scenario: Conservative (90 days, 12 weeks)${NC}"
    ;;
  minimal)
    TOTAL_WEEKS=16
    TIMELINE_DAYS=120
    echo -e "${GREEN}Scenario: Minimal Team (120 days, 16 weeks)${NC}"
    ;;
  *)
    echo -e "${RED}Error: Unknown scenario '$SCENARIO'${NC}"
    echo "Valid scenarios: aggressive, standard, conservative, minimal"
    exit 1
    ;;
esac

echo ""
echo "⚙️  Calculating deployment windows..."
echo ""

# Interactive blackout period collection
echo -e "${YELLOW}Do you have any blackout periods (no deployments allowed)?${NC}"
echo "Examples: holiday freeze, major releases, planned outages"
read -p "Enter blackout periods? (y/n): " HAS_BLACKOUTS

BLACKOUT_PERIODS=()
if [[ "$HAS_BLACKOUTS" =~ ^[Yy]$ ]]; then
  echo ""
  echo "Enter blackout periods (press Enter with empty input to finish):"
  while true; do
    read -p "  Start date (YYYY-MM-DD) or Enter to finish: " blackout_start
    if [ -z "$blackout_start" ]; then
      break
    fi
    read -p "  End date (YYYY-MM-DD): " blackout_end
    read -p "  Reason: " blackout_reason
    BLACKOUT_PERIODS+=("$blackout_start|$blackout_end|$blackout_reason")
    echo -e "  ${GREEN}✓ Added blackout: $blackout_start to $blackout_end${NC}"
  done
fi

# Interactive preferences
echo ""
echo -e "${YELLOW}Deployment Preferences:${NC}"

echo "Preferred deployment day of week?"
echo "  1) Tuesday (recommended - allows Monday for prep, rest of week for monitoring)"
echo "  2) Wednesday (safe middle of week)"
echo "  3) Thursday (later in week, less time to fix before weekend)"
echo "  4) Monday (risky - weekend issues may linger)"
echo "  5) Friday (not recommended - no weekend coverage)"
read -p "Select (1-5): " DAY_CHOICE

case "$DAY_CHOICE" in
  1) DEPLOY_DAY="Tuesday" ;;
  2) DEPLOY_DAY="Wednesday" ;;
  3) DEPLOY_DAY="Thursday" ;;
  4) DEPLOY_DAY="Monday" ;;
  5) DEPLOY_DAY="Friday" ;;
  *) DEPLOY_DAY="Tuesday" ;;
esac

echo "Preferred deployment time (timezone)?"
read -p "Enter timezone (e.g., PST, EST, UTC): " TIMEZONE
read -p "Enter time (e.g., 10:00 AM): " DEPLOY_TIME

echo "Weekend deployments allowed?"
read -p "(y/n): " WEEKEND_OK

# Calculate key deployment windows
echo ""
echo "📅 Calculating key deployment dates..."
echo ""

# Function to calculate date N days from start
calc_date() {
  local days=$1
  date -d "$START_DATE + $days days" +%Y-%m-%d
}

# Function to get day of week
get_day() {
  local date=$1
  date -d "$date" +%A
}

# Function to check if date is in blackout
is_blackout() {
  local check_date=$1
  for period in "${BLACKOUT_PERIODS[@]}"; do
    IFS='|' read -r start end reason <<< "$period"
    if [[ "$check_date" > "$start" || "$check_date" == "$start" ]] && \
       [[ "$check_date" < "$end" || "$check_date" == "$end" ]]; then
      return 0  # In blackout
    fi
  done
  return 1  # Not in blackout
}

# Function to find next valid deployment day
find_next_deploy_day() {
  local start=$1
  local target_day=$2
  local current=$start

  for i in {0..14}; do  # Search up to 2 weeks ahead
    current=$(date -d "$start + $i days" +%Y-%m-%d)
    current_day=$(get_day "$current")

    # Check if it matches target day
    if [ "$current_day" == "$target_day" ]; then
      # Check if in blackout
      if ! is_blackout "$current"; then
        # Check if weekend and weekend deployments not allowed
        if [[ "$current_day" == "Saturday" || "$current_day" == "Sunday" ]]; then
          if [[ ! "$WEEKEND_OK" =~ ^[Yy]$ ]]; then
            continue
          fi
        fi
        echo "$current"
        return
      fi
    fi
  done

  echo "$start"  # Fallback to start date
}

# Generate schedule based on scenario
case "$SCENARIO" in
  aggressive)
    # 4-week aggressive timeline
    WEEK1_START=$(calc_date 0)
    WEEK2_START=$(calc_date 7)
    WEEK2_DEPLOY=$(find_next_deploy_day "$WEEK2_START" "$DEPLOY_DAY")
    WEEK3_START=$(calc_date 14)
    WEEK3_DEPLOY=$(find_next_deploy_day "$WEEK3_START" "$DEPLOY_DAY")
    WEEK4_START=$(calc_date 21)
    WEEK4_DAY1=$(find_next_deploy_day "$WEEK4_START" "$DEPLOY_DAY")
    WEEK4_DAY2=$(calc_date 22)
    WEEK4_DAY3=$(calc_date 23)
    WEEK4_DAY4=$(calc_date 24)
    WEEK4_DAY5=$(calc_date 25)
    ;;

  standard)
    # 8-week standard timeline
    WEEK1_START=$(calc_date 0)
    WEEK2_START=$(calc_date 7)
    WEEK3_START=$(calc_date 14)
    WEEK3_DEPLOY=$(find_next_deploy_day "$WEEK3_START" "$DEPLOY_DAY")
    WEEK5_START=$(calc_date 28)
    WEEK5_DEPLOY=$(find_next_deploy_day "$WEEK5_START" "$DEPLOY_DAY")
    WEEK7_START=$(calc_date 42)
    WEEK7_DAY1=$(find_next_deploy_day "$WEEK7_START" "$DEPLOY_DAY")
    WEEK7_DAY3=$(calc_date 44)
    WEEK7_DAY5=$(calc_date 46)
    WEEK8_START=$(calc_date 49)
    WEEK8_DAY1=$(find_next_deploy_day "$WEEK8_START" "$DEPLOY_DAY")
    WEEK8_DAY3=$(calc_date 51)
    ;;

  conservative)
    # 12-week conservative timeline
    MONTH1_START=$(calc_date 0)
    MONTH2_START=$(calc_date 30)
    MONTH2_DEPLOY=$(find_next_deploy_day "$MONTH2_START" "$DEPLOY_DAY")
    MONTH3_START=$(calc_date 60)
    MONTH3_DEPLOY=$(find_next_deploy_day "$MONTH3_START" "$DEPLOY_DAY")
    MONTH3_WEEK3=$(calc_date 74)
    MONTH3_WEEK3_DEPLOY=$(find_next_deploy_day "$MONTH3_WEEK3" "$DEPLOY_DAY")
    ;;

  minimal)
    # 16-week minimal timeline
    MONTH1_START=$(calc_date 0)
    MONTH3_START=$(calc_date 60)
    MONTH3_DEPLOY=$(find_next_deploy_day "$MONTH3_START" "$DEPLOY_DAY")
    MONTH4_START=$(calc_date 90)
    MONTH4_DEPLOY=$(find_next_deploy_day "$MONTH4_START" "$DEPLOY_DAY")
    ;;
esac

# Generate markdown output
cat > "$OUTPUT_FILE" << EOF
# Deployment Windows Schedule

**Generated:** $(date)
**Scenario:** $SCENARIO
**Start Date:** $START_DATE ($(get_day "$START_DATE"))
**Total Duration:** $TIMELINE_DAYS days ($TOTAL_WEEKS weeks)
**Deployment Day:** $DEPLOY_DAY at $DEPLOY_TIME $TIMEZONE

---

## 📅 Key Deployment Windows

EOF

# Add blackout periods if any
if [ ${#BLACKOUT_PERIODS[@]} -gt 0 ]; then
  cat >> "$OUTPUT_FILE" << EOF
### ⚠️ Blackout Periods (No Deployments)

EOF
  for period in "${BLACKOUT_PERIODS[@]}"; do
    IFS='|' read -r start end reason <<< "$period"
    cat >> "$OUTPUT_FILE" << EOF
- **$start to $end:** $reason
EOF
  done
  cat >> "$OUTPUT_FILE" << EOF

---

EOF
fi

# Add scenario-specific schedule
case "$SCENARIO" in
  aggressive)
    cat >> "$OUTPUT_FILE" << EOF
## Aggressive Timeline (30 Days / 4 Weeks)

### Week 1: Preparation
**Dates:** $WEEK1_START ($(get_day "$WEEK1_START")) to $(calc_date 6) ($(get_day "$(calc_date 6)"))

**Activities:**
- [ ] Day 1-2: Team kickoff, certificate generation
- [ ] Day 3-4: Staging environment prep
- [ ] Day 5-7: Prometheus alerts, Grafana dashboards, team training

**Team:** 8-10 people actively working

---

### Week 2: Staging Deployment
**Dates:** $WEEK2_START ($(get_day "$WEEK2_START")) to $(calc_date 13) ($(get_day "$(calc_date 13)"))

**Deployment Window:** $WEEK2_DEPLOY ($(get_day "$WEEK2_DEPLOY")) at $DEPLOY_TIME $TIMEZONE

**Activities:**
- [ ] $WEEK2_DEPLOY: Deploy to staging
- [ ] $(calc_date 8)-$(calc_date 9): Security testing
- [ ] $(calc_date 10)-$(calc_date 11): Performance testing
- [ ] $(calc_date 12)-$(calc_date 13): Load testing, validation

**Team:** 6-8 people actively working

---

### Week 3: Production Canary (10%)
**Dates:** $WEEK3_START ($(get_day "$WEEK3_START")) to $(calc_date 20) ($(get_day "$(calc_date 20)"))

**Deployment Window:** $WEEK3_DEPLOY ($(get_day "$WEEK3_DEPLOY")) at $DEPLOY_TIME $TIMEZONE

**Activities:**
- [ ] $WEEK3_DEPLOY: Deploy 10% canary
- [ ] $(calc_date 15)-$(calc_date 16): Hourly monitoring
- [ ] $(calc_date 17)-$(calc_date 20): 72-hour soak test

**Team:** 4-6 people monitoring, rest on standby

---

### Week 4: Full Production Rollout
**Dates:** $WEEK4_START ($(get_day "$WEEK4_START")) to $(calc_date 29) ($(get_day "$(calc_date 29)"))

**Deployment Windows:**
- [ ] $WEEK4_DAY1 ($(get_day "$WEEK4_DAY1")): 20% rollout at $DEPLOY_TIME $TIMEZONE
- [ ] $WEEK4_DAY2 ($(get_day "$WEEK4_DAY2")): 40% rollout at $DEPLOY_TIME $TIMEZONE
- [ ] $WEEK4_DAY3 ($(get_day "$WEEK4_DAY3")): 60% rollout at $DEPLOY_TIME $TIMEZONE
- [ ] $WEEK4_DAY4 ($(get_day "$WEEK4_DAY4")): 80% rollout at $DEPLOY_TIME $TIMEZONE
- [ ] $WEEK4_DAY5 ($(get_day "$WEEK4_DAY5")): 100% rollout at $DEPLOY_TIME $TIMEZONE

**Post-Deployment:**
- [ ] $(calc_date 26)-$(calc_date 29): Stabilization, first secrets rotation

**Team:** 4-6 people for rollout

EOF
    ;;

  standard)
    cat >> "$OUTPUT_FILE" << EOF
## Standard Timeline (60 Days / 8 Weeks)

### Weeks 1-2: Extended Preparation
**Dates:** $WEEK1_START to $(calc_date 13)

**Week 1 Activities:**
- [ ] Day 1-3: Kickoff, planning, initial setup
- [ ] Day 4-7: Certificate generation, secrets management

**Week 2 Activities:**
- [ ] Day 8-10: Staging environment configuration
- [ ] Day 11-14: Prometheus alerts, Grafana dashboards

---

### Weeks 3-4: Staging Deployment & Testing
**Dates:** $WEEK3_START to $(calc_date 27)

**Deployment Window:** $WEEK3_DEPLOY ($(get_day "$WEEK3_DEPLOY")) at $DEPLOY_TIME $TIMEZONE

**Activities:**
- [ ] $WEEK3_DEPLOY: Deploy to staging
- [ ] Week 3: Security testing, performance validation
- [ ] Week 4: Load testing, rollback testing, final validation

---

### Weeks 5-6: Production Canary
**Dates:** $WEEK5_START to $(calc_date 41)

**Deployment Window:** $WEEK5_DEPLOY ($(get_day "$WEEK5_DEPLOY")) at $DEPLOY_TIME $TIMEZONE

**Activities:**
- [ ] $WEEK5_DEPLOY: Deploy 10% canary
- [ ] Days $(calc_date 29)-$(calc_date 34): Extended monitoring (5 days)
- [ ] Week 6: Full week soak test, go/no-go decision

---

### Weeks 7-8: Production Rollout
**Dates:** $WEEK7_START to $(calc_date 59)

**Deployment Windows:**
- [ ] $WEEK7_DAY1 ($(get_day "$WEEK7_DAY1")): 20% rollout at $DEPLOY_TIME $TIMEZONE
- [ ] $WEEK7_DAY3 ($(get_day "$WEEK7_DAY3")): 40% rollout at $DEPLOY_TIME $TIMEZONE
- [ ] $WEEK7_DAY5 ($(get_day "$WEEK7_DAY5")): 60% rollout at $DEPLOY_TIME $TIMEZONE
- [ ] $WEEK8_DAY1 ($(get_day "$WEEK8_DAY1")): 80% rollout at $DEPLOY_TIME $TIMEZONE
- [ ] $WEEK8_DAY3 ($(get_day "$WEEK8_DAY3")): 100% rollout at $DEPLOY_TIME $TIMEZONE

**Post-Deployment:**
- [ ] Days $(calc_date 52)-$(calc_date 59): Stabilization, secrets rotation

EOF
    ;;

  conservative)
    cat >> "$OUTPUT_FILE" << EOF
## Conservative Timeline (90 Days / 12 Weeks)

### Month 1: Preparation & Planning
**Dates:** $MONTH1_START to $(calc_date 29)

**Activities:**
- Weeks 1-2: Planning and resource allocation
- Weeks 3-4: Infrastructure preparation, monitoring setup

---

### Month 2: Staging Deployment & Testing
**Dates:** $MONTH2_START to $(calc_date 59)

**Deployment Window:** $MONTH2_DEPLOY ($(get_day "$MONTH2_DEPLOY")) at $DEPLOY_TIME $TIMEZONE

**Activities:**
- [ ] $MONTH2_DEPLOY: Deploy to staging
- Weeks 5-6: Initial and security testing
- Weeks 7-8: Extended validation, issue remediation

---

### Month 3: Production Rollout
**Dates:** $MONTH3_START to $(calc_date 89)

**Deployment Windows:**
- [ ] $MONTH3_DEPLOY ($(get_day "$MONTH3_DEPLOY")): 5% canary at $DEPLOY_TIME $TIMEZONE
- [ ] $(calc_date 67) ($(get_day "$(calc_date 67)")): Review after 1 week
- [ ] $MONTH3_WEEK3_DEPLOY ($(get_day "$MONTH3_WEEK3_DEPLOY")): Begin gradual rollout
  - 10% → 25% → 50% → 75% → 100% over 2 weeks

**Post-Deployment:**
- Secrets rotation deferred to Month 4

EOF
    ;;

  minimal)
    cat >> "$OUTPUT_FILE" << EOF
## Minimal Team Timeline (120 Days / 16 Weeks)

### Months 1-2: Incremental Preparation
**Dates:** $MONTH1_START to $(calc_date 59)

**Activities:**
- Weeks 1-4: Planning and prioritization (2-4 hours/week)
- Weeks 5-8: Infrastructure work in small chunks

---

### Months 3-4: Staged Rollout
**Dates:** $MONTH3_START to $(calc_date 119)

**Deployment Windows:**
- [ ] $MONTH3_DEPLOY ($(get_day "$MONTH3_DEPLOY")): Deploy to staging
- [ ] $MONTH4_DEPLOY ($(get_day "$MONTH4_DEPLOY")): 5% canary at $DEPLOY_TIME $TIMEZONE
- [ ] $(calc_date 104) ($(get_day "$(calc_date 104)")): 25% rollout
- [ ] $(calc_date 111) ($(get_day "$(calc_date 111)")): 50% rollout
- [ ] $(calc_date 118) ($(get_day "$(calc_date 118)")): 100% rollout

**Note:** Very gradual pace for minimal team

EOF
    ;;
esac

# Add calendar export section
cat >> "$OUTPUT_FILE" << EOF

---

## 📆 Add to Calendar

### Key Dates to Block

Copy these dates into your team calendar:

EOF

# Add key dates based on scenario
case "$SCENARIO" in
  aggressive)
    cat >> "$OUTPUT_FILE" << EOF
1. **Week 1 Kickoff:** $WEEK1_START at 10:00 AM $TIMEZONE (2 hours)
2. **Staging Deployment:** $WEEK2_DEPLOY at $DEPLOY_TIME $TIMEZONE (4 hours, all-hands)
3. **Canary Deployment:** $WEEK3_DEPLOY at $DEPLOY_TIME $TIMEZONE (4 hours, all-hands)
4. **Production Rollout:** $WEEK4_DAY1 through $WEEK4_DAY5 (daily at $DEPLOY_TIME $TIMEZONE)
5. **Secrets Rotation:** $(calc_date 27) at $DEPLOY_TIME $TIMEZONE (2 hours)

EOF
    ;;
  standard)
    cat >> "$OUTPUT_FILE" << EOF
1. **Week 1 Kickoff:** $WEEK1_START at 10:00 AM $TIMEZONE (2 hours)
2. **Staging Deployment:** $WEEK3_DEPLOY at $DEPLOY_TIME $TIMEZONE (4 hours)
3. **Canary Deployment:** $WEEK5_DEPLOY at $DEPLOY_TIME $TIMEZONE (4 hours)
4. **Production Rollout Start:** $WEEK7_DAY1 at $DEPLOY_TIME $TIMEZONE
5. **Production Rollout Complete:** $WEEK8_DAY3 at $DEPLOY_TIME $TIMEZONE
6. **Secrets Rotation:** $(calc_date 55) at $DEPLOY_TIME $TIMEZONE (2 hours)

EOF
    ;;
esac

# Add iCal format events
cat >> "$OUTPUT_FILE" << EOF

---

## 🔔 Reminders & Notifications

### Set Up Alerts

**1 Week Before Staging Deployment:**
- [ ] Review DEPLOYMENT_VERIFICATION_CHECKLIST.md
- [ ] Verify all certificates generated and valid
- [ ] Confirm team availability
- [ ] Book war room/video call

**1 Day Before Each Deployment:**
- [ ] Run pre-deployment health check
- [ ] Verify rollback procedure ready
- [ ] Brief on-call team
- [ ] Post deployment announcement (if needed)

**During Deployment:**
- [ ] Join war room
- [ ] Monitor metrics dashboard
- [ ] Communicate progress every hour
- [ ] Document any issues

---

## ✅ Next Steps

1. [ ] Review this schedule with your team
2. [ ] Add key dates to team calendar
3. [ ] Reserve deployment windows officially
4. [ ] Send calendar invites to all participants
5. [ ] Update JIRA/project tracker with deployment dates
6. [ ] Communicate schedule to stakeholders

---

**Generated by:** scripts/calculate_deployment_windows.sh
**Scenario:** $SCENARIO
**Reference:** TIMELINE_SCENARIOS.md, PROJECT_IMPLEMENTATION_ROADMAP.md
**Customize:** Edit PLANNING_QUESTIONNAIRE.md and re-run this script
EOF

echo ""
echo -e "${GREEN}✅ Deployment schedule generated!${NC}"
echo ""
echo "📄 Output file: $OUTPUT_FILE"
echo ""
echo "Key deployment windows:"

case "$SCENARIO" in
  aggressive)
    echo "  - Staging: $WEEK2_DEPLOY ($(get_day "$WEEK2_DEPLOY"))"
    echo "  - Canary: $WEEK3_DEPLOY ($(get_day "$WEEK3_DEPLOY"))"
    echo "  - Rollout: $WEEK4_DAY1 through $WEEK4_DAY5"
    ;;
  standard)
    echo "  - Staging: $WEEK3_DEPLOY ($(get_day "$WEEK3_DEPLOY"))"
    echo "  - Canary: $WEEK5_DEPLOY ($(get_day "$WEEK5_DEPLOY"))"
    echo "  - Rollout: $WEEK7_DAY1 through $WEEK8_DAY3"
    ;;
  conservative)
    echo "  - Staging: $MONTH2_DEPLOY ($(get_day "$MONTH2_DEPLOY"))"
    echo "  - Canary: $MONTH3_DEPLOY ($(get_day "$MONTH3_DEPLOY"))"
    echo "  - Rollout: $MONTH3_WEEK3_DEPLOY onwards"
    ;;
  minimal)
    echo "  - Staging: $MONTH3_DEPLOY ($(get_day "$MONTH3_DEPLOY"))"
    echo "  - Canary: $MONTH4_DEPLOY ($(get_day "$MONTH4_DEPLOY"))"
    echo "  - Rollout: Very gradual over 4 weeks"
    ;;
esac

echo ""
echo "🎯 Next steps:"
echo "  1. Review $OUTPUT_FILE"
echo "  2. Add dates to your team calendar"
echo "  3. Reserve deployment windows officially"
echo "  4. Begin Week 1 preparation"
echo ""
echo "Reference: PROJECT_IMPLEMENTATION_ROADMAP.md, TIMELINE_SCENARIOS.md"
