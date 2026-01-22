package dbg_server

import (
	"log/slog"
	"net/http"
)

func Run(port string, log *slog.Logger) {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	addr := "127.0.0.1:" + port
	log.Info("starting debug server", "addr", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Error("debug server error", "error", err)
	}
}
