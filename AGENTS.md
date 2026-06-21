# World of Hyatt CLI Agent Guide

This repository contains the standalone World of Hyatt award availability CLI and MCP server. Keep local edits narrow, verify behavior before shipping, and avoid unrelated cleanup.

## Local Operating Contract

Start by asking the CLI for current runtime truth:

```bash
hyatt-cli doctor --json
hyatt-cli agent-context --pretty
```

Use runtime discovery instead of relying on a copied command list:

```bash
hyatt-cli which "<capability>" --json
hyatt-cli <command> --help
```

Add `--agent` to command invocations for JSON, compact output, non-interactive defaults, no color, and confirmation-safe scripting:

```bash
hyatt-cli <command> --agent
```

Before running an unfamiliar command that may mutate remote state, inspect its help and prefer a dry run:

```bash
hyatt-cli <command> --help
hyatt-cli <command> --dry-run --agent
```

Use `--yes --no-input` only after the target, arguments, and side effects are clear.

For install, auth, examples, and longer product guidance, read `README.md` and `SKILL.md`. This file intentionally stays small so repo-local agents get invariant local guidance without duplicating the user docs.

## Release Notes

Use `CHANGELOG.md` for user-facing release notes when behavior changes. Do not bump versions or create release tags unless the user explicitly asks.
