# Planning Questionnaire & Customization Worksheet

**Purpose:** Customize the implementation roadmap to your team's specific constraints and requirements
**Date:** 2025-11-22
**Instructions:** Fill out this questionnaire, then run `./scripts/generate_custom_roadmap.sh` to create your personalized plan

---

## Section 1: Team Availability & Capacity

### 1.1 Team Size & Composition

**Security Team:**
- Available engineers: _____ (FTE count)
- Names/roles: _________________
- Current workload: ☐ Light (0-25%) | ☐ Medium (25-50%) | ☐ Heavy (50-75%) | ☐ Full (75-100%)
- Available for roadmap work: _____ hours/week per person

**DevOps/SRE Team:**
- Available engineers: _____ (FTE count)
- Names/roles: _________________
- Current workload: ☐ Light | ☐ Medium | ☐ Heavy | ☐ Full
- Available for roadmap work: _____ hours/week per person

**Platform/Backend Engineering Team:**
- Available engineers: _____ (FTE count)
- Names/roles: _________________
- Current workload: ☐ Light | ☐ Medium | ☐ Heavy | ☐ Full
- Available for roadmap work: _____ hours/week per person

**Management/Leadership:**
- Engineering Manager: _________________
- Security Lead: _________________
- VP Engineering: _________________
- Available for reviews/approvals: ☐ Daily | ☐ Weekly | ☐ Bi-weekly

---

### 1.2 Timeline Constraints

**Overall Timeline Preference:**
☐ Aggressive (30 days as planned)
☐ Standard (45-60 days)
☐ Conservative (90+ days)
☐ Custom: _____ days

**Specific Constraints:**
- [ ] Holiday freeze: __________ to __________
- [ ] Major release planned: __________ (date)
- [ ] Team member vacation: __________ (dates)
- [ ] External audit: __________ (date)
- [ ] Budget year end: __________
- [ ] Other: _________________

**Preferred Start Date:**
- Week 1 starts on: __________ (e.g., 2025-12-01)

**Blackout Windows (no deployments):**
- Blackout 1: __________ to __________
- Blackout 2: __________ to __________
- Blackout 3: __________ to __________

---

### 1.3 Resource Allocation

**How much team capacity can you dedicate to this roadmap?**

Week 1-4 (Security Deployment):
- Security team: ☐ 100% | ☐ 75% | ☐ 50% | ☐ 25% | ☐ <25%
- DevOps/SRE: ☐ 100% | ☐ 75% | ☐ 50% | ☐ 25% | ☐ <25%
- Platform: ☐ 100% | ☐ 75% | ☐ 50% | ☐ 25% | ☐ <25%

Q1 2026 (Tech Debt + Observability):
- Engineering team: ☐ 100% | ☐ 75% | ☐ 50% | ☐ 25% | ☐ 20% | ☐ <20%

Q2 2026 (Performance + DX):
- Engineering team: ☐ 100% | ☐ 75% | ☐ 50% | ☐ 25% | ☐ 20% | ☐ <20%

**Other Concurrent Projects:**
- Project 1: _________________ (% of team capacity: ___%)
- Project 2: _________________ (% of team capacity: ___%)
- Project 3: _________________ (% of team capacity: ___%)

---

## Section 2: Deployment Windows & Scheduling

### 2.1 Environment Availability

**Staging Environment:**
- Available for deployment: ☐ 24/7 | ☐ Business hours only | ☐ Specific windows
- Specific windows: _________________
- Requires approval: ☐ Yes (from: _______) | ☐ No
- Shared with other teams: ☐ Yes | ☐ No
- If shared, booking required: ☐ Yes (lead time: ___ days) | ☐ No

**Production Environment:**
- Deployment windows allowed:
  - ☐ Monday-Friday, business hours (9am-5pm)
  - ☐ Monday-Friday, after hours (5pm-9am)
  - ☐ Weekends only
  - ☐ Specific days: _________________
  - ☐ 24/7 (anytime)
- Change freeze windows:
  - From: __________ To: __________
  - From: __________ To: __________
- Approval required: ☐ Yes (from: _______) | ☐ No
- Notice period required: ___ days/hours
- Change Advisory Board (CAB) schedule: ☐ Weekly | ☐ Bi-weekly | ☐ Monthly | ☐ N/A

