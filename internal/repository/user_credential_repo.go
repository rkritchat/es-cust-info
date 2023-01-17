package repository

import (
	"database/sql"
	"fmt"
	"time"
)

type UserCredentialEntity struct {
	ID          int       `json:"id"`
	Username    string    `json:"username"`
	Password    string    `json:"password"`
	Email       string    `json:"email"`
	CreatedDate time.Time `json:"created_date"`
	UpdatedDate time.Time `json:"updated_date"`
}

type UserCredentialRepo interface {
	Create(entity UserCredentialEntity) error
	FindByUsernameOrEmail(username, email string) (r UserCredentialEntity, err error)
	FindByUsername(username string) (e UserCredentialEntity, err error)
	FindAllUsername() (r []string, err error)
}

type userCredentialRepo struct {
	db *sql.DB
}

func NewUserCredentialRepo(db *sql.DB) UserCredentialRepo {
	return &userCredentialRepo{
		db: db,
	}
}

func (repo userCredentialRepo) Create(entity UserCredentialEntity) error {
	query := fmt.Sprintf("INSERT INTO user_credential (username, password, email, created_date, updated_date) VALUES  (?, ?, ?, ?, ?)")
	stmt, err := repo.db.Prepare(query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(entity.Username, entity.Password, entity.Email, entity.CreatedDate, entity.UpdatedDate)
	return err
}

func (repo userCredentialRepo) FindByUsernameOrEmail(username, email string) (r UserCredentialEntity, err error) {
	query := fmt.Sprintf("SELECT username, email FROM user_credential WHERE username = ? OR email = ? limit 1")
	var stmt *sql.Stmt
	stmt, err = repo.db.Prepare(query)
	if err != nil {
		return
	}
	defer stmt.Close()

	var rows *sql.Rows
	rows, err = stmt.Query(username, email)
	if err != nil {
		return
	}
	defer rows.Close()

	//case found username or password
	if rows.Next() {
		err = rows.Scan(&r.Username, &r.Email)
		if err != nil {
			return
		}
		return
	}

	//not found
	return
}

func (repo userCredentialRepo) FindByUsername(username string) (r UserCredentialEntity, err error) {
	query := fmt.Sprintf("SELECT username, password FROM user_credential WHERE username = ? LIMIT 1")
	var stmt *sql.Stmt
	stmt, err = repo.db.Prepare(query)
	if err != nil {
		return
	}
	defer stmt.Close()

	var rows *sql.Rows
	rows, err = stmt.Query(username)
	if err != nil {
		return
	}
	defer rows.Close()

	if rows.Next() {
		err = rows.Scan(&r.Username, &r.Password)
		if err != nil {
			return
		}
		return
	}

	return
}

func (repo userCredentialRepo) FindAllUsername() (r []string, err error) {
	query := fmt.Sprintf("SELECT username FROM user_credential")
	var stmt *sql.Stmt
	var rows *sql.Rows
	stmt, err = repo.db.Prepare(query)
	if err != nil {
		return
	}
	defer stmt.Close()
	rows, err = stmt.Query()
	if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		var tmp string
		err = rows.Scan(&tmp)
		if err != nil {
			return
		}
		r = append(r, tmp)
	}
	return
}
