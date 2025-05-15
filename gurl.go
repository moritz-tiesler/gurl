package main

import (
	"context"
	"database/sql"
	_ "embed"
	"log"
	"reflect"

	"gurl/repository"
	"gurl/repository/tutorial"

	_ "modernc.org/sqlite"
)

//go:embed schema.sql
var ddl string

type AuthorRepo interface {
	*tutorial.Queries
}

func run() error {
	ctx := context.Background()

	queries, err := repository.New(ctx, "./gurl.db", ddl)
	if err != nil {
		return err
	}
	// list all authors
	authors, err := queries.ListAuthors(ctx)
	if err != nil {
		return err
	}
	log.Println(authors)

	// create an author
	insertedAuthor, err := queries.CreateAuthor(ctx, tutorial.CreateAuthorParams{
		Name: "Brian Kernighan",
		Bio:  sql.NullString{String: "Co-author of The C Programming Language and The Go Programming Language", Valid: true},
	})
	if err != nil {
		return err
	}
	log.Println(insertedAuthor)

	// get the author we just inserted
	fetchedAuthor, err := queries.GetAuthor(ctx, insertedAuthor.ID)
	if err != nil {
		return err
	}

	// prints true
	log.Println(reflect.DeepEqual(insertedAuthor, fetchedAuthor))
	return nil
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err)

	}
}