### 2.2 Preferred Deployment Schedule

**Week 2: Staging Deployment**
- Preferred day: ☐ Mon | ☐ Tue | ☐ Wed | ☐ Thu | ☐ Fri | ☐ Sat | ☐ Sun
- Preferred time: __________ (timezone: ______)
- Duration: ___ hours
- Team availability: ☐ Full team | ☐ Skeleton crew | ☐ Remote only

**Week 3: Production Canary (10%)**
- Preferred day: ☐ Mon | ☐ Tue | ☐ Wed | ☐ Thu | ☐ Fri | ☐ Sat | ☐ Sun
- Preferred time: __________ (timezone: ______)
- Rollback window: ___ hours (how long to wait before declaring success)
- On-call coverage: ☐ 24/7 | ☐ Business hours | ☐ Best effort

**Week 4: Production Full Rollout**
- Rollout pace: ☐ Fast (1 day: 20%→100%) | ☐ Standard (5 days: 20%/day) | ☐ Slow (10 days: 10%/day)
- Preferred start day: ☐ Mon | ☐ Tue | ☐ Wed | ☐ Thu | ☐ Fri
- Daily increment time: __________ (timezone: ______)
- Weekend deployments: ☐ OK | ☐ Avoid if possible | ☐ Never

### 2.3 On-Call & Support

**On-Call Rotation:**
- Current rotation: ☐ 24/7 | ☐ Business hours | ☐ Best effort | ☐ None
- During deployment: ☐ Dedicated team | ☐ Normal rotation | ☐ Extended coverage
- Escalation path:
  - L1: _________________
  - L2: _________________
  - L3: _________________
  - Executive: _________________

**War Room Setup:**
- During deployment: ☐ Physical room | ☐ Zoom/video | ☐ Slack channel | ☐ Not needed
- Participants: _________________
- Communication channel: _________________

---

## Section 3: Environment-Specific Requirements

### 3.1 Staging Environment

**Configuration:**
- Database: ☐ PostgreSQL | ☐ MySQL | ☐ Cassandra | ☐ SQLite | ☐ Other: _____
- Database version: _____
- Cloud provider: ☐ AWS | ☐ GCP | ☐ Azure | ☐ On-prem | ☐ Other: _____
- Region: _____
- Kubernetes version: _____ (or ☐ N/A)
- Number of nodes: _____
- Resources per service:
  - Frontend: ___ CPU, ___ GB RAM
  - History: ___ CPU, ___ GB RAM
  - Matching: ___ CPU, ___ GB RAM
  - Worker: ___ CPU, ___ GB RAM

**Monitoring:**
- Prometheus: ☐ Already deployed | ☐ Need to deploy | ☐ Different tool: _____
- Grafana: ☐ Already deployed | ☐ Need to deploy | ☐ Different tool: _____
- Alerting: ☐ PagerDuty | ☐ Opsgenie | ☐ Slack | ☐ Other: _____

**TLS Certificates:**
- Certificate authority: ☐ Internal CA | ☐ Let's Encrypt | ☐ Commercial CA: _____ | ☐ Self-signed (dev only)
- Certificate management: ☐ Manual | ☐ cert-manager | ☐ AWS ACM | ☐ Other: _____
- Expiration: ☐ 90 days | ☐ 1 year | ☐ Other: _____

### 3.2 Production Environment

**Configuration:**
- Database: ☐ PostgreSQL | ☐ MySQL | ☐ Cassandra | ☐ SQLite | ☐ Other: _____
- Database version: _____
- High availability: ☐ Single region | ☐ Multi-AZ | ☐ Multi-region
- Replica count: _____
- Cloud provider: ☐ AWS | ☐ GCP | ☐ Azure | ☐ On-prem | ☐ Other: _____
- Regions: _____
- Kubernetes version: _____ (or ☐ N/A)
- Number of nodes: _____
- Auto-scaling: ☐ Enabled | ☐ Disabled

**Scale:**
- Current request volume: _____ req/sec (average)
- Peak request volume: _____ req/sec
- Number of workflows: _____
- Number of namespaces: _____
- Number of workers: _____

