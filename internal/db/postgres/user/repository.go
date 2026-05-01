package user

import (
	"context"

	"github.com/jmoiron/sqlx"
	auth "github.com/nevinmanoj/bhavana-backend/internal/auth"
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
	passwordHash, err := auth.HashPassword(password)
	if err != nil {
		return user.ErrInternal
	}
	userToCreate.PasswordHash = passwordHash

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
