package main

import (
	"context"
	"database/sql"
	"fmt"
)

type QueryExec interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

func main() {
	ctx := context.Background()
	h, _ := ctx.Value("tx").(*sql.DB)
	fmt.Println(h)
}