**Compliance Requirements:**
- ☐ PCI-DSS
- ☐ HIPAA
- ☐ SOC 2
- ☐ GDPR
- ☐ ISO 27001
- ☐ Other: _____

---

## Section 4: Specific Feature Requirements

### 4.1 Security Features (Phase 1-3)

**Required Features (must have):**
- [ ] TLS 1.3 with FIPS cipher suites
- [ ] mTLS client authentication
- [ ] JWT-based authorization
- [ ] Authentication rate limiting
- [ ] Audit logging
- [ ] Context timeout enforcement
- [ ] Certificate pinning
- [ ] Configuration sanitization

**Optional Features (nice to have):**
- [ ] Certificate pinning for remote clusters (if multi-region)
- [ ] Advanced security metrics
- [ ] SIEM integration

**Not Needed (skip for now):**
- [ ] _________________
- [ ] _________________

### 4.2 Audit Logging Requirements

**Compliance Drivers:**
- Primary: ☐ PCI-DSS | ☐ HIPAA | ☐ SOC 2 | ☐ Internal policy | ☐ None yet
- Secondary: _________________

**SIEM Integration:**
- SIEM tool: ☐ Splunk | ☐ Elasticsearch | ☐ Datadog | ☐ Sumo Logic | ☐ Other: _____ | ☐ None
- Already configured: ☐ Yes | ☐ No | ☐ Partially
- Log retention: ___ days (compliance requirement: ___ days)
- Log format: ☐ JSON (recommended) | ☐ Plain text | ☐ Other: _____

**Audit Events to Capture:**
- [ ] All authorization decisions (required for most compliance)
- [ ] Authentication successes (HIPAA, SOC 2)
- [ ] Authentication failures (PCI-DSS, HIPAA, SOC 2)
- [ ] Configuration changes (SOC 2)
- [ ] Certificate rotations (SOC 2)
- [ ] Admin operations (all compliance frameworks)
- [ ] Custom events: _________________

### 4.3 Rate Limiting Configuration

**Authentication Rate Limiting:**
- Enable: ☐ Yes (recommended) | ☐ No | ☐ Undecided
- Max failures per minute: ___ (default: 10)
- Lockout duration: ___ minutes (default: 5)

**Per-Namespace Rate Limiting:**
- Enable: ☐ Yes | ☐ No | ☐ Undecided
- Default namespace limit: ___ req/sec (default: 1000)
- Custom limits needed:
  - Namespace 1: _______ → ___ req/sec
  - Namespace 2: _______ → ___ req/sec
  - Namespace 3: _______ → ___ req/sec

### 4.4 Certificate Pinning

**Remote Clusters:**
- Do you have remote clusters? ☐ Yes | ☐ No | ☐ Planned
- Number of remote clusters: ___
- Cluster 1: _______ (endpoint: _______)
- Cluster 2: _______ (endpoint: _______)
- Cluster 3: _______ (endpoint: _______)

**Pinning Strategy:**
- ☐ Strict mode (reject on mismatch) - recommended for production
- ☐ Non-strict mode (warn only) - for testing
- ☐ Not needed (single cluster deployment)

---

## Section 5: Technical Debt Priorities

### 5.1 Technical Debt Analysis

From TECHNICAL_DEBT_AUDIT.md, we found **913 items**. How aggressively do you want to address this?

**Q1 2026 Target:**
- ☐ Aggressive: Reduce to <100 items (89% reduction, ~20-25% team capacity)
- ☐ Standard: Reduce to <200 items (78% reduction, ~15-20% team capacity)
- ☐ Conservative: Reduce to <400 items (56% reduction, ~10-15% team capacity)
- ☐ Minimal: Address only Critical/High (5-10% team capacity)
- ☐ Defer: Focus on other priorities first

**Top Priority Components:**
Which components are most critical to your business? (check top 3-5)
- [ ] service/matching (121 items) - Task queue management
- [ ] service/history/workflow (94 items) - Core workflow execution
- [ ] service/history (52 items) - History service
- [ ] common/persistence (39 items) - Database layer
- [ ] tests (37 items) - Test infrastructure
- [ ] chasm (35 items) - Execution framework
- [ ] service/frontend (26 items) - API gateway
- [ ] Other: _________________

### 5.2 Specific Technical Debt Items

