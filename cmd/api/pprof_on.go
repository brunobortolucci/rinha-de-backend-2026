//go:build pprof

package main

import (
	"log/slog"
	"net/http"
	_ "net/http/pprof"
)

func init() {
	go func() {
		if err := http.ListenAndServe(":6060", nil); err != nil {
			slog.Error("pprof: erro ao escutar", "err", err)
		}
	}()
}
