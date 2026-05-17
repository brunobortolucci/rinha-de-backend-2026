package main

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/brunobortolucci/rinha-de-backend-2026/internal/vector"
)

var ready atomic.Bool

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /ready",
		func(w http.ResponseWriter, r *http.Request) {
			if !ready.Load() {
				w.WriteHeader(http.StatusServiceUnavailable)
				return
			}
			w.WriteHeader(http.StatusOK)
			log.Printf("Servidor pronto!")
		},
	)
	mux.HandleFunc("POST /fraud-score", handleFraudScore)
	srv := &http.Server{
		Addr:         ":9999",
		Handler:      mux,
		ReadTimeout:  2 * time.Second,
		WriteTimeout: 2 * time.Second,
		IdleTimeout:  30 * time.Second,
	}
	ready.Store(true)

	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("listen: %s\n", err)
		}
	}()
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("server shutdown: %s\n", err)
	}
}

func handleFraudScore(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var req vector.Request
	res := vector.Response{Approved: false, FraudScore: 0.8}
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		log.Printf("Decode falhou: %s\n", err)
		return
	}
	log.Printf("Body recebido com sucesso, %+v\n", req)
	err = json.NewEncoder(w).Encode(&res)
	if err != nil {
		log.Printf("Encode falhou: %s\n", err)
		return
	}
}
