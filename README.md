# Phoenix Lab – Laptop Repair Tracker

Phoenix Lab is a simple web application to track laptop repair tickets for a single office (Newark, DE).

## Quick Start (Docker)

1. Clone the repository:

   ```bash
   git clone https://github.com/chandan-cmd-dev/phoenixlab.git
   cd PhoenixLab
   ```

2. Copy the environment file:

   ```bash
   cp .env.example .env
   ```

   The default `.env.example` is configured for local Docker use with database `phoenixlab`.

3. Start the app:

   ```bash
   docker-compose up --build
   ```

4. Open in your browser:

   - URL: `http://localhost:8080`
   - Default login:
     - Email: `admin@phoenixlab.local`
     - Password: `changeme123`

## Local Development (without Docker)

1. Requirements:

   - Go (version as in `go.mod`)
   - PostgreSQL

2. Database:

   - Create a PostgreSQL database named `phoenixlab`.
   - Ensure `conf/app.conf` points to this database (it already uses `phoenixlab` by default).

3. Run the app:

   ```bash
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
- All migration SQL uses `IF NOT EXISTS`, `ON CONFLICT DO NOTHING`, and `ADD COLUMN IF NOT EXISTS` — they are safe to re-run.

### DANGER ZONE

**NEVER run any of these commands on production:**

```
docker compose down -v        # DELETES ALL DATABASE DATA
docker volume rm pgdata       # DELETES ALL DATABASE DATA
rm -rf pgdata/                # DELETES ALL DATABASE DATA
```

Always use `docker compose down` (without `-v`) if you need to stop services.
