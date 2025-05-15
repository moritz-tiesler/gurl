package repository

import (
	"context"
	"database/sql"
	_ "embed"
	"gurl/repository/tutorial"
)

func New(ctx context.Context, connString string, schema string) (*tutorial.Queries, error) {

	db, err := sql.Open("sqlite", connString)
	if err != nil {
		return nil, err
	}
	// create tables
	if _, err := db.ExecContext(ctx, schema); err != nil {
		return nil, err
	}

	queries := tutorial.New(db)
	return queries, nil
}
