# Git Cache & go-git Migration

## Overview

Scribe uses a local clone cache and the pure-Go [go-git](https://github.com/go-git/go-git) library for all git operations. This eliminates the need for the `git` binary to be installed on the system and significantly speeds up repeated operations like `scribe check` and `scribe update`.

## How It Works

### Cache Location

Cloned repositories are cached at `~/.scribe/cache/`, organized by host and owner:

```
~/.scribe/cache/
├── github.com/
│   ├── vercel-labs/agent-skills/
│   └── owner/repo/
├── gitlab.com/
│   └── org/repo/
└── bitbucket.org/
    └── team/repo/
```

### Cache Behavior

| Operation | First Run | Subsequent Runs |
|-----------|-----------|-----------------|
| `scribe install owner/repo` | Shallow clone (depth=1) into cache | Fetch updates, then install |
| `scribe check` | Shallow clone into cache | Fetch updates, compare hashes |
| `scribe update` | Shallow clone into cache | Fetch updates, copy new content |

- **First access**: The repository is cloned with `depth=1` into `~/.scribe/cache/<host>/<owner>/<repo>/`.
- **Subsequent access**: Scribe opens the cached clone and fetches the latest changes. If a specific branch or tag is requested, it checks out that ref.
- **Corrupted cache**: If the cached clone can't be opened or fetched, Scribe deletes it and performs a fresh clone. This is logged at warn level.

### What Gets Cached

| Source Type | Cached | Reason |
|-------------|--------|--------|
| GitHub (`owner/repo`) | Yes | Keyed by `github.com/owner/repo` |
| GitLab (`gitlab.com/owner/repo`) | Yes | Keyed by `gitlab.com/owner/repo` |
| Bitbucket (`bitbucket.org/owner/repo`) | Yes | Keyed by `bitbucket.org/owner/repo` |
| Local path (`./path`) | No | Already on disk |
| Zip URL (`https://example.com/skills.zip`) | No | No natural cache key |

### Branch/Tag Handling

One cache entry is stored per repository, not per branch. When a ref is specified (e.g., `owner/repo#v2.0`), Scribe fetches the latest and resets the worktree to that ref. Refs are tried as a branch first, then as a tag.

## CLI Commands

### `scribe cache path`

Print the cache directory path.

```bash
scribe cache path
# /Users/you/.scribe/cache
```

### `scribe cache clear`

Remove all cached repositories.

```bash
scribe cache clear
# Cache cleared
```

## Performance Impact

Before this change, every `scribe check`, `scribe update`, and `scribe install` performed a fresh `git clone --depth 1` via the system `git` binary into a temporary directory, then deleted it immediately after use. With 5 installed skills from 3 different repos, `scribe check` would perform 5 separate clones.

Now:

- **`scribe check`** with 5 skills from 3 repos performs 3 fetches (one per repo, cached).
- **`scribe update`** reuses the cache populated by its internal check step — no double-cloning.
- **`scribe install owner/repo`** populates the cache on first run, making future checks and updates near-instant for that repo.

## Authentication

Scribe resolves credentials automatically for private repositories:

- **HTTPS sources:** Scribe calls the system `git credential fill` command, which queries whatever credential helper the user has configured (`gh auth login`, macOS Keychain, Windows Credential Manager, etc.). This requires the `git` binary to be installed.
- **SSH sources** (`git@host:owner/repo.git`): Scribe connects through the user's SSH agent using go-git's SSH transport.

Authentication is resolved per clone/fetch operation and is never stored by Scribe. For public repositories, no credentials are needed and the behavior is unchanged.

## Git Binary

All core git operations (clone, fetch, reset) use [go-git/v5](https://github.com/go-git/go-git), a pure Go implementation. The system `git` binary is only used for credential resolution (`git credential fill`) when accessing private HTTPS repositories. If `git` is not installed, public repositories still work and SSH authentication works via the SSH agent.

## Storage

The cache directory is created automatically alongside other Scribe directories during `EnsureScribeDirs()`. The full storage layout:

```
~/.scribe/
├── scrolls/        # Installed skills (canonical storage)
├── workspaces/     # Workspace definitions
├── cache/          # Cached git clones
└── config.json     # Global configuration
```

## Troubleshooting

### Cache takes up too much disk space

Clear it at any time — Scribe will re-clone on the next operation:

```bash
scribe cache clear
```

### Stale cache or unexpected behavior

If skills seem outdated after an install or update, clear the cache and retry:

```bash
scribe cache clear
scribe update --force
```
