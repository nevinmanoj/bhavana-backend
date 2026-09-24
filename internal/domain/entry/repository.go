package entry

import (
	"context"

	"github.com/jmoiron/sqlx"
)

type EntryWriteRepository interface {
	EntryReadRepository
	CreateEntry(ctx context.Context, db sqlx.ExtContext, entryToCreate *Entry) error
	DeleteEntry(ctx context.Context, db sqlx.ExtContext, entryID int64) error

	CreateEntryMember(ctx context.Context, db sqlx.ExtContext, entryMemberToCreate *EntryMember) error
	DeleteEntryMember(ctx context.Context, db sqlx.ExtContext, entryID, studentId int64) error
}
type EntryReadRepository interface {
	GetAllEntries(ctx context.Context, db sqlx.ExtContext, filter EntryFilter) ([]EntryFull, error)
	// CountEntries returns the total row count for filter, ignoring
	// filter.Limit/Offset - the total a paginated list needs alongside its page.
	CountEntries(ctx context.Context, db sqlx.ExtContext, filter EntryFilter) (int64, error)
	GetEntryByID(ctx context.Context, db sqlx.ExtContext, entryId int64) (*EntryFull, error)

	GetEntryMembers(ctx context.Context, db sqlx.ExtContext, entryId int64) ([]EntryMember, error)
}
