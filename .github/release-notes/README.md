# Release Notes

Each public tag must have a matching release note file:

```text
.github/release-notes/<tag>.md
```

Example:

```text
.github/release-notes/v0.3.0.md
```

Required structure:

```md
## [Add]
- ...

## [Change]
- ...

## [Fix]
- ...

## [Remove]
- ...
```

Rules:

1. Keep the four sections in the exact order above.
2. Use `- None.` if a section is empty.
3. Draft content against the previous formal release when possible.
4. The GitHub release workflow appends a compare-link changelog automatically.
5. Do not rely on GitHub auto-generated release notes as the final release body.
