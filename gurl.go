package main

import (
	"context"
	_ "embed"
	"log"
	"reflect"

	"gurl/repository"
	"gurl/repository/tutorial"

	_ "modernc.org/sqlite"
)

//go:embed schema.sql
var ddl string

type URLRepo interface {
	*tutorial.Queries
}

func run() error {
	ctx := context.Background()

	queries, err := repository.New(ctx, "./gurl.db", ddl)
	if err != nil {
		return err
	}
	// list all authors
	urls, err := queries.ListUrls(ctx)
	if err != nil {
		return err
	}
	log.Println(urls)

	// create an author
	insertedUrl, err := queries.CreateUrl(ctx, tutorial.CreateUrlParams{
		Original: "https://www.zeit.de",
		Short:    "gurl.me/abba",
	})
	if err != nil {
		return err
	}
	log.Println(insertedUrl)

	// get the author we just inserted
	fetchedAuthor, err := queries.GetUrl(ctx, insertedUrl.ID)
	if err != nil {
		return err
	}

	// prints true
	log.Println(reflect.DeepEqual(insertedUrl, fetchedAuthor))
	return nil
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err)

	}
}
