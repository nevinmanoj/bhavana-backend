package user

import (
	"context"
	"database/sql"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	auth "github.com/nevinmanoj/bhavana-backend/internal/auth"
	"github.com/nevinmanoj/bhavana-backend/internal/core"
	user "github.com/nevinmanoj/bhavana-backend/internal/domain/user"
)

type userRepository struct {
}

func NewUserWriteRepository() user.UserWriteRepository {
	return &userRepository{}
}
func NewUserReadRepository() user.UserReadRepository {
	return &userRepository{}
}

func (r *userRepository) CreateUser(ctx context.Context, db sqlx.ExtContext, password string, userToCreate *user.User) error {
	//check if email already exists
	var exists bool
	err := db.QueryRowxContext(
		ctx,
		`SELECT EXISTS (
        SELECT 1
        FROM users
        WHERE email = $1
    )`,
		userToCreate.Email,
	).Scan(&exists)

	if err != nil {
		return user.ErrInternal
	}

	if exists {
		return user.ErrAlreadyExists
	}
	// "" means "no password yet": only RegisterSchool passes it, and the
	// account can't sign in until acceptance emails it a generated one.
	// CreateUserRequest requires min=6, so the Users page can't create one.
	userToCreate.PasswordHash = ""
	if password != "" {
		passwordHash, err := auth.HashPassword(password)
		if err != nil {
			return user.ErrInternal
		}
		userToCreate.PasswordHash = passwordHash
	}

	query := `
		INSERT INTO users (
			name,
			email,
			password_hash,
			role
		)
		VALUES (
			:name,
			:email,
			:password_hash,
			:role
		)
		RETURNING id, created_at
	`

	rows, err := sqlx.NamedQueryContext(ctx, db, query, userToCreate)
	if err != nil {
		return user.ErrInternal
	}
	defer rows.Close()

	if rows.Next() {
		rows.Scan(&userToCreate.ID, &userToCreate.CreatedAt)
		return nil
	}

	return user.ErrInternal
}

func (r *userRepository) GetUserByEmail(ctx context.Context, db sqlx.ExtContext, email string) (*user.User, error) {
	users := []user.User{}
	err := sqlx.SelectContext(
		ctx, db,
		&users,
		`SELECT * FROM users
		 WHERE email = $1`,
		email,
	)
	if err != nil {
		return nil, user.ErrInternal
	}
	if len(users) == 0 {

		return nil, user.ErrNotFound
	}
	user := users[0]
	return &user, nil
}

func (r *userRepository) GetUserByID(ctx context.Context, db sqlx.ExtContext, id int64) (*user.User, error) {
	users := []user.User{}
	err := sqlx.SelectContext(
		ctx, db,
		&users,
		`SELECT * FROM users
		 WHERE id = $1`,
		id,
	)
	if err != nil {
		return nil, user.ErrInternal
	}
	if len(users) == 0 {
		return nil, user.ErrNotFound
	}
	user := users[0]
	return &user, nil

}

func (r *userRepository) GetAllUsers(ctx context.Context, db sqlx.ExtContext, filter user.UserFilter) ([]user.User, error) {
	users := []user.User{}
	baseQuery := `SELECT * FROM users`
	finalQuery, finalArgs, err := buildUserQuery(baseQuery, filter)
	err = sqlx.SelectContext(
		ctx, db,
		&users,
		finalQuery, finalArgs...,
	)
	if err != nil {
		return nil, err
	}
	return users, nil
}
func (r *userRepository) ExistsAsJudge(ctx context.Context, db sqlx.ExtContext, userID int64) (bool, error) {
	var count int
	err := db.QueryRowxContext(ctx,
		`SELECT COUNT(*) FROM users WHERE id = $1 AND role = 'judge'`, userID,
	).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *userRepository) DeleteUser(ctx context.Context, db sqlx.ExtContext, id int64) error {
	_, err := db.ExecContext(ctx, `DELETE FROM users WHERE id = $1`, id)
	return err
}

func (r *userRepository) GetSchoolStatusForAdmin(ctx context.Context, db sqlx.ExtContext, userID int64) (*core.SchoolStatus, error) {
	var status core.SchoolStatus
	err := sqlx.GetContext(ctx, db, &status, `SELECT status FROM schools WHERE school_admin = $1`, userID)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &status, nil
}

func (r *userRepository) SetPasswordHash(ctx context.Context, db sqlx.ExtContext, id int64, hash string) error {
	result, err := db.ExecContext(ctx, `UPDATE users SET password_hash = $1 WHERE id = $2`, hash, id)
	if err != nil {
		return user.ErrInternal
	}
	if rows, _ := result.RowsAffected(); rows == 0 {
		return user.ErrNotFound
	}
	return nil
}

func (r *userRepository) UpdateEmail(ctx context.Context, db sqlx.ExtContext, id int64, email string) error {
	result, err := db.ExecContext(ctx, `UPDATE users SET email = $1 WHERE id = $2`, email, id)
	if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
		return user.ErrAlreadyExists
	}
	if err != nil {
		return user.ErrInternal
	}
	if rows, _ := result.RowsAffected(); rows == 0 {
		return user.ErrNotFound
	}
	return nil
}

func (r *userRepository) CreateResetToken(ctx context.Context, db sqlx.ExtContext, userID int64, hash []byte, expiresAt time.Time) error {
	_, err := db.ExecContext(ctx,
		`INSERT INTO password_reset_tokens (user_id, token_hash, expires_at) VALUES ($1, $2, $3)`,
		userID, hash, expiresAt,
	)
	return err
}

func (r *userRepository) HasRecentResetToken(ctx context.Context, db sqlx.ExtContext, userID int64, since time.Time) (bool, error) {
	var exists bool
	err := sqlx.GetContext(ctx, db, &exists,
		`SELECT EXISTS (SELECT 1 FROM password_reset_tokens WHERE user_id = $1 AND created_at > $2)`,
		userID, since,
	)
	return exists, err
}

func (r *userRepository) ConsumeResetToken(ctx context.Context, db sqlx.ExtContext, hash []byte) (int64, error) {
	var userID int64
	err := sqlx.GetContext(ctx, db, &userID, `
		UPDATE password_reset_tokens
		SET used_at = NOW()
		WHERE token_hash = $1 AND used_at IS NULL AND expires_at > NOW()
		RETURNING user_id`,
		hash,
	)
	if err == sql.ErrNoRows {
		return 0, user.ErrInvalidResetToken
	}
	if err != nil {
		return 0, user.ErrInternal
	}
	return userID, nil
}

func (r *userRepository) InvalidateResetTokens(ctx context.Context, db sqlx.ExtContext, userID int64) error {
	_, err := db.ExecContext(ctx,
		`UPDATE password_reset_tokens SET used_at = NOW() WHERE user_id = $1 AND used_at IS NULL`,
		userID,
	)
	return err
}
