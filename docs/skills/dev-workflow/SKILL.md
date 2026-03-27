---
name: dev-workflow
description: Define the development workflow before the first commit - branching strategy, versioning, review gates, and how code moves from idea to main. Use during planning, before implementation starts.
---

# Development Workflow

Establish how development will be conducted before writing any code. These decisions are easy to make up front and painful to retrofit.

## Process

### Step 1: Branching strategy

Decide between two approaches:

**Option A: Feature branches (recommended for phased projects)**
- Each phase/sub-phase gets its own branch
- Work happens on the branch
- User reviews and tests at each sub-phase boundary
- Merge to main when the sub-phase is accepted
- Main always contains accepted, working code

**Option B: Trunk-based (working directly on main)**
- All work happens on main
- Suitable for solo projects with very small increments
- Requires discipline to never leave main broken

Discuss trade-offs with the user and decide.

### Step 2: Versioning

For planned, phased projects, use phase-based versioning during development:

```
v0.{phase}.{subphase}
```

- `v0.1.0` - Phase 1 starts
- `v0.1.1` - Phase 1, sub-phase 1 complete
- `v0.1.2` - Phase 1, sub-phase 2 complete
- `v0.2.0` - Phase 2 starts (sub-phase resets)
- `v1.0.0` - Plan fully implemented, first release

Version is updated when the user accepts a sub-phase, not on every commit.

### Step 3: Sub-phase planning

Before starting each sub-phase, create a plan:

1. Scope what the sub-phase will deliver
2. List the files to create/modify
3. Define verification steps and manual test checklist
4. Get user approval of the plan before writing code

This ensures each sub-phase is well-scoped and avoids wasted work. Never combine multiple sub-phases into one plan — each sub-phase is planned and delivered independently, no matter how small.

### Step 4: Review and acceptance flow

Define the workflow for each sub-phase:

1. Create branch from main (if using feature branches)
2. Plan the sub-phase (Step 3 above)
3. Implement the sub-phase
4. Commit and push code
5. Present to user for review with a manual test checklist
6. User tests what's available
7. If accepted: tag with version, push tag
8. If not accepted: discuss issues, fix, return to step 5

**Important:** Tagging happens only after the user has tested and accepted the sub-phase. Do not tag immediately after coding — the user's acceptance is the gate.

### Step 4b: Test data strategy

Agree at the start of the project on:
- **Shared test accounts** — standard credentials everyone uses (e.g. admin@test.com / password1)
- **When to wipe** — only when schema changes require it, not on every rebuild
- **When to preserve** — don't use `docker compose down -v` unless necessary; the user's manual test data should survive rebuilds
- **Who creates test data** — the developer creates standard accounts via API; the user can also use them

### Step 4c: Testing gate

Automated tests must pass before presenting a sub-phase for acceptance:
- Run `go test ./...` (or equivalent) before pushing
- If the testing strategy says backend tests are required, they must exist and pass
- Manual test checklist is for the user; automated tests are for the developer

### Step 4d: LEARNINGS.md

Create a `LEARNINGS.md` file at project setup. Update it during development whenever you discover something surprising or useful:

**What to capture:**
- Library incompatibilities (e.g. "svelte-dnd-action doesn't work with Svelte 5 $state proxies")
- Framework quirks (e.g. "PostgreSQL 18 changed the default volume mount path")
- Workarounds that aren't obvious (e.g. "UNIQUE constraints need temporary negative values during reorder")
- Performance discoveries
- Things that looked simple but weren't (e.g. "click vs drag detection needs distance-based approach")
- Useful patterns that emerged

**Format:** Keep it simple — date, topic, what happened, what the solution was.

```markdown
## 2026-03-27: svelte-dnd-action + Svelte 5
svelte-dnd-action has known issues with Svelte 5's $state proxy objects.
Switched to @thisux/sveltednd which is built for Svelte 5 runes.
```

**When to review:**
- During the project debrief — check if learnings should become skill updates
- Before starting the next version — refresh your memory on gotchas
- When onboarding someone new to the project

### Step 5: Commit practices

- Commit messages: `type: description` (feat, fix, refactor, test, docs, chore)
- Small commits: one logical change per commit
- Never commit broken code to main
- Tag accepted sub-phases: `git tag v0.1.1`

### Step 6: Document the workflow

Write down the agreed workflow so it's followed consistently throughout the project. Include:
- Branching strategy chosen and why
- Versioning scheme
- Review/acceptance process
- Commit message format

## Exit criteria

- [ ] Branching strategy decided
- [ ] Versioning scheme agreed
- [ ] Review and acceptance flow defined
- [ ] Commit practices established
- [ ] User understands and agrees to the workflow
