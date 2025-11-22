# Technical Debt Audit Report

**Generated:** $(date)
**Repository:** Temporal Server
**Branch:** $(git branch --show-current)

---

## Executive Summary

This report provides a comprehensive audit of all technical debt markers (TODO, FIXME, XXX, HACK) found in the codebase.

**Total Technical Debt Items:** 913

| Category | Count | Percentage |
|----------|-------|------------|
| TODO | 901 | 98.6% |
| FIXME | 2 | .2% |
| XXX | 8 | .8% |
| HACK | 2 | .2% |

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

### Top 10 Components by Debt Count

- `./service/matching`: 121 items
- `./service/history/workflow`: 94 items
- `./service/history`: 52 items
- `./common/persistence/sql/sqlplugin/tests`: 39 items
- `./tests`: 37 items
- `./chasm`: 35 items
- `.`: 34 items
- `./service/frontend`: 26 items
- `./tests/xdc`: 24 items
- `./service/history/ndc`: 24 items

---

## Detailed Listings

### TODO Items (901)

Items marked with TODO indicate planned future work or improvements.


#### `api/enums/v1/workflow_task_type.pb.go:31`
**Comment:** (alex): TRANSIENT is not current used. Needs to be set when Attempt>1.

**Categorization:** [ ] Critical | [ ] High | [ ] Medium | [ ] Low

**Estimated Effort:** [ ] 1-2 hours | [ ] 1 day | [ ] 2-3 days | [ ] 1 week | [ ] >1 week

**GitHub Issue:** #_____

**Notes:**
- Impact:
- Dependencies:
- Assigned to:

---


#### `api/historyservice/v1/request_response.pb.go:8735`
**Comment:** (Tianyu): This is the same as NexusOperationsCompletion but obviously is not about Nexus. This is because

**Categorization:** [ ] Critical | [ ] High | [ ] Medium | [ ] Low

**Estimated Effort:** [ ] 1-2 hours | [ ] 1 day | [ ] 2-3 days | [ ] 1 week | [ ] >1 week

**GitHub Issue:** #_____

**Notes:**
- Impact:
- Dependencies:
- Assigned to:

---


#### `api/matchingservice/v1/service_grpc.pb.go:149`
**Comment:** Shivam - remove this in 123. Present for backwards compatibility.

**Categorization:** [ ] Critical | [ ] High | [ ] Medium | [ ] Low

**Estimated Effort:** [ ] 1-2 hours | [ ] 1 day | [ ] 2-3 days | [ ] 1 week | [ ] >1 week

**GitHub Issue:** #_____

**Notes:**
- Impact:
- Dependencies:
- Assigned to:

---


#### `api/matchingservice/v1/service_grpc.pb.go:654`
**Comment:** Shivam - remove this in 123. Present for backwards compatibility.

**Categorization:** [ ] Critical | [ ] High | [ ] Medium | [ ] Low

**Estimated Effort:** [ ] 1-2 hours | [ ] 1 day | [ ] 2-3 days | [ ] 1 week | [ ] >1 week

**GitHub Issue:** #_____

**Notes:**
- Impact:
- Dependencies:
- Assigned to:

---


#### `api/matchingservice/v1/request_response.pb.go:2905`
**Comment:** Shivam - Please remove this in 123

**Categorization:** [ ] Critical | [ ] High | [ ] Medium | [ ] Low

**Estimated Effort:** [ ] 1-2 hours | [ ] 1 day | [ ] 2-3 days | [ ] 1 week | [ ] >1 week

**GitHub Issue:** #_____

**Notes:**
- Impact:
- Dependencies:
- Assigned to:

---


#### `api/matchingservice/v1/request_response.pb.go:2966`
**Comment:** Shivam - Please remove this in 123

**Categorization:** [ ] Critical | [ ] High | [ ] Medium | [ ] Low

**Estimated Effort:** [ ] 1-2 hours | [ ] 1 day | [ ] 2-3 days | [ ] 1 week | [ ] >1 week

**GitHub Issue:** #_____

**Notes:**
- Impact:
- Dependencies:
- Assigned to:

