# Development Workflow

## Branching strategy

Trunk-based development. All work on main.

Rationale: solo developer project, stop-and-review workflow at each sub-phase. Tags mark accepted milestones.

## Versioning

```
v0.{phase}.{subphase}
```

- Sub-phase number resets on new phase
- Version updated at sub-phase acceptance, not on every commit
- v1.0.0 when the plan is fully implemented

## Push strategy

Push after every commit. Tags mark the accepted milestone points.

## Review and acceptance flow

1. Implement the sub-phase on main
2. Stop and present to user for review
3. User tests what's available
4. If accepted: tag with version, move to next sub-phase
5. If not accepted: discuss issues, fix, return to step 2

## Commit practices

- Format: `type: description` (feat, fix, refactor, test, docs, chore)
- Small commits, one logical change each
- Never commit broken code to main

## Database migrations

Hand-rolled migration runner in Go:
- `migrations` table tracks which migrations have been applied
- On startup, scans migrations folder and applies pending ones in order
- Each migration is a plain `.sql` file
- Initial migration in Phase 1.2 creates full schema
- Later phases add migrations only if schema changes
