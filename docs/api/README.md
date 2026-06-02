# Go Scheduler - REST API Documentation

This document describes the REST API endpoints provided by the Go Scheduler server.

The server's HTTP engine is implemented in [internal/http/server.go](file:///home/zb_bamboo/DEV/__NEW__/Go/go_scheduler/internal/http/server.go). The port is dynamically loaded from the database configuration.

---

## 1. API Endpoints Table

| URL | Method | Description | Options / Parameters | Authentication / Role |
| :--- | :--- | :--- | :--- | :--- |
| `/health` | `GET` | Health check (verifies database ping) | None | Public (Unauthenticated) |
| `/info` | `GET` | Service descriptor, name, version | None | Public (Unauthenticated) |
| `/admin/jobs` | `GET` | List all scheduled job configurations | None | Admin (HTTP Basic Auth) |
| `/admin/update-jobs`| `POST` | Create or update one or more job configurations | Request Body: JSON array of job configs | Admin (HTTP Basic Auth) |
| `/admin/delete-job` | `DELETE`| Remove a job configuration and reload scheduler | `name` (Query parameter, e.g. `?name=RemoteJob`) | Admin (HTTP Basic Auth) |
| `/admin/logs/system`| `GET` | Download system logs as a JSON file | `from` (Query parameter, optional)<br>`to` (Query parameter, optional) | Admin (HTTP Basic Auth) |
| `/admin/logs/job-audit`| `GET` | Download job audit logs as a JSON file | `from` (Query parameter, optional)<br>`to` (Query parameter, optional) | Admin (HTTP Basic Auth) |
| `/admin/logs/admin-audit`| `GET`| Download administrative audit logs as JSON | `from` (Query parameter, optional)<br>`to` (Query parameter, optional) | Admin (HTTP Basic Auth) |

---

## 2. Endpoints Detail

