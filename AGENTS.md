### YAGNI — You Aren't Gonna Need It

Add complexity only when there is a concrete, immediate need. Specifically:

- **No speculative abstractions**: Don't extract a helper, base class, or utility unless two or more places use it today. Three similar lines of code beat a premature abstraction.
- **No one-line helper wrappers**: Don't add a helper that wraps a single expression or call. Inline it unless it removes real duplication or names repeated domain behavior.
- **No future-proofing**: Don't add config flags, extension points, or generic interfaces for requirements nobody has yet.
- **No defensive over-engineering**: Don't handle errors, add fallbacks, or validate against edge cases that can't happen in the current codebase. Trust internal contracts.
- **No unused flexibility**: Don't make things configurable or overridable until a second consumer exists.
- **Scope-bound changes**: A bug fix changes only the broken behavior. A feature adds only what was asked for. No opportunistic refactors, no drive-by cleanup.

### Balanced Testing

Testing effort follows risk, not test count or coverage percentage:

- Rank failures by how likely they are and how much they hurt. Concentrate on payments, auth, tenant isolation, data integrity, irreversible actions, retries, and public contracts.
- Automate the regression paths that are stable and worth protecting. Cosmetic, trivial, or stable glue changes can rely on existing checks and review.
- Put each check at the lowest, fastest layer that can catch the failure. Reserve end-to-end and UI tests for critical user journeys, and delete the high-level checks that duplicate lower ones.
- Assert observable behavior and public contracts, not private implementation. Mock internals sparingly: a refactor that preserves behavior should leave the tests passing.
- Explore new, complex, or risky behavior by hand in a time-boxed session. Automate a finding only if it repeats and is worth protecting.
- Treat coverage as a diagnostic for risky gaps, never as a delivery target.
- Keep tests fast, isolated, and deterministic. Fix a flaky test, delete it, or quarantine it with a named owner and a condition for removal. Never paper over one with blanket retries.
- Write the test before the code (TDD) when the logic is complex or the behavior is still uncertain. Skip it for trivial glue and throwaway spikes.
