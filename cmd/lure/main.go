package main

import (
	"context"
	"github.com/akimsavvin/gonet/v2/graceful"
	"github.com/akimsavvin/test_go/internal/infra/app"
	"log"
)

func main() {
	ctx, cancel := graceful.Context(context.Background())
	defer cancel()

	if err := app.Run(ctx); err != nil {
		log.Fatalln(err)
	}
}
