# Refracture Analysis

Date: 2026-02-12
Scope: `/root/Sunspear` backend + frontend codebase review (static analysis)

## Executive Summary
The project is feature-rich but currently has high fracture risk in three areas:
1. Security and auth boundary hardening.
2. Reliability and recoverability of compose/app lifecycle operations.
3. Maintainability due to large view/service files and missing automated tests.

## Implementation Status (This Pass)

Completed:
1. Hardened auth token handling so query-parameter tokens are accepted only for WebSocket routes.
2. Improved rate limiting behavior by using one shared limiter instance for auth endpoints and normalized client IP extraction.
3. Added compose lifecycle error aggregation to prevent silent stop/delete/rollback failures and avoid deleting DB state on partial teardown failure.

Not completed:
1. Distributed/shared rate limit backend (Redis/proxy policy) is still pending.
2. Auth model migration from `localStorage` token to HttpOnly cookies is pending.
3. Automated test coverage and large-file decomposition are pending.

## Key Findings

### Critical

1. JWT accepted via query parameter on all protected API routes.
- Evidence: `backend/api/middleware/auth.go:30`
- Risk: URL tokens can leak via logs/history/proxies and may be replayed.
- Recommendation: Only allow query-param token for WebSocket paths (`/api/ws/*`) and require `Authorization: Bearer` elsewhere.

### High

2. In-memory, per-process rate limiting is easy to bypass in distributed or restarted deployments.
- Evidence: `backend/api/middleware/ratelimit.go:73`
- Risk: Brute-force protection weak if multiple backend instances exist or process restarts.
- Recommendation: Move auth rate limiting to a shared backend (Redis) or enforce at reverse proxy (Caddy) with trusted client IP extraction.

3. Compose lifecycle ignores some destructive-operation errors, causing partial state drift.
- Evidence: `backend/services/compose.go:236`, `backend/services/compose.go:289`, `backend/services/compose.go:290`, `backend/services/compose.go:295`
- Risk: DB may mark project stopped/removed while resources remain running/orphaned.
- Recommendation: Aggregate and return operation errors; persist per-resource status; add reconciliation job for orphan cleanup.

4. Compose service is a monolith (~795 LOC) mixing parsing, orchestration, rollback, templates, and persistence.
- Evidence: `backend/services/compose.go`
- Risk: Change collisions and regression probability increase as features expand.
- Recommendation: Split into `compose_parser`, `compose_orchestrator`, `compose_templates`, and `compose_repository` modules.

### Medium

5. Frontend auth token stored in `localStorage`.
- Evidence: `frontend/src/stores/auth.js:6`, `frontend/src/stores/auth.js:13`
- Risk: Token exposure under XSS conditions.
- Recommendation: Prefer HttpOnly secure cookie session or short-lived access token + refresh token rotation.

6. Very large Vue view files increase refactor breakage risk.
- Evidence: `frontend/src/views/AppStore.vue` (778 LOC), `frontend/src/views/ContainerDetail.vue` (765 LOC), `frontend/src/views/Compose.vue` (748 LOC)
- Risk: UI behavior coupling, hard-to-test branching, merge conflicts.
- Recommendation: Extract page-specific composables + subcomponents (filters/panels/modals/action bars).

7. Minimal password policy (length-only).
- Evidence: `backend/api/handlers/validation.go:5`
- Risk: Weak credentials accepted.
- Recommendation: Add configurable password strength requirements and compromised-password checks.

### Low

8. Legacy header `X-XSS-Protection` still set.
- Evidence: `backend/api/middleware/security.go:10`
- Risk: Low direct risk; mostly obsolete behavior in modern browsers.
- Recommendation: Replace with stronger CSP strategy and keep modern hardening headers.

## Testability and Verification Gaps

1. No backend test files discovered under `backend/`.
2. No frontend test framework/scripts in `frontend/package.json`.
3. Command validation gap: `go test ./...` could not run because `go` is unavailable in this environment.

## Prioritized Recommendations

### Phase 1 (Immediate, 1-3 days)
1. Restrict token-in-query usage to WebSocket endpoints only.
2. Implement robust rate limiting at the proxy or shared store.
3. Add operation-level error handling for compose stop/delete/rollback.

### Phase 2 (Short term, 1-2 weeks)
1. Introduce backend integration tests for auth, compose deploy/delete, and lifecycle rollback.
2. Add frontend unit tests for auth store, API interceptor, and key page composables.
3. Break the largest 3 view files into smaller components/composables.

### Phase 3 (Medium term, 2-4 weeks)
1. Refactor compose domain into smaller services with clear interfaces.
2. Adopt a safer auth session model (HttpOnly cookie or rotated short-lived tokens).
3. Add a CSP rollout plan and tighten security headers.

## Success Criteria
1. Auth token appears only in headers/cookies (except explicit WS handshake path).
2. Failed compose operations produce deterministic error reports and no silent orphan resources.
3. Backend and frontend CI include automated tests for critical paths.
4. Largest frontend pages reduced to maintainable component boundaries (<300 LOC per view target).