---


#### `api/persistence/v1/executions.pb.go:1525`
**Comment:** Remove this field when state-based replication is stable and

**Categorization:** [ ] Critical | [ ] High | [ ] Medium | [ ] Low

**Estimated Effort:** [ ] 1-2 hours | [ ] 1 day | [ ] 2-3 days | [ ] 1 week | [ ] >1 week

**GitHub Issue:** #_____

**Notes:**
- Impact:
- Dependencies:
- Assigned to:

---


#### `api/replication/v1/message.pb.go:688`
**Comment:** Deprecate this definition, it only used by the deprecated replication DLQ v1 logic

**Categorization:** [ ] Critical | [ ] High | [ ] Medium | [ ] Low

**Estimated Effort:** [ ] 1-2 hours | [ ] 1 day | [ ] 2-3 days | [ ] 1 week | [ ] >1 week

**GitHub Issue:** #_____

**Notes:**
- Impact:
- Dependencies:
- Assigned to:

---


#### `chasm/lib/callback/nexus_invocation.go:139`
**Comment:** This logic is duplicated in the frontend handler for forwarded requests. Eventually it should live in the Nexus SDK.

**Categorization:** [ ] Critical | [ ] High | [ ] Medium | [ ] Low

**Estimated Effort:** [ ] 1-2 hours | [ ] 1 day | [ ] 2-3 days | [ ] 1 week | [ ] >1 week

**GitHub Issue:** #_____

**Notes:**
- Impact:
- Dependencies:
- Assigned to:

---


#### `chasm/lib/callback/executors_test.go:58`
**Comment:** (seankane): Move this helper to the chasm/chasmtest package

**Categorization:** [ ] Critical | [ ] High | [ ] Medium | [ ] Low

**Estimated Effort:** [ ] 1-2 hours | [ ] 1 day | [ ] 2-3 days | [ ] 1 week | [ ] >1 week

**GitHub Issue:** #_____

**Notes:**
- Impact:
- Dependencies:
- Assigned to:

---


#### `chasm/lib/callback/component.go:54`
**Comment:** (seankane): implement lifecycle state

**Categorization:** [ ] Critical | [ ] High | [ ] Medium | [ ] Low

**Estimated Effort:** [ ] 1-2 hours | [ ] 1 day | [ ] 2-3 days | [ ] 1 week | [ ] >1 week

**GitHub Issue:** #_____

**Notes:**
- Impact:
- Dependencies:
- Assigned to:

---


#### `chasm/lib/scheduler/invoker_tasks.go:86`
**Comment:** - dial this up/remove it

**Categorization:** [ ] Critical | [ ] High | [ ] Medium | [ ] Low

**Estimated Effort:** [ ] 1-2 hours | [ ] 1 day | [ ] 2-3 days | [ ] 1 week | [ ] >1 week

**GitHub Issue:** #_____

**Notes:**
- Impact:
- Dependencies:
- Assigned to:

---


#### `chasm/lib/scheduler/invoker_tasks.go:545`
**Comment:** - set search attributes

**Categorization:** [ ] Critical | [ ] High | [ ] Medium | [ ] Low

**Estimated Effort:** [ ] 1-2 hours | [ ] 1 day | [ ] 2-3 days | [ ] 1 week | [ ] >1 week

**GitHub Issue:** #_____

**Notes:**
- Impact:
- Dependencies:
- Assigned to:

---


#### `chasm/lib/scheduler/backfiller_tasks_test.go:241`
**Comment:** - remove this when CHASM has unit testing hooks for task generation

**Categorization:** [ ] Critical | [ ] High | [ ] Medium | [ ] Low

**Estimated Effort:** [ ] 1-2 hours | [ ] 1 day | [ ] 2-3 days | [ ] 1 week | [ ] >1 week

**GitHub Issue:** #_____

**Notes:**
- Impact:
- Dependencies:
- Assigned to:

---


#### `chasm/lib/scheduler/backfiller_tasks_test.go:250`
**Comment:** - check that a pure task to continue driving backfill exists here. Because

**Categorization:** [ ] Critical | [ ] High | [ ] Medium | [ ] Low

