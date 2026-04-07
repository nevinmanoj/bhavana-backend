package user

import (
	"context"
	"log"

	"github.com/jmoiron/sqlx"
	auth "github.com/nevinmanoj/bhavana-backend/internal/auth"
	"github.com/nevinmanoj/bhavana-backend/internal/core"
	user "github.com/nevinmanoj/bhavana-backend/internal/domain/user"
)

type userRepository struct {
	db *sqlx.DB
}

func NewUserWriteRepository(db *sqlx.DB) user.UserWriteRepository {
	return &userRepository{db: db}
}
func NewUserReadRepository(db *sqlx.DB) user.UserReadRepository {
	return &userRepository{db: db}
}

func (r *userRepository) CreateUser(ctx context.Context, email, password, name string, role core.UserRole) (*user.User, error) {
	//check if email already exists
	var exists bool
	err := r.db.QueryRowContext(
		ctx,
		`SELECT EXISTS (
		SELECT 1
		FROM users
		WHERE email = $1
	)`,
		email,
	).Scan(&exists)

	if err != nil {
		log.Println("Error checking if email exists:", err)
		return nil, user.ErrInternal
	}

	if exists {
		return nil, user.ErrAlreadyExists
	}
	passwordHash, err := auth.HashPassword(password)
	if err != nil {
		log.Println("Error hashing password:", err)
		return nil, user.ErrInternal
	}
	userToCreate := &user.User{
		Email:        email,
		PasswordHash: passwordHash,
		Name:         name,
		Role:         role,
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

	rows, err := r.db.NamedQueryContext(ctx, query, userToCreate)
	if err != nil {
		log.Println("Error inserting user:", err)
		return nil, user.ErrInternal
	}
	defer rows.Close()

	if rows.Next() {
		rows.Scan(&userToCreate.ID, &userToCreate.CreatedAt)
		return userToCreate, nil
	}

	return nil, user.ErrInternal
}

func (r *userRepository) GetUserByEmail(ctx context.Context, email string) (*user.User, error) {
	users := []user.User{}
	err := r.db.SelectContext(
		ctx,
		&users,
		`SELECT * FROM users
		 WHERE email = $1`,
		email,
	)
	if err != nil {
		log.Println("Error fetching user by email:", err)
		return nil, user.ErrInternal
	}
	if len(users) == 0 {

		return nil, user.ErrNotFound
	}
	user := users[0]
	return &user, nil
}

func (r *userRepository) GetUserByID(ctx context.Context, id int64) (*user.User, error) {
	users := []user.User{}
	err := r.db.SelectContext(
		ctx,
		&users,
		`SELECT * FROM users
		 WHERE id = $1`,
		id,
	)
	if err != nil {
		log.Println("Error fetching user by id:", err)
		return nil, user.ErrInternal
	}
	if len(users) == 0 {
		return nil, user.ErrNotFound
	}
	user := users[0]
	return &user, nil

}

func (r *userRepository) GetAllUsers(ctx context.Context, filter user.UserFilter) ([]user.User, error) {
	users := []user.User{}
	baseQuery := `SELECT * FROM users`
	finalQuery, finalArgs, err := buildUserQuery(baseQuery, filter)
	err = r.db.SelectContext(
		ctx,
		&users,
		finalQuery, finalArgs...,
	)
	if err != nil {
		return nil, err
	}
	return users, nil
}
