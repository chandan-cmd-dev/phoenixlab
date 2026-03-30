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