**High-Priority Issues to Address:** (from audit, add line numbers)
1. ___________________ (file:line)
2. ___________________ (file:line)
3. ___________________ (file:line)
4. ___________________ (file:line)
5. ___________________ (file:line)

**Can be deferred:**
- _________________
- _________________

---

## Section 6: Performance & Observability

### 6.1 Current Performance Baseline

**Latency (measure from your monitoring):**
- P50: ___ ms
- P95: ___ ms
- P99: ___ ms
- P99.9: ___ ms

**Throughput:**
- Current: ___ req/sec (average)
- Peak: ___ req/sec
- Target: ___ req/sec

**Error Rate:**
- Current: ___% (target: <0.1%)

### 6.2 Performance Targets (Q2 2026)

**Latency Improvement:**
- ☐ Aggressive: -30% (e.g., P95: 50ms → 35ms)
- ☐ Standard: -20% (e.g., P95: 50ms → 40ms) - default
- ☐ Conservative: -10% (e.g., P95: 50ms → 45ms)
- ☐ Maintain current (focus on other areas)

**Throughput Improvement:**
- ☐ +50% (significant investment)
- ☐ +30% (standard) - default
- ☐ +15% (conservative)
- ☐ Maintain current

### 6.3 Observability Requirements

**Distributed Tracing:**
- Priority: ☐ Critical | ☐ High | ☐ Medium | ☐ Low | ☐ Defer
- Backend: ☐ Jaeger | ☐ Zipkin | ☐ Datadog | ☐ New Relic | ☐ Other: _____
- Already deployed: ☐ Yes | ☐ No

**Advanced Metrics:**
- Per-namespace metrics: ☐ Critical | ☐ High | ☐ Medium | ☐ Low
- Replication metrics: ☐ Critical | ☐ High | ☐ Medium | ☐ Low | ☐ N/A (single region)
- Task queue metrics: ☐ Critical | ☐ High | ☐ Medium | ☐ Low

---

## Section 7: Disaster Recovery & Business Continuity

### 7.1 DR Requirements

**RTO (Recovery Time Objective):**
- Target: ___ minutes (default: <5 minutes)
- Acceptable: ___ minutes
- Current: ___ minutes (if known)

**RPO (Recovery Point Objective):**
- Target: ___ minutes (default: <1 minute)
- Acceptable: ___ minutes
- Current: ___ minutes (if known)

**Multi-Region Setup:**
- ☐ Already multi-region (regions: _________)
- ☐ Planning multi-region (timeline: _______)
- ☐ Single region only (no plans for multi-region)

### 7.2 DR Testing

**Testing Frequency:**
- ☐ Quarterly (recommended)
- ☐ Semi-annually
- ☐ Annually
- ☐ Ad-hoc
- ☐ Never tested

**Testing Scope:**
- ☐ Full regional failover
- ☐ Database failover only
- ☐ Service restart only
- ☐ Tabletop exercise only

---

## Section 8: Budget & Resource Constraints

### 8.1 Infrastructure Budget

**Cloud Costs:**
- Current monthly spend: $_____
- Budget for new infrastructure: $_____
- Budget constraints: ☐ Tight | ☐ Moderate | ☐ Flexible

**Tools/Services:**
- Budget for new tools: $_____ /year
- Approved vendors only: ☐ Yes (list: _______) | ☐ No

### 8.2 Team Budget

**Hiring Plans:**
- Security engineers: ___ (planned hires)
- SRE engineers: ___ (planned hires)
- Backend engineers: ___ (planned hires)
- Timeline: _____

**Training Budget:**
- Available for team training: $_____ /year
- Training priorities:
  - ☐ Security best practices
  - ☐ Kubernetes/cloud native
  - ☐ Observability tools
  - ☐ Performance optimization
  - ☐ Other: _____

---

## Section 9: Risk Tolerance & Approach

### 9.1 Deployment Risk Tolerance

**How risk-averse is your organization?**
- ☐ Very conservative (slow, careful, extensive testing)
- ☐ Moderate (balanced approach, standard testing)
- ☐ Aggressive (move fast, iterate quickly)

**Rollback Tolerance:**
- Acceptable rollback time: ___ minutes (default: <5 minutes)
- Rollback frequency acceptable: ☐ Multiple times OK | ☐ Once is OK | ☐ Zero rollbacks required

