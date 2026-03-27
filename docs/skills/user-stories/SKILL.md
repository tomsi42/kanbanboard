---
name: user-stories
description: Elicit and define user stories as the starting point of planning. Captures what users need before domain modeling, architecture, or UX design. Stories are revisited during UX layout to ensure the experience serves them.
---

# User Stories

User stories are the foundation of the plan. They capture what the user needs to accomplish, independent of how the system is built. Start here before domain encoding, architecture, or UX.

## Process

### Step 1: Identify the users

Ask:
- Who will use this application?
- Are there different types of users with different needs?
- Which user type is the primary audience?

Keep it simple. Most applications have 1-3 user types.

### Step 2: Elicit stories

For each user type, ask:
- What do they need to accomplish with this application?
- What problems are they trying to solve?
- What would make them choose this tool over their current approach?

Capture each need as a user story:

```
As a [user type], I want to [action] so that [benefit].
```

Focus on the **benefit** - it reveals the real need behind the request. If you can't articulate the benefit, the story isn't understood well enough.

### Step 3: Prioritize

Not all stories are equal. Sort them:

1. **Must have** - the application is useless without these
2. **Should have** - important but the app works without them
3. **Nice to have** - valuable but can wait

Be ruthless. Most projects have 3-5 must-have stories. If you have more than 7, you're probably not prioritizing hard enough.

### Step 4: Add acceptance criteria

For each must-have and should-have story, define:
- How do we know this story is done?
- What are the specific conditions that must be true?
- What are the edge cases?

```
Story: As a user, I want to create a task so that I can track my work.

Acceptance criteria:
- User can enter a task title
- Task appears on the board after creation
- Empty title is not allowed
```

Keep criteria concrete and testable.

### Step 5: Challenge the stories

- Are any stories really two stories combined? Split them.
- Are any stories solving the same need differently? Merge them.
- Does every must-have story genuinely block the application from being useful?
- Are you describing the solution instead of the need? ("I want a dropdown" vs "I want to categorize tasks")

## When to revisit

Stories are revisited during **UX Layout**:
- Do the screens and journeys serve every must-have story?
- Are there screens that don't map to any story? (Why do they exist?)
- Does the navigation make the highest-priority stories the easiest to accomplish?
- Did the UX design reveal new stories? (e.g. onboarding, settings screens)
- If new stories were discovered, add them and re-prioritize.

**Important:** The UX phase often uncovers stories that weren't obvious during initial elicitation. This is expected and healthy. Update the stories document when this happens.

Stories also feed into **implementation planning**:
- Must-have stories become early phases
- Should-have stories become later phases
- Nice-to-have stories go in a backlog

## Exit criteria

- [ ] User types identified
- [ ] Stories written with action and benefit
- [ ] Stories prioritized (must/should/nice)
- [ ] Must-have and should-have stories have acceptance criteria
- [ ] No combined stories - each story is one need
- [ ] Stories describe needs, not solutions
- [ ] User confirms the stories capture what they want to build
