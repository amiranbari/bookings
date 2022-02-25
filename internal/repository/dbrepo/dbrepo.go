package dbrepo

import (
	"database/sql"
	"github.com/amiranbari/bookings/internal/repository"
	"github.com/amiranbari/bookings/pkg/config"
)

type PostgresDBRepo struct {
	App *config.AppConfig
	DB  *sql.DB
}

func NewPostgresRepo(conn *sql.DB, a *config.AppConfig) repository.DatabaseRepo {
	return &PostgresDBRepo{
		App: a,
		DB:  conn,
	}
}