**Estimated Effort:** [ ] 1-2 hours | [ ] 1 day | [ ] 2-3 days | [ ] 1 week | [ ] >1 week

**GitHub Issue:** #_____

**Notes:**
- Impact:
- Dependencies:
- Assigned to:

---


#### `chasm/lib/scheduler/scheduler.go:115`
**Comment:** namespace name should be resolved from namespace_id via namespace registry

**Categorization:** [ ] Critical | [ ] High | [ ] Medium | [ ] Low

**Estimated Effort:** [ ] 1-2 hours | [ ] 1 day | [ ] 2-3 days | [ ] 1 week | [ ] >1 week

**GitHub Issue:** #_____

**Notes:**
- Impact:
- Dependencies:
- Assigned to:

---


#### `chasm/lib/scheduler/scheduler.go:125`
**Comment:** - use visibility component to update SAs

**Categorization:** [ ] Critical | [ ] High | [ ] Medium | [ ] Low

**Estimated Effort:** [ ] 1-2 hours | [ ] 1 day | [ ] 2-3 days | [ ] 1 week | [ ] >1 week

**GitHub Issue:** #_____

**Notes:**
- Impact:
- Dependencies:
- Assigned to:

---


#### `chasm/lib/scheduler/scheduler.go:401`
**Comment:** - also record payload sizes once we have metrics wired into CHASM context.

**Categorization:** [ ] Critical | [ ] High | [ ] Medium | [ ] Low

**Estimated Effort:** [ ] 1-2 hours | [ ] 1 day | [ ] 2-3 days | [ ] 1 week | [ ] >1 week

**GitHub Issue:** #_____

**Notes:**
- Impact:
- Dependencies:
- Assigned to:

---


#### `chasm/lib/scheduler/scheduler.go:470`
**Comment:** - softassert here when we have a logger wired into CHASM. Skip recording

**Categorization:** [ ] Critical | [ ] High | [ ] Medium | [ ] Low

**Estimated Effort:** [ ] 1-2 hours | [ ] 1 day | [ ] 2-3 days | [ ] 1 week | [ ] >1 week

**GitHub Issue:** #_____

**Notes:**
- Impact:
- Dependencies:
- Assigned to:

---


#### `chasm/lib/scheduler/scheduler.go:514`
**Comment:** - memo and search_attributes are handled by visibility (separate PR)

**Categorization:** [ ] Critical | [ ] High | [ ] Medium | [ ] Low

**Estimated Effort:** [ ] 1-2 hours | [ ] 1 day | [ ] 2-3 days | [ ] 1 week | [ ] >1 week

**GitHub Issue:** #_____

**Notes:**
- Impact:
- Dependencies:
- Assigned to:

---


#### `chasm/lib/scheduler/scheduler.go:540`
**Comment:** - use visibility component to update SAs

**Categorization:** [ ] Critical | [ ] High | [ ] Medium | [ ] Low

**Estimated Effort:** [ ] 1-2 hours | [ ] 1 day | [ ] 2-3 days | [ ] 1 week | [ ] >1 week

**GitHub Issue:** #_____

**Notes:**
- Impact:
- Dependencies:
- Assigned to:

---


#### `chasm/lib/workflow/workflow.go:18`
**Comment:** populate with actual callback component type

**Categorization:** [ ] Critical | [ ] High | [ ] Medium | [ ] Low

**Estimated Effort:** [ ] 1-2 hours | [ ] 1 day | [ ] 2-3 days | [ ] 1 week | [ ] >1 week

**GitHub Issue:** #_____

**Notes:**
- Impact:
- Dependencies:
- Assigned to:

---


#### `chasm/test_task_test.go:1`
**Comment:** move this to chasm_test package

**Categorization:** [ ] Critical | [ ] High | [ ] Medium | [ ] Low

**Estimated Effort:** [ ] 1-2 hours | [ ] 1 day | [ ] 2-3 days | [ ] 1 week | [ ] >1 week

**GitHub Issue:** #_____

**Notes:**
- Impact:
- Dependencies:
- Assigned to:

---


#### `chasm/tree.go:100`
**Comment:** Consider using unique package here.

**Categorization:** [ ] Critical | [ ] High | [ ] Medium | [ ] Low

**Estimated Effort:** [ ] 1-2 hours | [ ] 1 day | [ ] 2-3 days | [ ] 1 week | [ ] >1 week

**GitHub Issue:** #_____

**Notes:**
- Impact:
- Dependencies:
- Assigned to:

---


#### `chasm/tree.go:160`
**Comment:** Return tree size changes in NodesMutation as well. MutateState needs to

**Categorization:** [ ] Critical | [ ] High | [ ] Medium | [ ] Low

**Estimated Effort:** [ ] 1-2 hours | [ ] 1 day | [ ] 2-3 days | [ ] 1 week | [ ] >1 week

**GitHub Issue:** #_____

**Notes:**
- Impact:
- Dependencies:
- Assigned to:

---


#### `chasm/tree.go:179`
**Comment:** Add methods needed from MutateState here.

**Categorization:** [ ] Critical | [ ] High | [ ] Medium | [ ] Low

**Estimated Effort:** [ ] 1-2 hours | [ ] 1 day | [ ] 2-3 days | [ ] 1 week | [ ] >1 week

**GitHub Issue:** #_____

**Notes:**
- Impact:
- Dependencies:
- Assigned to:

---


#### `chasm/tree.go:203`
**Comment:** Return a iterator on node name instead of []string,

**Categorization:** [ ] Critical | [ ] High | [ ] Medium | [ ] Low

**Estimated Effort:** [ ] 1-2 hours | [ ] 1 day | [ ] 2-3 days | [ ] 1 week | [ ] >1 week

**GitHub Issue:** #_____

**Notes:**
- Impact:
- Dependencies:
- Assigned to:

---


#### `chasm/tree.go:420`
**Comment:** - check if this is a detached node, operations are always allowed.

**Categorization:** [ ] Critical | [ ] High | [ ] Medium | [ ] Low

**Estimated Effort:** [ ] 1-2 hours | [ ] 1 day | [ ] 2-3 days | [ ] 1 week | [ ] >1 week

**GitHub Issue:** #_____

**Notes:**
- Impact:
- Dependencies:
- Assigned to:

---


#### `chasm/tree.go:1217`
**Comment:** Consider using node's LastUpdateVersionedTransition for checking staleness here.

**Categorization:** [ ] Critical | [ ] High | [ ] Medium | [ ] Low

**Estimated Effort:** [ ] 1-2 hours | [ ] 1 day | [ ] 2-3 days | [ ] 1 week | [ ] >1 week

**GitHub Issue:** #_____

**Notes:**
- Impact:
- Dependencies:
- Assigned to:

---


#### `chasm/tree.go:1269`
**Comment:** Now() could be different for components after we support Pause for CHASM components.

**Categorization:** [ ] Critical | [ ] High | [ ] Medium | [ ] Low

**Estimated Effort:** [ ] 1-2 hours | [ ] 1 day | [ ] 2-3 days | [ ] 1 week | [ ] >1 week

**GitHub Issue:** #_____

**Notes:**
- Impact:
- Dependencies:
- Assigned to:

---


#### `chasm/tree.go:1280`
**Comment:** remove the task type check after scheduler unit tests are fixed.

**Categorization:** [ ] Critical | [ ] High | [ ] Medium | [ ] Low

**Estimated Effort:** [ ] 1-2 hours | [ ] 1 day | [ ] 2-3 days | [ ] 1 week | [ ] >1 week

**GitHub Issue:** #_____

**Notes:**
- Impact:
- Dependencies:
- Assigned to:

---


#### `chasm/tree.go:1359`
**Comment:** Maintain a mapping from deserialized component value to node

**Categorization:** [ ] Critical | [ ] High | [ ] Medium | [ ] Low

**Estimated Effort:** [ ] 1-2 hours | [ ] 1 day | [ ] 2-3 days | [ ] 1 week | [ ] >1 week

**GitHub Issue:** #_____

**Notes:**
- Impact:
- Dependencies:
- Assigned to:

