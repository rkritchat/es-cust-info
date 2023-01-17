package custinfo

import (
	"encoding/json"
	"errors"
	"es-cust-info/internal/common"
	"es-cust-info/internal/repository"
	"es-cust-info/internal/utils"
	"fmt"
	"net/http"
	"time"
)

const (
	invalidJsonFormat = "invalid request json format"
	internalServerErr = "internal server error"
)

type Service interface {
	Signup(w http.ResponseWriter, r *http.Request)
	GetAllUsername(w http.ResponseWriter, r *http.Request)
}

type service struct {
	userCredentialRepo repository.UserCredentialRepo
}

func NewService(userCredentialRepo repository.UserCredentialRepo) Service {
	return &service{
		userCredentialRepo: userCredentialRepo,
	}
}

type SignupReq struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Email    string `json:"email"`
}

type SignupResp struct {
	common.Resp
}

func (s service) Signup(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("content-type", "application/json")

	var req SignupReq
	var errMsg string
	var errCode int
	var hashedPwd string
	now := time.Now()
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		errMsg = invalidJsonFormat
		goto FAILED
	}

	//validate duplicate username and email
	err = s.validateDuplicate(req.Username, req.Email)
	if err != nil {
		errMsg = err.Error()
		errCode = http.StatusBadRequest
		goto FAILED
	}

	//hashed password
	hashedPwd, err = utils.HashPassword(req.Password)
	if err != nil {
		errMsg = internalServerErr
		errCode = http.StatusInternalServerError
		goto FAILED
	}

	//create user credential
	err = s.userCredentialRepo.Create(repository.UserCredentialEntity{
		Username:    req.Username,
		Password:    hashedPwd,
		Email:       req.Email,
		CreatedDate: now,
		UpdatedDate: now,
	})
	if err != nil {
		errMsg = internalServerErr
		errCode = http.StatusInternalServerError
		goto FAILED
	}

	//signup success
	_ = json.NewEncoder(w).Encode(&SignupResp{
		Resp: common.Resp{
			Message: "success",
		},
	})

	return
FAILED:
	w.WriteHeader(errCode)
	_ = json.NewEncoder(w).Encode(&SignupResp{
		Resp: common.Resp{
			Message: "failed",
			Error:   errMsg,
		},
	})
}

type GetAllUsernameResp struct {
	common.Resp
	Usernames []string `json:"usernames"`
}

func (s service) GetAllUsername(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("content-type", "application/json")
	users, err := s.userCredentialRepo.FindAllUsername()
	if err != nil {
		_ = json.NewEncoder(w).Encode(&GetAllUsernameResp{
			Resp: common.Resp{Message: "failed", Error: "internal server error"},
		})
		return
	}

	fmt.Println(users)
	_ = json.NewEncoder(w).Encode(&GetAllUsernameResp{
		Resp:      common.Resp{Message: "success"},
		Usernames: users,
	})
}

func (s service) validateDuplicate(username, email string) error {
	entity, err := s.userCredentialRepo.FindByUsernameOrEmail(username, email)
	if err != nil {
		return errors.New("internal server error")
	}

	if entity.Username == username {
		return errors.New("username is already exist")
	}

	if entity.Email == email {
		return errors.New("email is already exist")
	}
	return nil
}
