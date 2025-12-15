# Repository Guidelines

## Project Structure & Module Organization
- `main.go` boots the CLI/server from `cmd/` and wires configuration flags (single-network and multi-chain modes).
- Core logic lives in `internal/`: `chain/` (keystore handling, transactions, rate limits), `config/` (flag + JSON config parsing), and `server/` (HTTP handlers, CAPTCHA, middleware).
- Frontend is in `web/` (Svelte + Vite). Built assets land in `web/dist/` and are served by the Go binary; source edits should target `web/src/`.
- Sample configs live at `multichain-config.json`; example keystores are under `keystore/`. Docker assets are under `docker/`.

## Build, Test, and Development Commands
- Bundle frontend + embed: `go generate` (runs `npm run build`, which installs yarn, builds `web/`, and refreshes `web/dist/`).
- Build backend: `go build -o multi-chain-faucet` (outputs a self-contained binary).
- Run locally (single network): `./multi-chain-faucet -faucet.name sepolia -provider <rpc> -privatekey <hex>`.
- Run multi-chain: `./multi-chain-faucet -multichain multichain-config.json`.
- Frontend dev server: `cd web && yarn install && yarn dev` (fast iteration before running `go generate`).

## Coding Style & Naming Conventions
- Go: follow `gofmt` defaults (tabs, export comments when needed). Keep functions small and prefer clear error wrapping. Use table-driven tests with descriptive `TestX_Y` names.
- Frontend: Svelte + Bulma; prefer component names in `PascalCase.svelte`. Run `cd web && yarn prettier` for formatting.
- Config files: JSON keys use lowercase with underscores as in `multichain-config.json`.

## Testing Guidelines
- Backend tests: `go test ./...` (uses `testify` assertions and `gomonkey` where mocking is needed). Add unit tests alongside code in `*_test.go`.
- Aim to cover rate limiting, provider selection, and transaction sending logic for new features. For multi-chain changes, test at least one happy path and one rate-limit path per network.
- No frontend test suite exists; lean on manual checks via `yarn dev` and screenshots for UI changes.

## Commit & Pull Request Guidelines
- Commit messages follow a `type(scope): subject` style seen in history (e.g., `eat(S24-5077): update`). Use a ticket/area in scope and a concise verb.
- PRs should include: clear description of behavior change, flags/config touched, test evidence (`go test` output), and screenshots for UI tweaks. Mention any secrets or keys replaced with placeholders.

## Security & Configuration Tips
- Never commit real private keys or hCaptcha secrets; keep samples in `multichain-config.json` placeholder form. Prefer environment injection or local config files ignored by git.
- Set `-proxycount` correctly when running behind reverse proxies to keep rate limiting accurate.
- For mainnet use, double-check payout and interval values; consider setting minimal balances and monitoring logs before opening access.
