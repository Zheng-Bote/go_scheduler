# Changelog

All notable changes to the MitM Scheduler will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [v0.6.0] - 2026-06-06

### Added
- Scheduler now automatically injects `MITM_DB_*` environment variables (e.g., `MITM_DB_HOST`, `MITM_DB_PORT`, `MITM_DB_USER`, `MITM_DB_PASSWORD`, `MITM_DB_NAME`, `MITM_DB_CONFIG_JSON`) into child processes, providing target database connection details without requiring CLI arguments.

## [v0.5.0] - 2026-06-05

### Added
- Migration `005_change_args_to_jsonb.sql` converting job arguments column in `scheduled_programs` from `TEXT[]` to `JSONB`.
- Integrated JSON validation in `scheduler-admin` arguments input box (Fyne GUI client).
- Support in scheduler daemon to deserialize `JSONB` args and forward them to executed collectors as a serialized JSON string in `os.Args[2]`.

### Changed
- Updated database schema documentation to reflect JSONB representation.
- Updated README documentation with instructions on JSONB overrides.

## [v0.4.0] - 2026-06-04

### Added
- HTTP REST API endpoints to download `system`, `job_status_events`, `job_audit_logs`, and `admin_audit_logs` filtered by date ranges.
- Integrated downloading and local file saving of logs directly from the `scheduler-admin` Fyne GUI.
- Added HELO authentication handshake for remote API connections.

## [v0.1.0] - 2026-06-03

### Added
- Core Scheduler engine supporting Linux cron schedules.
- AES-256-GCM encrypted database connection configurations.
- Unix Domain Socket IPC listener for receiving runtime status/audit notifications from running collectors.
- Database auditing schema including `system_logs`, `job_status_events`, and `job_audit_logs`.
- Multi-platform `scheduler-admin` client built with Fyne.
