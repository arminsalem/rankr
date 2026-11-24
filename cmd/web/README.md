# Web scaffold

Part of issue #218. Currently provides an HTTP server, health check,
configuration loading, and graceful shutdown.

An opt-in leaderboard preview renders sample, empty, and error states.
An opt-in leaderboard page fetches data from the leaderboardstat HTTP API.

## Run locally

From the repository root:

```bash
RANKR_FEATURE_WEB_ENABLED=on go run ./cmd/web serve
```

Health endpoint:

http://localhost:8085/v1/health-check

Expected response:

```json
{ "status": "ok" }
```

Press Ctrl+C to stop the server.

## Leaderboard preview

Run from the repository root:

```bash
go run ./cmd/web serve --preview
```

Open http://localhost:8085/preview.

Use the page links to inspect the sample table, empty state, and error state.
The preview uses sample data and is disabled by default.

Templates are embedded in the binary. Restart the web process after editing them.

## API-backed leaderboard

Start leaderboardstat and its dependencies, then run:

```bash
RANKR_FEATURE_WEB_ENABLED=on RANKR_FEATURE_WEB_LEADERBOARD=on go run ./cmd/web serve
```

Open http://localhost:8085/leaderboard?project_id=1028435569.

`project_id` is the numeric GitHub repository ID.
The page displays up to 10 rows, or an empty or error state.

Configure `leaderboard_api.base_url` and `leaderboard_api.timeout`
in YAML. Environment overrides are also supported:

```bash
RANKR_FEATURE_WEB_ENABLED=on RANKR_FEATURE_WEB_LEADERBOARD=on \
  WEB__LEADERBOARD_API__BASE_URL=http://127.0.0.1:8081 go run ./cmd/web serve
```

The refresh button updates the leaderboard using htmx.
htmx 2.0.11 is loaded from jsDelivr with an integrity check.

## Project page

For local testing, add `/v1/projects/public/:id` to the project service's
`http_server.public_paths` and start project and leaderboardstat. With a
public GitHub repository linked to a project, start the web service:

```bash
RANKR_FEATURE_WEB_ENABLED=on RANKR_FEATURE_WEB_PROJECT=on go run ./cmd/web serve
```

Open `http://localhost:8085/projects/<project-uuid>` using the Rankr project ID.
The page shows the repository owner, provider, project status, daily
contributor count, and top 10 daily scores. Configure the project service
URL and timeout under `project_api`. The page is off by default.

## Configuration

The default file is deploy/web/development/config.yml.
Set CONFIG_PATH to use another YAML file.

The [shared flag registry](../../docs/flags.md) reads web flags from `features:`.
Web flags default to `off`. Set these to `on` to start the server and expose
the leaderboard:

```yaml
features:
  web_enabled: on
  web_leaderboard: on
```

`RANKR_FEATURE_WEB_ENABLED` and `RANKR_FEATURE_WEB_LEADERBOARD` override YAML.
For local use, `--preview` and `--leaderboard` supply `on` defaults for the
flags they need. Explicit feature settings take precedence over these defaults.

Environment variables override YAML values. For example, in Git Bash:

```bash
RANKR_FEATURE_WEB_ENABLED=on WEB__HTTP_SERVER__PORT=8086 go run ./cmd/web serve
```

The health endpoint then uses port 8086.

## Checks

```bash
make check
```

Without Make, in Git Bash:

```bash
CGO_ENABLED=1 bash scripts/check.sh
```

Race tests require a working C compiler on PATH.
