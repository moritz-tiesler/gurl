package repository

import (
	"context"
	"database/sql"
	_ "embed"
	"gurl/repository/tutorial"
	"log"
)

func New(ctx context.Context, connString string, schema string) (*tutorial.Queries, error) {

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
	return queries, nil
}
