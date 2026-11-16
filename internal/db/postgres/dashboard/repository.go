package dashboard

import (
	"context"

	"github.com/jmoiron/sqlx"
	"github.com/nevinmanoj/bhavana-backend/internal/core"
	"github.com/nevinmanoj/bhavana-backend/internal/domain/dashboard"
)

type dashboardRepository struct {
}

func NewDashboardReadRepository() dashboard.DashboardReadRepository {
	return &dashboardRepository{}
}

func (r *dashboardRepository) GetTotals(ctx context.Context, db sqlx.ExtContext) (dashboard.Totals, error) {
	var totals dashboard.Totals
	query := `
		SELECT (SELECT COUNT(*) FROM schools)  AS schools,
		       (SELECT COUNT(*) FROM students) AS students,
		       (SELECT COUNT(*) FROM events)   AS events,
		       (SELECT COUNT(*) FROM entries)  AS entries
	`
	if err := sqlx.GetContext(ctx, db, &totals, query); err != nil {
		return totals, dashboard.ErrInternal
	}
	return totals, nil
}

func (r *dashboardRepository) GetSchoolsInReview(ctx context.Context, db sqlx.ExtContext) ([]dashboard.SchoolInReview, error) {
	schools := []dashboard.SchoolInReview{}
	query := `
		SELECT id, name, created_at
		FROM schools
		WHERE status = $1
		ORDER BY created_at ASC, id ASC
	`
	if err := sqlx.SelectContext(ctx, db, &schools, query, core.SchoolStatusInReview); err != nil {
		return nil, dashboard.ErrInternal
	}
	return schools, nil
}

func (r *dashboardRepository) GetIncompleteSchools(ctx context.Context, db sqlx.ExtContext) ([]dashboard.IncompleteSchool, error) {
	schools := []dashboard.IncompleteSchool{}
	query := `
		SELECT id, name, students, entries
		FROM (
		    SELECT sc.id, sc.name,
		           (SELECT COUNT(*) FROM students st WHERE st.school_id = sc.id) AS students,
		           (SELECT COUNT(*) FROM entries t   WHERE t.school_id = sc.id)  AS entries
		    FROM schools sc
		    WHERE sc.status = $1
		) c
		WHERE students = 0 OR entries = 0
		ORDER BY name ASC, id ASC
	`
	if err := sqlx.SelectContext(ctx, db, &schools, query, core.SchoolStatusAccepted); err != nil {
		return nil, dashboard.ErrInternal
	}
	return schools, nil
}

func (r *dashboardRepository) GetEventsByStatus(ctx context.Context, db sqlx.ExtContext) ([]dashboard.StatusCount, error) {
	counts := []dashboard.StatusCount{}
	query := `SELECT status, COUNT(*) AS count FROM events GROUP BY status`
	if err := sqlx.SelectContext(ctx, db, &counts, query); err != nil {
		return nil, dashboard.ErrInternal
	}
	return counts, nil
}

func (r *dashboardRepository) GetEventFacts(ctx context.Context, db sqlx.ExtContext) ([]dashboard.EventFacts, error) {
	facts := []dashboard.EventFacts{}
	query := `
		SELECT e.id, e.title, e.status,
		       (SELECT COUNT(*) FROM event_judges ej   WHERE ej.event_id = e.id) AS judges,
		       (SELECT COUNT(*) FROM event_criteria ec WHERE ec.event_id = e.id) AS criteria,
		       (SELECT COUNT(*) FROM entries t         WHERE t.event_id = e.id)  AS entries,
		       EXISTS (SELECT 1 FROM event_standings es WHERE es.event_id = e.id) AS has_standings
		FROM events e
		WHERE e.status NOT IN ($1, $2)
		ORDER BY e.title ASC, e.id ASC
	`
	err := sqlx.SelectContext(ctx, db, &facts, query, core.EventStatusDraft, core.EventStatusFinalized)
	if err != nil {
		return nil, dashboard.ErrInternal
	}
	return facts, nil
}

// GetJudgeScored is the all-events counterpart of result.GetJudgeGaps: it
// counts each assigned judge's scores against the event's current entries and
// criteria, leaving the "how many are missing" arithmetic to the service.
func (r *dashboardRepository) GetJudgeScored(ctx context.Context, db sqlx.ExtContext) ([]dashboard.JudgeScored, error) {
	rows := []dashboard.JudgeScored{}
	query := `
		SELECT ej.event_id, u.id, u.name, COUNT(s.id) AS scored
		FROM event_judges ej
		JOIN events e ON e.id = ej.event_id
		JOIN users u  ON u.id = ej.user_id
		LEFT JOIN scores s ON s.judge_id = ej.user_id
		    AND s.entry_id    IN (SELECT id FROM entries        WHERE event_id = ej.event_id)
		    AND s.criteria_id IN (SELECT id FROM event_criteria WHERE event_id = ej.event_id)
		WHERE e.status IN ($1, $2)
		GROUP BY ej.event_id, u.id, u.name
		ORDER BY u.name ASC, u.id ASC
	`
	err := sqlx.SelectContext(ctx, db, &rows, query, core.EventStatusOpen, core.EventStatusClosed)
	if err != nil {
		return nil, dashboard.ErrInternal
	}
	return rows, nil
}
