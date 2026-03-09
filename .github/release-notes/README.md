# Release Notes

If a file named `.github/release-notes/<tag>.md` exists, the release workflow prepends it to the GitHub-generated changelog for that tag.

Example:

```text
.github/release-notes/v0.3.0.md
```

Recommended flow:

1. Add or update the summary file for the tag you are about to release.
2. Commit and push the change to `main`.
3. Create and push the tag.
