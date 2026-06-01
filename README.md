# Go Scheduler

A Linux command-line scheduler that executes external Go programs based on a PostgreSQL configuration.

## Features

- **Cron Scheduling**: Jobs are scheduled via standard cron expressions.
- **Encrypted Config**: Database credentials and admin tokens are stored in an encrypted JSON file (AES-256-GCM + Argon2id).
- **Dynamic Reloading**: Jobs can be updated via the API and reloaded without restarting the service.
- **IPC over Unix Sockets**: Jobs report status events back to the scheduler via JSON-Lines.
- **Admin API**: Remote job management with authentication.
- **Enhanced Logging**:
  - `system_logs`: Core scheduler lifecycle events.
  - `job_status_events`: Real-time job progress tracking.
  - `job_audit_logs`: Custom audit messages from jobs.
  - `admin_audit_logs`: Audit trail for administrative actions via API.
- **HTTP API**: Health check (`/health`) and info (`/info`) endpoints.

## Project Structure

- `cmd/scheduler`: Main application entry point.
- `cmd/encrypt-config`: Utility to encrypt a JSON configuration file.
- `cmd/job1`, `cmd/job2`: Example jobs demonstrating IPC and Audit features.
- `internal/`: Core logic (crypto, db, ipc, http, scheduler).
- `migrations/`: SQL files for database setup.

## Setup & Build

### 1. Prerequisites

- Go 1.25+
- PostgreSQL Server

### 2. Build

```bash
go build -o ./cmd/scheduler/scheduler ./cmd/scheduler
go build -o ./cmd/encrypt-config/encrypt-config ./cmd/encrypt-config
go build -o ./cmd/job1/job1 ./cmd/job1
go build -o ./cmd/job2/job2 ./cmd/job2
# Linux
go build -o ./cmd/scheduler-admin/scheduler-admin ./cmd/scheduler-admin
# Windows
go build -v -ldflags="-H=windowsgui" -o .\cmd\scheduler-admin\scheduler-admin.exe .\cmd\scheduler-admin\main.go
```

### 3. Database Setup

Apply the migrations in order:

```bash
psql -h <host> -U <user> -d <db> -f migrations/001_init.sql
psql -h <host> -U <user> -d <db> -f migrations/002_logging_and_audit.sql
psql -h <host> -U <user> -d <db> -f migrations/003_admin_and_api.sql
```

### 4. Configuration

Create a `config.json` (see `example_config.json` for a template):

```json
{
  "host": "your-db-host",
  "port": 5432,
  "user": "your-user",
  "password": "your-password",
  "database": "your-dbname",
  "log_level": "DEBUG",
  "admins": [
    {
      "username": "admin1",
      "token": "your_secure_token"
    }
  ]
}
```

Encrypt it:

```bash
./encrypt-config config.json config.json.enc
```

## Running the Scheduler

```bash
./scheduler config.json.enc
```

Decryption password can be provided via prompt or `SCHEDULER_PASSWORD` environment variable.

## Administrative Tools

### 1. GUI Admin Tool (Fyne)

A cross-platform GUI application is available in `cmd/scheduler-admin`. It allows you to:

- Connect to the scheduler via URL and HELO token.
- List all current jobs.
- Create, Edit, and Delete jobs.
- Automatically reload the scheduler after changes.

**Build & Run:**

```bash
cd cmd/scheduler-admin
go build -o scheduler-admin
./scheduler-admin
```

### 2. Update/Create Jobs via curl

```bash
curl -u "admin1:your_secure_token" -X POST -H "Content-Type: application/json" \
     -d '[{
            "name": "RemoteJob",
            "command": "./job1",
            "args": [],
            "cron_expr": "*/2 * * * *",
            "enabled": true,
            "restart_on_exit": false
          }]' \
     http://localhost:8080/admin/update-jobs
```

This request will:

1. Authenticate the user.
2. Upsert the job into the database (by name).
3. Trigger a scheduler reload.
4. Log the action in `admin_audit_logs`.

## IPC & Job Communication

Jobs can send JSON messages to the Unix Domain Socket specified in `SCHEDULER_SOCKET_PATH`.

**Types of messages:**

- **Status (default)**: Updates `job_status_events`.
- **Audit**: Updates `job_audit_logs`.

Example (Go):

```go
client.SendEvent(ipc.StatusEvent{
    RunID:   runID,
    Type:    "audit",
    Message: "Sensitive operation performed",
})
```

## Docker

Build:

```bash
docker build -t go-scheduler .
```

Run:

```bash
docker run -p 8080:8080 -e SCHEDULER_PASSWORD=mypassword go-scheduler ./scheduler /app/config.json.enc
```