**Customer Impact Tolerance:**
- ☐ Zero customer impact required (may extend timeline)
- ☐ Minimal impact acceptable (<0.01% error rate)
- ☐ Moderate impact acceptable (<0.1% error rate)

### 9.2 Testing Requirements

**Testing Depth:**
- ☐ Extensive (2-4 weeks in staging, 1 week canary)
- ☐ Standard (1-2 weeks in staging, 3 days canary) - default
- ☐ Minimal (few days staging, 1 day canary)

**Load Testing:**
- Required: ☐ Yes | ☐ No | ☐ Optional
- Target load: ___x current traffic (default: 2x)
- Duration: ___ hours (default: 4-8 hours)

**Security Testing:**
- Penetration testing: ☐ Required | ☐ Optional | ☐ Skip
- Vendor: _________________ (or internal team)
- Timeline: _____

---

## Section 10: Communication & Stakeholders

### 10.1 Stakeholder Communication

**Key Stakeholders:**
- Engineering: _________________
- Security: _________________
- Product: _________________
- Customer Success: _________________
- Legal/Compliance: _________________
- Executive Sponsor: _________________

**Communication Frequency:**
- ☐ Daily updates during deployment
- ☐ Weekly updates during normal periods
- ☐ Bi-weekly updates
- ☐ Monthly updates

**Communication Channels:**
- Primary: ☐ Email | ☐ Slack | ☐ Teams | ☐ Confluence | ☐ Other: _____
- For incidents: ☐ PagerDuty | ☐ Phone | ☐ Slack | ☐ Other: _____

### 10.2 Customer Communication

**Customer-Facing Changes:**
- Are customers affected by security deployment? ☐ Yes | ☐ No | ☐ Potentially
- Customer notification required: ☐ Yes | ☐ No
- Notice period required: ___ days

**Communication Templates:**
- Pre-deployment announcement: ☐ Needed | ☐ Not needed
- During deployment updates: ☐ Needed | ☐ Not needed
- Post-deployment summary: ☐ Needed | ☐ Not needed

---

## Section 11: Custom Requirements

### 11.1 Organization-Specific Requirements

**Compliance/Audit:**
- Specific compliance requirements: _________________
- Internal audit schedule: _________________
- Documentation requirements: _________________

**Security Policies:**
- Password policy: _________________
- MFA requirements: _________________
- Certificate policy: _________________
- Other: _________________

### 11.2 Additional Customizations

**Specific needs not covered above:**
1. _________________
2. _________________
3. _________________

**Questions/Concerns:**
1. _________________
2. _________________
3. _________________

---

## Summary & Next Steps

### Filled Out By:
- Name: _________________
- Role: _________________
- Date: _________________

### Review & Approval:
- Reviewed by: _________________ (Engineering Manager)
- Approved by: _________________ (VP Engineering/CTO)
- Date: _________________

### Generate Custom Roadmap:

Once this questionnaire is complete, run:

```bash
# Generate customized roadmap based on your answers
./scripts/generate_custom_roadmap.sh \
  --input=PLANNING_QUESTIONNAIRE.md \
  --output=CUSTOM_ROADMAP_[YOUR_ORG].md

# This will create:
# - CUSTOM_ROADMAP_[YOUR_ORG].md (your personalized timeline)
# - CUSTOM_WEEK_1_GUIDE_[YOUR_ORG].md (your Week 1 plan)
# - CUSTOM_DEPLOYMENT_CHECKLIST_[YOUR_ORG].md (your checklist)
```

Or manually review your answers and adjust the timeline in:
- PROJECT_IMPLEMENTATION_ROADMAP.md
- WEEK_1_QUICK_START.md
- DEPLOYMENT_VERIFICATION_CHECKLIST.md

---

**Next Steps After Completion:**
1. [ ] Share with leadership for approval
2. [ ] Schedule kickoff meeting
3. [ ] Reserve deployment windows based on your answers
4. [ ] Begin Week 1 preparation

**Reference:**
- PROJECT_IMPLEMENTATION_ROADMAP.md (base plan to customize)
- WEEK_1_QUICK_START.md (base Week 1 to customize)
- TECHNICAL_DEBT_AUDIT.md (913 items found)
