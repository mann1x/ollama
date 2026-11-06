# CLAUDE.md

See `AGENTS.md` for the shared agent instructions for this repository.

## Before touching a carried patch

This fork's patches are consumed by **xollama**, which merges them by sha from a
manifest. How that works — the `up-<slug>` branch naming, `PATCHES.json` and its
apply order, the rebase-onto-the-upstream-tag rule, and why `main` is a plain
mirror — is written down once, canonically, at:

    /srv/dev-disk-by-label-opt/dev/xollama/docs/protocols/FORK-SYNC.md
    (mann1x/xollama@dev, commit ee5331f7)

Read it before authoring a patch, renaming or rebasing an `up-*` branch, or
cutting a release. Two rules from it are easy to break by accident:

- **Never hand-copy a hunk between this fork and xollama.** A fix authored on
  the far side of a file both repos own is how `model/parsers/gemma4.go` came to
  differ in three places without anyone noticing.
- **`patches[]` order in `PATCHES.json` is apply order.** xollama merges each
  entry as its own `--no-ff` merge; a wrong order is a silent conflict
  resolution rather than an error.

Releases are cut from `think-budget`, never from `thinkbudget-<ver>` — see
`docs/protocols/` in the Cerebriline tree for the build/deploy half.
