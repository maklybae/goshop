package db

import (
	"embed"
	"log"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

//go:embed migrations/*.sql
var embedMigrations embed.FS

func MustMigrate(cfg *pgx.ConnConfig) {
	goose.SetBaseFS(embedMigrations)

	if err := goose.SetDialect("postgres"); err != nil {
		log.Fatal("goose.SetDialect", err)
	}

	db := stdlib.OpenDB(*cfg)
	defer func() { _ = db.Close() }()

	if err := goose.Up(db, "migrations"); err != nil {
		panic(err)
	}
}
