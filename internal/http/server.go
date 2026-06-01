package http

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"go-scheduler/internal/config"
	"go-scheduler/internal/db"
)

// SchedulerInterface allows the HTTP server to trigger reloads
type SchedulerInterface interface {
	Reload(ctx context.Context) error
}

// Server handles health, info, and admin HTTP requests
type Server struct {
	Repo      *db.Repository
	Port      int
	Admins    []config.AdminUser
	Scheduler SchedulerInterface
}

// Start runs the HTTP server
func (s *Server) Start() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/info", s.handleInfo)
	mux.HandleFunc("/health", s.handleHealth)
	mux.HandleFunc("/admin/update-jobs", s.handleUpdateJobs)
	mux.HandleFunc("/admin/jobs", s.handleGetJobs)
	mux.HandleFunc("/admin/delete-job", s.handleDeleteJob)

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", s.Port),
		Handler: mux,
	}

	go func() {
		_ = srv.ListenAndServe()
	}()

	return nil
}

func (s *Server) authenticate(r *http.Request) (string, bool) {
	user, token, ok := r.BasicAuth()
	if !ok {
		return "", false
	}

	for _, admin := range s.Admins {
		if admin.Username == user && admin.Token == token {
			return user, true
		}
	}
	return "", false
}

func (s *Server) handleUpdateJobs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	username, ok := s.authenticate(r)
	if !ok {
		w.Header().Set("WWW-Authenticate", `Basic realm="Admin API"`)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var jobs []db.ScheduledProgram
	if err := json.NewDecoder(r.Body).Decode(&jobs); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	ctx := context.Background()
	for _, job := range jobs {
		if err := s.Repo.UpsertScheduledProgram(ctx, job); err != nil {
			s.Repo.LogAdminAction(ctx, username, "update_job_fail", map[string]interface{}{"job": job.Name, "error": err.Error()})
			http.Error(w, fmt.Sprintf("Failed to update job %s: %v", job.Name, err), http.StatusInternalServerError)
			return
		}
	}

	if err := s.Scheduler.Reload(ctx); err != nil {
		s.Repo.LogAdminAction(ctx, username, "reload_fail", err.Error())
		http.Error(w, "Failed to reload scheduler", http.StatusInternalServerError)
		return
	}

	s.Repo.LogAdminAction(ctx, username, "update_jobs_success", map[string]interface{}{"count": len(jobs)})
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Jobs updated and scheduler reloaded"))
}

func (s *Server) handleGetJobs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	username, ok := s.authenticate(r)
	if !ok {
		w.Header().Set("WWW-Authenticate", `Basic realm="Admin API"`)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	jobs, err := s.Repo.GetAllPrograms(context.Background())
	if err != nil {
		http.Error(w, "Failed to fetch jobs", http.StatusInternalServerError)
		return
	}

	s.Repo.LogAdminAction(context.Background(), username, "get_jobs", nil)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(jobs)
}

func (s *Server) handleDeleteJob(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	username, ok := s.authenticate(r)
	if !ok {
		w.Header().Set("WWW-Authenticate", `Basic realm="Admin API"`)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	name := r.URL.Query().Get("name")
	if name == "" {
		http.Error(w, "Missing job name", http.StatusBadRequest)
		return
	}

	ctx := context.Background()
	_, err := s.Repo.Pool.Exec(ctx, "DELETE FROM scheduled_programs WHERE name = $1", name)
	if err != nil {
		http.Error(w, "Failed to delete job", http.StatusInternalServerError)
		return
	}

	_ = s.Scheduler.Reload(ctx)
	s.Repo.LogAdminAction(ctx, username, "delete_job", map[string]string{"name": name})
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Job deleted"))
}

func (s *Server) handleInfo(w http.ResponseWriter, r *http.Request) {
	info := map[string]string{
		"name":        "Go Scheduler",
		"description": "A Linux commandline scheduler",
		"version":     "1.0.0",
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(info)
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := s.Repo.Pool.Ping(ctx); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "DB Error: %v", err)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}
