package dashboard

import (
	"context"

	"github.com/jmoiron/sqlx"
)

type DashboardReadRepository interface {
	GetTotals(ctx context.Context, db sqlx.ExtContext) (Totals, error)
	// GetSchoolsInReview returns schools awaiting acceptance, oldest first.
	GetSchoolsInReview(ctx context.Context, db sqlx.ExtContext) ([]SchoolInReview, error)
	// GetIncompleteSchools returns accepted schools with no students or no entries.
	GetIncompleteSchools(ctx context.Context, db sqlx.ExtContext) ([]IncompleteSchool, error)
	GetEventsByStatus(ctx context.Context, db sqlx.ExtContext) ([]StatusCount, error)
	// GetEventFacts returns every event that is neither draft nor finalized.
	GetEventFacts(ctx context.Context, db sqlx.ExtContext) ([]EventFacts, error)
	// GetJudgeScored returns, for open and closed events, each assigned judge's
	// count of entered scores, ordered by judge name.
	GetJudgeScored(ctx context.Context, db sqlx.ExtContext) ([]JudgeScored, error)
}
