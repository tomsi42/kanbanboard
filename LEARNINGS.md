# Learnings

Discoveries, gotchas, and workarounds found during development.

## 2026-03-27: svelte-dnd-action + Svelte 5

`svelte-dnd-action` has known compatibility issues with Svelte 5's `$state` proxy objects. Drag-and-drop silently breaks — items become invisible or the library's internal state fights with Svelte's reactivity.

**Solution:** Switched to `@thisux/sveltednd` which is built specifically for Svelte 5 runes. Simpler API (draggable + droppable actions) and works correctly with `$state`.

**Lesson:** Always verify library compatibility with the specific framework version before committing. Check GitHub issues for the library + framework version combination.

## 2026-03-27: PostgreSQL 18 volume mount path

PostgreSQL 18 changed the Docker image to use version-specific subdirectories under `/var/lib/postgresql/`. The old volume mount at `/var/lib/postgresql/data` causes a startup error.

**Solution:** Mount the volume at `/var/lib/postgresql` instead of `/var/lib/postgresql/data`.

## 2026-03-27: Click vs drag detection

When using a drag-and-drop library, click events on draggable items are unreliable. The drag library captures pointer events, preventing normal click handling.

Tried: timing-based (200ms threshold) — unreliable, small mouse movements during click triggered drag state.
Tried: pointermove-based — any tiny movement set isDragging=true.

**Solution:** Distance-based detection. Track pointer position on `pointerdown`, compare with position on `click`. If delta < 5px, it's a click; otherwise it's a drag.

## 2026-03-27: UNIQUE constraint during column reorder

The `columns` table has a UNIQUE constraint on `(project_id, position)`. When swapping column positions in a loop, two columns temporarily have the same position, violating the constraint.

**Solution:** Two-pass approach within a transaction: first set all positions to negative values (-(i+1)), then set to final values. Negative positions never conflict with the UNIQUE constraint.

## 2026-03-27: Svelte 5 $state with prop initial values

Svelte 5 warns when using `$state(prop.value)` because it only captures the initial value. The prop is reactive but the `$state` isn't synced automatically.

**Solution:** Use `$effect` to resync the local state when the prop changes. This is intentional in Svelte 5 — local mutable copies of props need explicit sync.
