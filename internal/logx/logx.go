package logx

import "log/slog"

func Service(name string) slog.Attr {
	return slog.String("service", name)
}

func Handler(path string) slog.Attr {
	return slog.String("handler", path)
}
