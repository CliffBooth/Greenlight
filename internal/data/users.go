package data

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
	"greenlight.vysotsky.com/internal/validator"
)

var (
	ErrDuplicateEmail = errors.New("duplicate email")
)

type User struct {
	ID        int64     `json:"id"`
	CreatedAt  time.Time `json:"created_at"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Password  password  `json:"-"`
	Activated bool      `json:"activated"`
	Version   int       `json:"-"`
}

type password struct {
	plaintext *string
	hash []byte
}

type UserDao struct {
	DB *sql.DB
}

func (p *password) Set(plaintextPassword string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(plaintextPassword), 12)
	if err != nil {
		return err
	}
	p.plaintext = &plaintextPassword
	p.hash = hash
	return nil
}

func (p *password) Matches(plaintextPassword string) (bool, error) {
	err := bcrypt.CompareHashAndPassword(p.hash, []byte(plaintextPassword))
	if err != nil {
		switch {
		case errors.Is(err, bcrypt.ErrMismatchedHashAndPassword):
			return false, nil
		default:
			return false, err
		}
	}
	return true, nil
}

func ValidateEmail(v *validator.Validator, email string) {
	v.Check(email != "", "email", "must be provided")
	v.Check(validator.Matches(email, validator.EmailRegexp), "email", "must be a valid email address")
}

func ValidatePasswordPlaintext(v *validator.Validator, passwordPlaintext string) {
	v.Check(passwordPlaintext != "", "password", "must not be empty")
	v.Check(len(passwordPlaintext) >= 8, "password", "must be at least 8 characters long")

	// bcrypt has a maximum password length of 72 bytes
	v.Check(len(passwordPlaintext) <= 72, "password", "must not be more han 72 bytes long") 
}

func ValidateUser(v *validator.Validator, user *User) {
	v.Check(user.Name != "", "name", "must be provided")
	v.Check(len(user.Name) <= 500, "name", "must not be longer than 500 characters")

	ValidateEmail(v, user.Email)

	if user.Password.plaintext != nil {
		ValidatePasswordPlaintext(v, *user.Password.plaintext)
	}

	// user from the database must always have hash
	if user.Password.hash == nil {
		panic("missing password hash for user")
	}
}

func (dao UserDao) GetByEmail(email string) (*User, error) {
	query := `
	SELECT id, created_at, name, email, password_hash, activated, version
	FROM users
	WHERE email = $1
	`
	var user User
	ctx, cancel := context.WithTimeout(context.Background(), 3 * time.Second)
	defer cancel()
	err := dao.DB.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.CreatedAt,
		&user.Name,
		&user.Email,
		&user.Password.hash,
		&user.Activated,
		&user.Version,
	)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}
	return &user, nil
}

func (dao UserDao) Insert(user *User) error {
	query := `
	INSERT INTO users (name, email, password_hash, activated)
	VALUES ($1, $2, $3, $4)
	RETURNING id, created_at, version
	`
	args := []interface{}{user.Name, user.Email, user.Password.hash, user.Activated}
	ctx, cancel := context.WithTimeout(context.Background(), 3 * time.Second)
	defer cancel()
	err := dao.DB.QueryRowContext(ctx, query, args...).Scan(&user.ID, &user.CreatedAt, &user.Version)
	if err != nil {
		message := err.Error()
		switch {
		//maybe better use contains
		case strings.Contains(message, "users_email_key") && strings.Contains(message, "unique constraint"):
			return ErrDuplicateEmail
		default:
			return err
		}
	}
	return nil
}

func (dao UserDao) Update(id int, user *User) error {
	query := `
	UPDATE users
	SET name = $1, email = $2, password_hash = $3, activated = $4, version = version + 1
	WHERE id = $5 AND version = $6
	RETURNING version
	`
	args := []interface{}{
		user.Name,
		user.Email,
		user.Password.hash,
		user.Activated,
		user.ID,
		user.Version,
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3 * time.Second)
	defer cancel()
	err := dao.DB.QueryRowContext(ctx, query, args...).Scan(&user.Version)
	if err != nil {
		message := err.Error()
		switch {
		case strings.Contains(message, "users_email_key") && strings.Contains(message, "unique constraint"):
			return ErrDuplicateEmail
		case errors.Is(err, sql.ErrNoRows):
			return ErrRecordNotFound
		default:
			return err
		}
	}
	return nil
}

// func (dao UserDao) Delete(id int) error {

// }