---


#### `chasm/tree.go:1381`
**Comment:** sync structure after every task execution instead of once per node to handle the case

**Categorization:** [ ] Critical | [ ] High | [ ] Medium | [ ] Low

**Estimated Effort:** [ ] 1-2 hours | [ ] 1 day | [ ] 2-3 days | [ ] 1 week | [ ] >1 week

**GitHub Issue:** #_____

**Notes:**
- Impact:
- Dependencies:
- Assigned to:

---


#### `chasm/tree.go:1455`
**Comment:** (), n)

**Categorization:** [ ] Critical | [ ] High | [ ] Medium | [ ] Low

**Estimated Effort:** [ ] 1-2 hours | [ ] 1 day | [ ] 2-3 days | [ ] 1 week | [ ] >1 week

**GitHub Issue:** #_____

**Notes:**
- Impact:
- Dependencies:
- Assigned to:

---


#### `chasm/tree.go:1525`
**Comment:** (), n)

**Categorization:** [ ] Critical | [ ] High | [ ] Medium | [ ] Low

**Estimated Effort:** [ ] 1-2 hours | [ ] 1 day | [ ] 2-3 days | [ ] 1 week | [ ] >1 week

**GitHub Issue:** #_____

**Notes:**
- Impact:
- Dependencies:
- Assigned to:

---


#### `chasm/tree.go:1642`
**Comment:** We cannot simply assert that all tasks in n.nodeBase.newTasks are processed.

**Categorization:** [ ] Critical | [ ] High | [ ] Medium | [ ] Low

**Estimated Effort:** [ ] 1-2 hours | [ ] 1 day | [ ] 2-3 days | [ ] 1 week | [ ] >1 week

**GitHub Issue:** #_____

**Notes:**
- Impact:
- Dependencies:
- Assigned to:

---


#### `chasm/tree.go:1983`
**Comment:** add assertion on IsDirty() once implemented

**Categorization:** [ ] Critical | [ ] High | [ ] Medium | [ ] Low

**Estimated Effort:** [ ] 1-2 hours | [ ] 1 day | [ ] 2-3 days | [ ] 1 week | [ ] >1 week

**GitHub Issue:** #_____

**Notes:**
- Impact:
- Dependencies:
- Assigned to:

---


#### `chasm/tree.go:2042`
**Comment:** combine this with the logic in CloseTransactionForceUpdateVisibility

**Categorization:** [ ] Critical | [ ] High | [ ] Medium | [ ] Low

**Estimated Effort:** [ ] 1-2 hours | [ ] 1 day | [ ] 2-3 days | [ ] 1 week | [ ] >1 week

**GitHub Issue:** #_____

**Notes:**
- Impact:
- Dependencies:
- Assigned to:

---


#### `chasm/tree.go:2416`
**Comment:** instead of tracking processed nodes,

**Categorization:** [ ] Critical | [ ] High | [ ] Medium | [ ] Low

**Estimated Effort:** [ ] 1-2 hours | [ ] 1 day | [ ] 2-3 days | [ ] 1 week | [ ] >1 week

**GitHub Issue:** #_____

**Notes:**
- Impact:
- Dependencies:
- Assigned to:

---


#### `chasm/tree.go:2507`
**Comment:** sync structure after each task since it's possible for a task to delete the node generated it

**Categorization:** [ ] Critical | [ ] High | [ ] Medium | [ ] Low

**Estimated Effort:** [ ] 1-2 hours | [ ] 1 day | [ ] 2-3 days | [ ] 1 week | [ ] >1 week

**GitHub Issue:** #_____

**Notes:**
- Impact:
- Dependencies:
- Assigned to:

---


#### `chasm/tree.go:2664`
**Comment:** consider pre-calculating the proto field num when registring the task type.

**Categorization:** [ ] Critical | [ ] High | [ ] Medium | [ ] Low

**Estimated Effort:** [ ] 1-2 hours | [ ] 1 day | [ ] 2-3 days | [ ] 1 week | [ ] >1 week

**GitHub Issue:** #_____

