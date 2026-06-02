package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"
)

// StatsHandler provides aggregated statistics across all projects.
type StatsHandler struct {
	db *gorm.DB
}

func NewStatsHandler(db *gorm.DB) *StatsHandler {
	return &StatsHandler{db: db}
}

func (h *StatsHandler) RegisterRoutes(r chi.Router) {
	r.Get("/stats/overview", h.overview)
}

// OverviewResponse is the aggregated health view for a user's orgs/projects.
type OverviewResponse struct {
	TotalIssues     int64            `json:"total_issues"`
	UnresolvedCount int64            `json:"unresolved_count"`
	ResolvedCount   int64            `json:"resolved_count"`
	IgnoredCount    int64            `json:"ignored_count"`
	Recent24h       int64            `json:"recent_24h"`
	TopProjects     []ProjectSummary `json:"top_projects"`
}

type ProjectSummary struct {
	ProjectID   string `json:"project_id"`
	ProjectName string `json:"project_name"`
	IssueCount  int64  `json:"issue_count"`
}

func (h *StatsHandler) overview(w http.ResponseWriter, r *http.Request) {
	var resp OverviewResponse

	now := time.Now().UTC()
	since := now.Add(-24 * time.Hour)

	// Total issues by status — single query
	type statusRow struct {
		Status string
		Count  int64
	}
	var rows []statusRow
	h.db.Raw(`
		SELECT status, COUNT(*) as count
		FROM issues
		GROUP BY status
	`).Scan(&rows)

	for _, row := range rows {
		resp.TotalIssues += row.Count
		switch row.Status {
		case "unresolved":
			resp.UnresolvedCount = row.Count
		case "resolved":
			resp.ResolvedCount = row.Count
		case "ignored":
			resp.IgnoredCount = row.Count
		}
	}

	// Recent 24h
	h.db.Raw(`
		SELECT COUNT(*)
		FROM events
		WHERE created_at >= ?
	`, since).Scan(&resp.Recent24h)

	// Top projects by issue count
	h.db.Raw(`
		SELECT p.id AS project_id, p.name AS project_name, COUNT(i.id) AS issue_count
		FROM issues i
		JOIN projects p ON p.id = i.project_id
		GROUP BY p.id, p.name
		ORDER BY issue_count DESC
		LIMIT 5
	`).Scan(&resp.TopProjects)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}
