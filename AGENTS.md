# AGENTS.md

## Release Rules

- Do not use GitHub auto-generated release notes as the final release body.
- For every release or pre-release tag, create `.github/release-notes/<tag>.md` before pushing the tag.
- Every release note file must contain these four sections in this exact order:
  - `## [Add]`
  - `## [Change]`
  - `## [Fix]`
  - `## [Remove]`
- If a section has no meaningful entries, write `- None.`
- When drafting release note content, compare against the previous formal release tag if one exists.
- If there is no previous formal release tag, treat the release as the initial stable release and summarize the current capability set.
- The `## Changelog` section must not be AI-summarized text. It should only contain a compare link in the style:
  - `**Full Changelog**: https://github.com/<owner>/<repo>/compare/<previous-tag>...<current-tag>`
- The compare link may use the previous tag overall, including a pre-release tag, when that is the closest published tag.

## Release Workflow Expectations

- The release workflow should publish the prewritten release note file plus the compare-link changelog.
- The release workflow should fail if the matching `.github/release-notes/<tag>.md` file is missing.
- If release notes are updated after a tag has already been published, the sync workflow should update the existing GitHub release body.

## Privacy Check

- Before pushing public changes, scan the repository for secrets, keys, tokens, embedded credentials, and accidental private data.
- Ignore `.git` metadata when checking source privacy, but explicitly report if source files or docs contain personal emails, secrets, or internal-only endpoints.
