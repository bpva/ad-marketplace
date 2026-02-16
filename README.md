# AD×CHANGE - Telegram Ads Marketplace

[![Go](https://img.shields.io/badge/Go-1.25.6-00ADD8?logo=go&logoColor=white)](go.mod)
[![React](https://img.shields.io/badge/React-19.2.3-61DAFB?logo=react&logoColor=111)](frontend/package.json)
[![TypeScript](https://img.shields.io/badge/TypeScript-5.9.3-3178C6?logo=typescript&logoColor=white)](frontend/package.json)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-16-4169E1?logo=postgresql&logoColor=white)](docker-compose.yml)

## How to run
<details>
<summary>Locally</summary>
dependencies: make + docker (compose)
optional: go and node for running some tooling

1. Create env file:

```bash
cp .env.copy .env
```

2. Fill required values in `.env`:

- `BOT_TOKEN` from [@BotFather](https://t.me/BotFather)
- `TG_API_ID` and `TG_API_HASH` from [my.telegram.org](https://my.telegram.org) (for Telegram channel analytics features)
- `JWT_SECRET` (any random string for local)
- `TG_PHONE`, `TG_PASSWORD` (if 2FA enabled)
then
```bash
make tg-login 
```
it should result in .session.json file in root dir
### finally

```bash
make up
```

Open frontend at [http://localhost:1313](http://localhost:1313)

### Local limitations

- Local mode uses a mocked Telegram user
- `WebApp.initData` is simulated in local mode, so it is not a real Telegram session.
- Wallet connect/linking is limited on `localhost`; validate TON wallet flow on deployed HTTPS build.
- Telegram clients cannot open `localhost` mini apps directly unless it's test environment
</details>

<details>
<summary>Production deployment</summary>

1. Clone this repository and create `.env`:

```bash
cp .env.copy .env
```

2. Set production values in `.env`:

- `ENV=prod`
- `FRONTEND_URL=https://<your-domain>` (reverse proxy to :1313)
- `VITE_API_URL=https://<your-api-url>` (reverse proxy to :8090)
- `BOT_TOKEN`, `TG_API_ID`, `TG_API_HASH`, `JWT_SECRET`, `TG_PHONE`, `TG_PASSWORD`
3. repeat steps from local run with tg-login
```bash
make up
```
</details>

## Demo

## 5-Minute walkthrough

## Architecture Snapshot
