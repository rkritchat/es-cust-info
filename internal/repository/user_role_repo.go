package repository

import (
	"database/sql"
	"fmt"
	"time"
)

type UserRoleEntity struct {
	Id          int       `json:"id"`
	Username    string    `json:"username"`
	RoleId      int       `json:"role_id"`
	RoleDesc    string    `json:"role_desc"`
	IsActive    string    `json:"is_active"`
	AssignedBy  string    `json:"assigned_by"`
	UpdatedDate time.Time `json:"updated_date"`
	CreatedDate time.Time `json:"created_date"`
}

type UserRoleRepo interface {
	GetRolesByUserId(username string) ([]UserRoleEntity, error)
}

type userRoleRepo struct {
	db *sql.DB
}

func NewUserRole(db *sql.DB) UserRoleRepo {
	return &userRoleRepo{
		db: db,
	}
}

func (repo *userRoleRepo) GetRolesByUserId(username string) (result []UserRoleEntity, err error) {
	query := fmt.Sprintf("SELECT role_id FROM user_role WHERE username = ? AND is_active = 1")
	var stmt *sql.Stmt
	var rows *sql.Rows
	stmt, err = repo.db.Prepare(query)
	if err != nil {
		return
	}
	defer stmt.Close()

	rows, err = stmt.Query(username)
	if err != nil {
		return
	}
	defer rows.Close()
	for rows.Next() {
		var tmp UserRoleEntity
		err = rows.Scan(&tmp.RoleId)
		if err != nil {
			return
		}
		result = append(result, tmp)
	}

	return
}
