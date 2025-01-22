package sl

import "log/slog"

func Op(op string) slog.Attr {
	return slog.String("operation", op)
}

func Err(err error) slog.Attr {
	return slog.String("error", err.Error())
}
