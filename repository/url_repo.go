package repository

import (
	"context"
	"database/sql"
	_ "embed"
	"gurl/repository/tutorial"
	"log"
)

type Repo interface {
	tutorial.Querier
	WithTx(tx *sql.Tx) Repo
	DB() *sql.DB
}

type Queries struct {
	*tutorial.Queries
	db *sql.DB
}

func (q *Queries) WithTx(tx *sql.Tx) Repo {
	return &Queries{
		Queries: q.Queries.WithTx(tx),
		db:      q.db,
	}
}

func (q *Queries) DB() *sql.DB {
	return q.db
}

func New(ctx context.Context, connString string, schema string) (*Queries, error) {

	db, err := sql.Open("sqlite", connString)
	if err != nil {
		return nil, err
	}

	_, err = db.Exec("PRAGMA journal_mode=WAL;")
	if err != nil {
		log.Fatalf("Failed to set journal_mode: %v", err)
	}

	db.SetMaxOpenConns(1)

	if _, err := db.ExecContext(ctx, schema); err != nil {
		return nil, err
	}

	queries := tutorial.New(db)

	return &Queries{Queries: queries, db: db}, nil
}
