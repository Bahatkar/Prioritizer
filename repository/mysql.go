package repository

import (
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
)

type Config struct {
	Host     string
	Port     string
	Username string
	Protocol string
	Password string
	DBName   string
}

func NewMySQLDB(cfg Config) (*sql.DB, error) {
	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@%s(%s:%s)/%s", cfg.Username, cfg.Password, cfg.Protocol,
		cfg.Host, cfg.Port, cfg.DBName))
	if err != nil {
		return nil, err
	}

	if err = db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}
