package router

import (
	"es-cust-info/internal/custinfo"
	"es-cust-info/internal/login"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/go-chi/jwtauth/v5"
	"net/http"
)

func InitRouter(userCredentialService custinfo.Service, loginService login.Service, auth *jwtauth.JWTAuth) *chi.Mux {
	r := chi.NewRouter()

	r.Use(cors.Handler(cors.Options{
		AllowOriginFunc:  AllowOriginFunc,
		AllowedMethods:   []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	//need valid token
	r.Group(func(router chi.Router) {
		router.Use(jwtauth.Verifier(auth))
		router.Use(jwtauth.Authenticator)
		router.Get("/users", userCredentialService.GetAllUsername)
	})

	r.Post("/user/signup", userCredentialService.Signup)
	r.Post("/user/login", loginService.Login)
	r.Get("/user/refresh", loginService.RefreshToken)
	return r
}

func AllowOriginFunc(r *http.Request, origin string) bool {
	if origin == "http://localhost:5173" {
		return true
	}
	return false
}
