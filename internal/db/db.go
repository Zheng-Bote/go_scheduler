package db

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// ScheduledProgram represents a job configuration from the DB
type ScheduledProgram struct {
	ID            int
	Name          string
	Command       string
	Args          []string
	CronExpr      string
	Enabled       bool
	RestartOnExit bool
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// ProgramRun tracks an execution instance of a job
type ProgramRun struct {
	ID         int
	ProgramID  int
	PID        int
	StartedAt  time.Time
	FinishedAt *time.Time
	ExitCode   *int
	Success    *bool
}

// SchedulerConfig holds runtime settings for the scheduler
type SchedulerConfig struct {
	HTTPPort   int
	SocketPath string
	LogLevel   string
}

// Repository handles database operations
type Repository struct {
	Pool *pgxpool.Pool
}

// NewRepository creates a new repository with a connection pool
func NewRepository(ctx context.Context, dsn string) (*Repository, error) {
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to create pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping db: %w", err)
	}

	return &Repository{Pool: pool}, nil
}

// GetSchedulerConfig fetches the scheduler configuration
func (r *Repository) GetSchedulerConfig(ctx context.Context) (*SchedulerConfig, error) {
	var cfg SchedulerConfig
	err := r.Pool.QueryRow(ctx, "SELECT http_port, socket_path, log_level FROM scheduler_config LIMIT 1").Scan(&cfg.HTTPPort, &cfg.SocketPath, &cfg.LogLevel)
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}

// LogSystem records a system-wide log entry
func (r *Repository) LogSystem(ctx context.Context, level, component, message string) {
	_, err := r.Pool.Exec(ctx, "INSERT INTO system_logs (level, component, message) VALUES ($1, $2, $3)", level, component, message)
	if err != nil {
		fmt.Printf("[DB-LOG-ERROR] Failed to write to system_logs: %v (Original Msg: [%s] %s: %s)\n", err, level, component, message)
	}
}

// CreateAuditLog records a job audit entry
func (r *Repository) CreateAuditLog(ctx context.Context, runID int, message string) error {
	_, err := r.Pool.Exec(ctx, "INSERT INTO job_audit_logs (run_id, message) VALUES ($1, $2)", runID, message)
	if err != nil {
		fmt.Printf("[DB-AUDIT-ERROR] Failed to write to job_audit_logs: %v\n", err)
	}
	return err
}

// LogAdminAction records an administrative action
func (r *Repository) LogAdminAction(ctx context.Context, username, action string, details interface{}) {
	detailsJSON, _ := json.Marshal(details)
	_, err := r.Pool.Exec(ctx, "INSERT INTO admin_audit_logs (username, action, details) VALUES ($1, $2, $3)", username, action, detailsJSON)
	if err != nil {
		fmt.Printf("[DB-ADMIN-LOG-ERROR] %v\n", err)
	}
}

// UpsertScheduledProgram inserts or updates a program configuration
func (r *Repository) UpsertScheduledProgram(ctx context.Context, p ScheduledProgram) error {
	query := `
		INSERT INTO scheduled_programs (name, command, args, cron_expr, enabled, restart_on_exit, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, CURRENT_TIMESTAMP)
		ON CONFLICT (name) DO UPDATE SET
			command = EXCLUDED.command,
			args = EXCLUDED.args,
			cron_expr = EXCLUDED.cron_expr,
			enabled = EXCLUDED.enabled,
			restart_on_exit = EXCLUDED.restart_on_exit,
			updated_at = CURRENT_TIMESTAMP`
	
	// Note: You need a UNIQUE constraint on 'name' for ON CONFLICT to work.
	// We'll assume the migration adds it or we use a manual check.
	_, err := r.Pool.Exec(ctx, query, p.Name, p.Command, p.Args, p.CronExpr, p.Enabled, p.RestartOnExit)
	return err
}

// GetAllPrograms fetches all scheduled programs
func (r *Repository) GetAllPrograms(ctx context.Context) ([]ScheduledProgram, error) {
	rows, err := r.Pool.Query(ctx, "SELECT id, name, command, args, cron_expr, enabled, restart_on_exit FROM scheduled_programs ORDER BY name ASC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var programs []ScheduledProgram
	for rows.Next() {
		var p ScheduledProgram
		if err := rows.Scan(&p.ID, &p.Name, &p.Command, &p.Args, &p.CronExpr, &p.Enabled, &p.RestartOnExit); err != nil {
			return nil, err
		}
		programs = append(programs, p)
	}
	return programs, nil
}

// GetEnabledPrograms fetches all enabled scheduled programs
func (r *Repository) GetEnabledPrograms(ctx context.Context) ([]ScheduledProgram, error) {
	rows, err := r.Pool.Query(ctx, "SELECT id, name, command, args, cron_expr, enabled, restart_on_exit FROM scheduled_programs WHERE enabled = true")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var programs []ScheduledProgram
	for rows.Next() {
		var p ScheduledProgram
		if err := rows.Scan(&p.ID, &p.Name, &p.Command, &p.Args, &p.CronExpr, &p.Enabled, &p.RestartOnExit); err != nil {
			return nil, err
		}
		programs = append(programs, p)
	}
	return programs, nil
}

// CreateProgramRun initializes a new program run entry
func (r *Repository) CreateProgramRun(ctx context.Context, programID int, pid int) (int, error) {
	var id int
	err := r.Pool.QueryRow(ctx, "INSERT INTO program_runs (program_id, pid, started_at) VALUES ($1, $2, CURRENT_TIMESTAMP) RETURNING id", programID, pid).Scan(&id)
	return id, err
}

// UpdateProgramRun updates a run entry upon completion
func (r *Repository) UpdateProgramRun(ctx context.Context, runID int, exitCode int, success bool) error {
	_, err := r.Pool.Exec(ctx, "UPDATE program_runs SET finished_at = CURRENT_TIMESTAMP, exit_code = $1, success = $2 WHERE id = $3", exitCode, success, runID)
	return err
}

// CreateStatusEvent logs an IPC status event
func (r *Repository) CreateStatusEvent(ctx context.Context, runID int, status string, message string, progress int) error {
	_, err := r.Pool.Exec(ctx, "INSERT INTO job_status_events (run_id, status, message, progress) VALUES ($1, $2, $3, $4)", runID, status, message, progress)
	return err
}
