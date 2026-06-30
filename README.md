# Phoenix Lab – Laptop Repair Tracker

Phoenix Lab is a web application to track laptop repair tickets for a single office.

## Features

- Ticket lifecycle management — statuses, assignments, parts, costs, RMA/DOA workflows, and ticket linking
- Role-based access (`super_admin`, `admin`, `technician`, `viewer`) with per-branch scoping
- Excel import for bulk ticket creation
- Two-way **Google Sheets sync** — OAuth connect, configurable column mapping, identity-based row linking, conflict & adoption review, and scheduled auto-sync
- Audit log, analytics dashboard, and multi-language UI (English, 中文, Español)

## Requirements

- Docker and Docker Compose, **or**
- Go 1.25+ and PostgreSQL 16 for local development

## Quick Start (Docker)

1. Clone the repository:

   ```bash
   git clone https://github.com/chandan-cmd-dev/phoenixlab.git
   cd phoenixlab
   ```

2. Create your environment file:

   ```bash
   cp .env.example .env
   ```

   `.env` holds the Google OAuth credentials used by the Google Sheets feature:

   - `GOOGLE_OAUTH_CLIENT_ID`
   - `GOOGLE_OAUTH_CLIENT_SECRET`
   - `GOOGLE_OAUTH_REDIRECT_URL` (default `http://localhost:8080/oauth/google/callback`)

   `docker-compose` loads `.env` automatically via `env_file`. The file is gitignored, so your secrets are never committed. If you don't need Google Sheets sync, the app runs fine without these values set.

3. Start the app:

   ```bash
   docker-compose up --build
   ```

4. Open in your browser:

   - URL: `http://localhost:8080`
   - Default login:
     - Email: `admin@phoenixlab.local`
     - Password: `changeme123`

## Google Sheets Integration

Phoenix Lab syncs tickets two-way with brand-specific Google Sheet tabs.

1. Set the `GOOGLE_OAUTH_*` values in `.env` (see above).
2. In your Google Cloud Console OAuth client, register the redirect URL `http://localhost:8080/oauth/google/callback` (or whatever you set `GOOGLE_OAUTH_REDIRECT_URL` to).
3. In the app, open **Google Sheets** in the sidebar → **Connect Google** → authorize the account.
4. **Connect a Spreadsheet**, confirm the auto-suggested column mapping, then **Import** (pull), **Push**, or run a **two-way reconcile**.
5. Review **conflicts** (both sides changed a field) and **adoptions** (one row matched several tickets), and optionally enable per-connection **auto-sync**.

The sync is identity-based (not row-position based), so rows can move around the sheet without breaking links. Connections, mappings, row links, conflicts, and adoptions are all persisted in the database.

## Local Development (without Docker)

1. Requirements:

   - Go 1.25+
   - PostgreSQL 16

2. Database:

   - Create a PostgreSQL database named `phoenixlab`.
   - Ensure `conf/app.conf` points to this database (it already uses `phoenixlab` by default).

3. Load environment variables (only required for Google Sheets sync) and run the app:

   ```bash
   set -a; source .env; set +a
   go run cmd/main.go
   ```

Then open `http://localhost:8080` and log in with the default credentials above.

## Production Deployment

### Safe Deploy (recommended)

Always use the deploy script to push code changes to production:

```bash
./scripts/deploy.sh
```

This script will:
1. **Backup the database** before any changes
2. Rebuild only the app container (DB is never touched)
3. Restart the app (DB stays running with all data intact)
4. Verify the app is healthy
5. Show migration results

### Manual Backup

To create a manual database backup at any time:

```bash
./scripts/backup.sh
```

Backups are saved to `./backups/` with timestamps. The last 30 are retained.

### Restore from Backup

```bash
cat backups/phoenixlab_YYYYMMDD_HHMMSS.sql | docker compose exec -T db psql -U tracker -d phoenixlab
```

### How Migrations Work

- Migrations are in `migrations/*.sql` and run automatically on app startup.
- The `schema_migrations` table tracks which migrations have been applied.
- On an existing production DB, the runner automatically detects pre-existing tables and skips old migrations — **no data is ever deleted**.
- Migrations run before the model schema sync, so a brand-new database is seeded correctly (admin user, branches) while existing data is left untouched.
- All migration SQL uses `IF NOT EXISTS`, `ON CONFLICT DO NOTHING`, and `ADD COLUMN IF NOT EXISTS` — they are safe to re-run.

### DANGER ZONE

**NEVER run any of these commands on production:**

```
docker compose down -v        DELETES ALL DATABASE DATA
docker volume rm pgdata       DELETES ALL DATABASE DATA
rm -rf pgdata/                DELETES ALL DATABASE DATA
```

Always use `docker compose down` (without `-v`) if you need to stop services.
