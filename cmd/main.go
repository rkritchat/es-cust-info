package main

import (
	"es-cust-info/internal/config"
	"es-cust-info/internal/custinfo"
	"es-cust-info/internal/login"
	"es-cust-info/internal/repository"
	"es-cust-info/internal/router"
	"fmt"
	"net/http"
)

func main() {
	//init config
	cfg := config.InitConfig()
	defer cfg.Free()

	//init repo
	userCredentialRepo := repository.NewUserCredentialRepo(cfg.DB)
	userRoleRepo := repository.NewUserRole(cfg.DB)

	//init service
	userCredentialService := custinfo.NewService(userCredentialRepo)
	loginService := login.NewService(userCredentialRepo, userRoleRepo, cfg.JwtAuth, cfg.Env)

	//init router
	r := router.InitRouter(userCredentialService, loginService, cfg.JwtAuth)

	fmt.Printf("start on port %v", cfg.Env.Port)
	err := http.ListenAndServe(cfg.Env.Port, r)
	if err != nil {
		panic(err)
	}
}
