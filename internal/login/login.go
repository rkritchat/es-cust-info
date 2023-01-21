package login

import (
	"encoding/json"
	"errors"
	"es-cust-info/internal/cache"
	"es-cust-info/internal/common"
	"es-cust-info/internal/config"
	"es-cust-info/internal/repository"
	"es-cust-info/internal/utils"
	"fmt"
	"github.com/go-chi/jwtauth/v5"
	"net/http"
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
	cache              cache.Cache
	env                config.Env
}

func NewService(userCredentialRepo repository.UserCredentialRepo, userRoleRepo repository.UserRoleRepo, jwtAuth *jwtauth.JWTAuth, cache cache.Cache, env config.Env) Service {
	return &service{
		userCredentialRepo: userCredentialRepo,
		userRoleRepo:       userRoleRepo,
		jwtAuth:            jwtAuth,
		cache:              cache,
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
	var refreshToken string
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
	token = s.generateToken(req.Username, time.Now().Add(time.Duration(s.env.JwtExpInMinute)*time.Minute))
	refreshToken = s.generateToken(req.Username, time.Now().Add(time.Duration(s.env.JwtExpInMinute)*time.Minute))
	err = s.cache.Set(req.Username, refreshToken, time.Duration(s.env.JwtRefreshExpInMinute)*time.Minute)
	if err != nil {
		errMsg = common.InternalServerError
		statusCode = http.StatusInternalServerError
		goto FAILED
	}
	http.SetCookie(w, &http.Cookie{Name: "token", Value: refreshToken, HttpOnly: false, Secure: false})

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
	c, err := r.Cookie("token")
	if err != nil {
		http.Error(w, "no token found", http.StatusUnauthorized)
		return
	}
	fmt.Println(c.Value)
	token, err := s.jwtAuth.Decode(c.Value)
	if err != nil {
		http.Error(w, "invalid token", http.StatusBadRequest)
		return
	}

	username, _ := token.Get("username")
	fmt.Print(username)
	_, err = s.cache.Get(username.(string))
	if err != nil {
		http.Error(w, "no token found", http.StatusUnauthorized)
		return
	}

	//get user roles
	roles, err := s.getUserRole(username.(string))
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	//newToken := s.generateToken(username.(string), time.Now().Add(time.Duration(s.env.JwtExpInMinute)*time.Minute))
	http.SetCookie(w, &http.Cookie{Name: "token", Value: c.Value, HttpOnly: false, Secure: false})

	w.Header().Add("Content-type", "application/json")
	_ = json.NewEncoder(w).Encode(&Resp{
		Resp: common.Resp{
			Message: "success",
		},
		Username: username.(string),
		Token:    c.Value,
		Roles:    roles,
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
