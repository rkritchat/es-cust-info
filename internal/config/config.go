package config

import (
	"database/sql"
	"fmt"
	"github.com/caarlos0/env/v6"
	"github.com/go-chi/jwtauth/v5"
	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
	"github.com/lestrrat-go/jwx/v2/jwa"
)

type Cfg struct {
	DB      *sql.DB
	Env     Env
	JwtAuth *jwtauth.JWTAuth
}

type Env struct {
	Port              string `env:"PORT"`
	MySQLUser         string `env:"MYSQL_USER"`
	MySQLPwd          string `env:"MYSQL_PWD"`
	MySQLHost         string `env:"MYSQL_HOST"`
	MySQLDBName       string `env:"MYSQL_DB_NAME"`
	JwtTokenSecret    string `env:"JWT_TOKEN_SECRET"`
	JwtExpiredInHours int    `env:"JWT_EXPIRED_IN_HOURS"`
}

func InitConfig() *Cfg {
	localEnv := initLocalEnv()
	return &Cfg{
		DB:      initMysqlCon(localEnv),
		Env:     localEnv,
		JwtAuth: jwtauth.New(jwa.HS256.String(), []byte(localEnv.JwtTokenSecret), nil),
	}
}

func initLocalEnv() Env {
	//load local env file
	err := godotenv.Load()
	if err != nil {
		panic(err)
	}

	var localEnv Env
	err = env.Parse(&localEnv)
	if err != nil {
		panic(err)
	}
	return localEnv
}

func initMysqlCon(env Env) *sql.DB {
	url := fmt.Sprintf("%v:%v@tcp(%v)/%v?parseTime=true", env.MySQLUser, env.MySQLPwd, env.MySQLHost, env.MySQLDBName)

	//connect db
	conn, err := sql.Open("mysql", url)
	if err != nil {
		panic(err)
	}

	//ping DB
	err = conn.Ping()
	if err != nil {
		panic(err)
	}
	return conn
}

func (c *Cfg) Free() {
	if c.DB != nil {
		err := c.DB.Close()
		if err != nil {
			panic(err)
		}
	}
}
