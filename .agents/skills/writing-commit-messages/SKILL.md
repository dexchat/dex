---
name: writing-commit-messages
description: Draft or apply Git commit messages for dexchat. Use when the user asks for a commit message, asks to commit changes, or wants an existing message reviewed.
---

# Writing Commit Messages

Write commit messages for one focused change. Follow the nearest applicable
`AGENTS.md` when it conflicts with this skill.

## Format

Use Conventional Commits:

```text
<type>[(optional-scope)]: <summary>

[optional references]

[optional body]

```

## Rules

### Subject line
- Choose the narrowest accurate semantic type, such as `fix`, `feat`, `chore`,
  `refactor`, `docs`, `test`, `ci`, `build`, `perf`, or `sec`.
- Add a scope only when it materially clarifies the subject.
- Start the summary in lowercase imperative mood, omit the trailing period,
  and keep the complete subject ideally under 60 characters.

### References
- Include issue, pull request, or discussion references only when known. Never
  invent them.

### Body
- Describe **what changed**, **what the previous behavior was**,
  and **how the new behavior works** at a high level.
- Use plain prose, not bullet points. Wrap lines at ~72 characters.
- Focus on the _why_ and _how_ rather than restating the diff.
- Keep the tone direct and technical without filler phrases.
- Don't exceed a handful of paragraphs; less is more.

## Workflow

- Run a diff to see what changes are present since the last commit.
- Inspect `git status`, the relevant diffs, and recent subjects. Determine the
  type from the change's intent, not only its file paths.
- Keep unrelated changes out of the commit, including unrelated hunks in a
  touched file.
- Draft the commit message following the format above.
- Apply the commit
- Don't push the commit; leave that to the user.
