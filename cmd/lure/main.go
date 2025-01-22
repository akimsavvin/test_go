package main

import (
	"context"
	"github.com/akimsavvin/test_go/internal/infra/app"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := app.
		New(ctx).
		Configure().
		AddServices().
		Run(); err != nil {
		panic(err)
	}
}