### 2.1 Health Check
*   **Path**: `/health`
*   **Method**: `GET`
*   **Auth**: Unauthenticated
*   **Handler**: [Server.handleHealth](file:///home/zb_bamboo/DEV/__NEW__/Go/go_scheduler/internal/http/server.go)
*   **Response**:
    *   `200 OK` (Body: `OK`)
    *   `500 Internal Server Error` (Body: `DB Error: <error_message>`)

### 2.2 Info
*   **Path**: `/info`
*   **Method**: `GET`
*   **Auth**: Unauthenticated
*   **Handler**: [Server.handleInfo](file:///home/zb_bamboo/DEV/__NEW__/Go/go_scheduler/internal/http/server.go)
*   **Response**:
    *   `200 OK` (Content-Type: `application/json`)
    ```json
    {
      "name": "Go Scheduler",
      "description": "A Linux commandline scheduler",
      "version": "1.0.0"
    }
    ```

### 2.3 List Jobs
*   **Path**: `/admin/jobs`
*   **Method**: `GET`
*   **Auth**: Admin (HTTP Basic Auth)
*   **Handler**: [Server.handleGetJobs](file:///home/zb_bamboo/DEV/__NEW__/Go/go_scheduler/internal/http/server.go)
*   **Response**:
    *   `200 OK` (Content-Type: `application/json`)
    ```json
    [
      {
        "id": 1,
        "name": "Job1",
        "command": "./bin/job1",
        "args": [],
        "cron_expr": "*/5 * * * *",
        "enabled": true,
        "restart_on_exit": false
      }
    ]
    ```

### 2.4 Create / Update Jobs
*   **Path**: `/admin/update-jobs`
*   **Method**: `POST`
*   **Auth**: Admin (HTTP Basic Auth)
*   **Handler**: [Server.handleUpdateJobs](file:///home/zb_bamboo/DEV/__NEW__/Go/go_scheduler/internal/http/server.go)
*   **Request Body**: JSON array of job structures.
    ```json
    [
      {
        "name": "RemoteJob",
        "command": "./job1",
        "args": [],
        "cron_expr": "*/2 * * * *",
        "enabled": true,
        "restart_on_exit": false
      }
    ]
    ```
*   **Response**:
    *   `200 OK` (Body: `Jobs updated and scheduler reloaded`)
    *   `400 Bad Request` (Invalid JSON structure)
    *   `500 Internal Server Error` (Database failure or failed to reload daemon)

### 2.5 Delete Job
*   **Path**: `/admin/delete-job`
*   **Method**: `DELETE`
*   **Auth**: Admin (HTTP Basic Auth)
*   **Handler**: [Server.handleDeleteJob](file:///home/zb_bamboo/DEV/__NEW__/Go/go_scheduler/internal/http/server.go)
*   **Parameters**: `name` (Query parameter) - exact unique name of the program configuration.
    *   Example: `/admin/delete-job?name=RemoteJob`
*   **Response**:
    *   `200 OK` (Body: `Job deleted`)
    *   `400 Bad Request` (Missing `name` parameter)
    *   `500 Internal Server Error` (Database delete query failed)

### 2.6 Download System Logs
*   **Path**: `/admin/logs/system`
*   **Method**: `GET`
*   **Auth**: Admin (HTTP Basic Auth)
*   **Handler**: [Server.handleDownloadSystemLogs](file:///home/zb_bamboo/DEV/__NEW__/Go/go_scheduler/internal/http/server.go)
*   **Parameters** (Optional):
    *   `from`: Query parameter (RFC3339 timestamp or YYYY-MM-DD date)
    *   `to`: Query parameter (RFC3339 timestamp or YYYY-MM-DD date)
*   **Response Headers**:
    *   `Content-Disposition`: `attachment; filename=system_logs.json`
    *   `Content-Type`: `application/json`
*   **Response Body**: Array of system log objects.
    ```json
    [
      {
        "id": 15,
        "level": "INFO",
        "component": "Scheduler",
        "message": "Scheduler started successfully",
        "ts": "2026-06-02T10:00:00Z"
      }
    ]
    ```

### 2.7 Download Job Audit Logs
*   **Path**: `/admin/logs/job-audit`
*   **Method**: `GET`
*   **Auth**: Admin (HTTP Basic Auth)
*   **Handler**: [Server.handleDownloadJobAuditLogs](file:///home/zb_bamboo/DEV/__NEW__/Go/go_scheduler/internal/http/server.go)
*   **Parameters** (Optional):
    *   `from`: Query parameter (RFC3339 timestamp or YYYY-MM-DD date)
    *   `to`: Query parameter (RFC3339 timestamp or YYYY-MM-DD date)
*   **Response Headers**:
    *   `Content-Disposition`: `attachment; filename=job_audit_logs.json`
    *   `Content-Type`: `application/json`
*   **Response Body**: Array of job audit objects.
    ```json
    [
      {
        "id": 2,
        "run_id": 42,
        "message": "Job 1 is performing a security-sensitive operation",
        "ts": "2026-06-02T10:01:05Z"
      }
    ]
    ```

### 2.8 Download Admin Audit Logs
*   **Path**: `/admin/logs/admin-audit`
*   **Method**: `GET`
*   **Auth**: Admin (HTTP Basic Auth)
*   **Handler**: [Server.handleDownloadAdminAuditLogs](file:///home/zb_bamboo/DEV/__NEW__/Go/go_scheduler/internal/http/server.go)
*   **Parameters** (Optional):
    *   `from`: Query parameter (RFC3339 timestamp or YYYY-MM-DD date)
    *   `to`: Query parameter (RFC3339 timestamp or YYYY-MM-DD date)
*   **Response Headers**:
    *   `Content-Disposition`: `attachment; filename=admin_audit_logs.json`
    *   `Content-Type`: `application/json`
*   **Response Body**: Array of administrative audit entries.
    ```json
    [
      {
        "id": 5,
        "username": "admin1",
        "action": "update_jobs_success",
        "details": {
          "count": 1
        },
        "ts": "2026-06-02T10:15:30Z"
      }
    ]
    ```
