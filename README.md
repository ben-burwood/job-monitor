# Job Monitor

Dashboard for Jobs - aggregated dashboard across GitHub Actions and AWS Glue, showing the latest run of each tracked job.

Go backend serving a Vue (Vite + Tailwind/daisyUI) frontend, deployed as a single Docker image.

## Environment and Configuration

1. `GITHUB_TOKEN` — PAT with `Actions: read` (and `repo` for private targets).
2. `AWS_REGION` — region the Glue jobs run in (default `eu-west-2`). Credentials come from the default AWS chain (env, `~/.aws`, instance role).
3. `AWS_GLUE_OBSERVE_ROLE_ARN` — optional role to assume for Glue read access (auto-refreshed).
4. `JOB_MONITOR_CONFIG_PATH` — path to the config file (default `config/config.yaml`).

`config/config.yaml` (a sample is included; the `config/` directory is bind-mounted into the container)

- `github[].repo`: `owner/name` — the latest run on the repo's default branch is shown
- `github[].workflow`: workflow filename in `.github/workflows/`
- `glue[].name`: Glue Job name in the configured region

Statuses are cached for 60 seconds; the refresh button in the header bypasses the cache.

## API

- `GET /api/jobs-sse` — Server Sent Events stream, one event per job as it resolves, then a `done` event
- `GET /api/jobs/history?id=<job id>` — JSON run history of a single job over the last 7 days, newest first

Both accept `?refresh=true` to bypass the cache.

## Development

```sh
# Backend (http://localhost:8080, serves ./frontend/dist)
GITHUB_TOKEN=... go run .

# Frontend
cd frontend && npm install && npm run build   # or `npm run dev` (proxies /api to :8080)
```

## Docker

```sh
docker compose up --build
```