**Notes:**
- Impact:
- Dependencies:
- Assigned to:

---


#### `chasm/tree.go:2715`
**Comment:** consider pre-calculating the proto field num when registring the task type.

**Categorization:** [ ] Critical | [ ] High | [ ] Medium | [ ] Low

**Estimated Effort:** [ ] 1-2 hours | [ ] 1 day | [ ] 2-3 days | [ ] 1 week | [ ] >1 week

**GitHub Issue:** #_____

**Notes:**
- Impact:
- Dependencies:
- Assigned to:

---


#### `chasm/tree.go:2800`
**Comment:** - a task validator must succeed validation after a task executes

**Categorization:** [ ] Critical | [ ] High | [ ] Medium | [ ] Low

**Estimated Effort:** [ ] 1-2 hours | [ ] 1 day | [ ] 2-3 days | [ ] 1 week | [ ] >1 week

**GitHub Issue:** #_____

**Notes:**
- Impact:
- Dependencies:
- Assigned to:

---


#### `chasm/tree.go:2883`
**Comment:** Change physical side effect task to reference logical task and

**Categorization:** [ ] Critical | [ ] High | [ ] Medium | [ ] Low

**Estimated Effort:** [ ] 1-2 hours | [ ] 1 day | [ ] 2-3 days | [ ] 1 week | [ ] >1 week

**GitHub Issue:** #_____

**Notes:**
- Impact:
- Dependencies:
- Assigned to:

---


#### `chasm/tree.go:2948`
**Comment:** Change physical side effect task to reference logical task and

**Categorization:** [ ] Critical | [ ] High | [ ] Medium | [ ] Low

**Estimated Effort:** [ ] 1-2 hours | [ ] 1 day | [ ] 2-3 days | [ ] 1 week | [ ] >1 week

**GitHub Issue:** #_____

**Notes:**
- Impact:
- Dependencies:
- Assigned to:

---


#### `chasm/registrable_component.go:59`
**Comment:** remove WithShardingFn, we don't need this functionality.

**Categorization:** [ ] Critical | [ ] High | [ ] Medium | [ ] Low

**Estimated Effort:** [ ] 1-2 hours | [ ] 1 day | [ ] 2-3 days | [ ] 1 week | [ ] >1 week

**GitHub Issue:** #_____

**Notes:**
- Impact:
- Dependencies:
- Assigned to:

---


#### `chasm/test_component_test.go:1`
**Comment:** move this to chasm_test package

**Categorization:** [ ] Critical | [ ] High | [ ] Medium | [ ] Low

**Estimated Effort:** [ ] 1-2 hours | [ ] 1 day | [ ] 2-3 days | [ ] 1 week | [ ] >1 week

**GitHub Issue:** #_____

**Notes:**
- Impact:
- Dependencies:
- Assigned to:

---


#### `chasm/field_test.go:243`
**Comment:** - this doesn't resolve, but I've manually verified the tree structure looks correct

**Categorization:** [ ] Critical | [ ] High | [ ] Medium | [ ] Low

**Estimated Effort:** [ ] 1-2 hours | [ ] 1 day | [ ] 2-3 days | [ ] 1 week | [ ] >1 week

**GitHub Issue:** #_____

**Notes:**
- Impact:
- Dependencies:
- Assigned to:

---


#### `chasm/field_test.go:244`
**Comment:** 

**Categorization:** [ ] Critical | [ ] High | [ ] Medium | [ ] Low

**Estimated Effort:** [ ] 1-2 hours | [ ] 1 day | [ ] 2-3 days | [ ] 1 week | [ ] >1 week

**GitHub Issue:** #_____

**Notes:**
- Impact:
- Dependencies:
- Assigned to:

---


#### `chasm/field_test.go:245`
**Comment:** 

**Categorization:** [ ] Critical | [ ] High | [ ] Medium | [ ] Low

**Estimated Effort:** [ ] 1-2 hours | [ ] 1 day | [ ] 2-3 days | [ ] 1 week | [ ] >1 week

**GitHub Issue:** #_____

**Notes:**
- Impact:
- Dependencies:
- Assigned to:

---

