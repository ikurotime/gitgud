# gitgud

A small, self-hosted GitHub clone written in Go. It serves real git repositories over
HTTP (clone / push / pull) and provides a web UI for browsing code, opening issues, and
reviewing and merging pull requests.

## Features

- User registration, login, and session auth
- Public and private repositories (private repos are invisible to others — 404, not 403)
- Real git over HTTP via `git http-backend` (clone, push, pull) with HTTP Basic auth
- Code browsing: file tree, syntax-highlighted files, rendered README, branch switcher,
  commit log, and commit diffs (powered by go-git)
- Issues: per-repo numbering, comments, open/close, markdown bodies
- Pull requests: compare two branches, view the diff and commits, comment, and merge
- Flash messages, friendly 404/403/500 pages, and CSRF-protected forms

## Run locally

Requirements: **Go 1.25+** and the **`git`** binary on your `PATH`.

```bash
go run ./cmd/server          # serves http://localhost:8080
# data (SQLite db + bare repos) lands in ./data
```

Configuration is read from the environment (a local `.env` file is loaded automatically):

| Variable             | Example           | Purpose                                 |
| -------------------- | ----------------- | --------------------------------------- |
| `GITGUD_ADDR`        | `:8080`           | Listen address                          |
| `GITGUD_DATA_DIR`    | `./data`          | Where the database and repos are stored |
| `GITGUD_SESSION_KEY` | `a-random-secret` | Session secret (use a random value)     |

Then register a user, create a repository, and follow the on-screen clone instructions:

```bash
git clone http://localhost:8080/<you>/<repo>.git
cd <repo>
echo "# <repo>" > README.md
git add README.md && git commit -m "first commit"
git push -u origin main      # prompts for your gitgud username/password
```

## How it works

gitgud has two git surfaces — a smart-HTTP server (`git http-backend` via CGI) for
clone/push/pull, and a go-git reader for the web browsing UI — sitting on top of a clean,
layered architecture (domain / application / interface / infrastructure). See
[docs/00-overview.md](docs/00-overview.md) for the full design, and the numbered files in
[docs/](docs/) for how each milestone was built.

## Known limitations (v1)

- Git over HTTP only (no SSH)
- Pull requests are between two branches of the **same** repo (no forks)
- Single-node; SQLite storage; bare repos on the local filesystem
- The clone URL shown in the UI is hardcoded to `http://localhost:8080`
