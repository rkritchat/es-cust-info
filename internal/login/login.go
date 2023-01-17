package login

import (
	"encoding/json"
	"errors"
	"es-cust-info/internal/common"
	"es-cust-info/internal/config"
	"es-cust-info/internal/repository"
	"es-cust-info/internal/utils"
	"fmt"
	"github.com/go-chi/jwtauth/v5"
	"net/http"
	"strings"
	"time"
)

const (
	invalidUsernameOrPassword = "invalid username or password"
)

type Service interface {
	Login(w http.ResponseWriter, r *http.Request)
	RefreshToken(w http.ResponseWriter, r *http.Request)
}

type service struct {
	userCredentialRepo repository.UserCredentialRepo
	userRoleRepo       repository.UserRoleRepo
	jwtAuth            *jwtauth.JWTAuth
	env                config.Env
}

func NewService(userCredentialRepo repository.UserCredentialRepo, userRoleRepo repository.UserRoleRepo, jwtAuth *jwtauth.JWTAuth, env config.Env) Service {
	return &service{
		userCredentialRepo: userCredentialRepo,
		userRoleRepo:       userRoleRepo,
		jwtAuth:            jwtAuth,
		env:                env,
	}
}

type Req struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type Resp struct {
	common.Resp
	Username string `json:"username,omitempty"`
	Roles    []int  `json:"roles,omitempty"`
	Token    string `json:"token,omitempty"`
}

func (s service) Login(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("content-type", "application/json")
	var req Req
	var errMsg string
	var statusCode int
	var entity repository.UserCredentialEntity
	var token string
	var roles []int
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		errMsg = common.InvalidReqJsonFormat
		statusCode = http.StatusInternalServerError
		goto FAILED
	}

	//get user by username
	entity, err = s.userCredentialRepo.FindByUsername(req.Username)
	if err != nil {
		errMsg = common.InternalServerError
		statusCode = http.StatusInternalServerError
		goto FAILED
	}

	//validate username and password
	err = validateLogin(req.Password, entity)
	if err != nil {
		errMsg = err.Error()
		statusCode = http.StatusBadRequest
		goto FAILED
	}

	//get user roles
	roles, err = s.getUserRole(req.Username)
	if err != nil {
		errMsg = common.InternalServerError
		statusCode = http.StatusInternalServerError
		goto FAILED
	}

	//generate token
	token = s.generateToken(req.Username, time.Now().Add(time.Duration(s.env.JwtExpiredInHours)*time.Hour))
	http.SetCookie(w, &http.Cookie{Name: "Token", Value: token, HttpOnly: false})

	_ = json.NewEncoder(w).Encode(&Resp{
		Resp: common.Resp{
			Message: "success",
		},
		Username: req.Username,
		Roles:    roles,
		Token:    token,
	})
	return
FAILED:
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(&Resp{
		Resp: common.Resp{
			Message: "failed",
			Error:   errMsg,
		},
	})
}

func (s service) RefreshToken(w http.ResponseWriter, r *http.Request) {
	reqToken := strings.Split(r.Header.Get("Authorization"), " ")
	if len(reqToken) != 2 {
		http.Error(w, "token is required", http.StatusBadRequest)
		return
	}
	fmt.Println(reqToken)
	token, err := s.jwtAuth.Decode(reqToken[1])
	if err != nil {
		http.Error(w, "invalid token", http.StatusBadRequest)
		return
	}

	username, _ := token.Get("username")
	fmt.Print(token.Get("username"))

	newToken := s.generateToken(username.(string), time.Now().Add(time.Duration(s.env.JwtExpiredInHours)*time.Hour))

	w.Header().Add("Content-type", "application/json")
	_ = json.NewEncoder(w).Encode(&Resp{
		Resp: common.Resp{
			Message: "success",
		},
		Username: username.(string),
		Token:    newToken,
	})
}

func validateLogin(password string, entity repository.UserCredentialEntity) error {
	if len(entity.Username) == 0 || len(entity.Password) == 0 {
		return errors.New(invalidUsernameOrPassword)
	}

	//validate plain password with hashed password
	if !utils.ComparePassword(entity.Password, password) {
		return errors.New(invalidUsernameOrPassword)
	}
	return nil
}

func (s service) generateToken(username string, expired time.Time) string {
	_, token, _ := s.jwtAuth.Encode(map[string]interface{}{"username": username, "exp": expired})
	return token
}

func (s service) getUserRole(username string) ([]int, error) {
	result, err := s.userRoleRepo.GetRolesByUserId(username)
	if err != nil {
		return nil, err
	}
	var r []int
	for _, v := range result {
		r = append(r, v.RoleId)
	}
	return r, nil
}
