package main

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/brunobortolucci/rinha-de-backend-2026/internal/config"
	"github.com/brunobortolucci/rinha-de-backend-2026/internal/index"
	"github.com/brunobortolucci/rinha-de-backend-2026/internal/knn"
	"github.com/brunobortolucci/rinha-de-backend-2026/internal/vector"
	"github.com/brunobortolucci/rinha-de-backend-2026/resources"
)

var ready atomic.Bool

func main() {
	opts := &slog.HandlerOptions{Level: slog.LevelInfo}
	logger := slog.New(slog.NewTextHandler(os.Stderr, opts))
	slog.SetDefault(logger)

	norm, err := config.LoadNormalization(resources.Normalization)
	if err != nil {
		slog.Error("Erro ao carregar arquivo normalization.json", "err", err)
		os.Exit(1)
	}
	slog.Info("Carregar arquivo normalization.json", "norm", norm)

	mcc, err := config.LoadMcc(resources.MccRisk)
	if err != nil {
		slog.Error("Erro ao carregar arquivo mcc_risk.json", "err", err)
		os.Exit(1)
	}
	slog.Info("Carregar arquivo mcc_risk.json", "mcc", mcc)

	binPath := os.Getenv("REFERENCES_BIN_PATH")
	if binPath == "" {
		binPath = "/references.bin"
	}
	idx, err := index.Load(binPath)
	if err != nil {
		slog.Error("erro ao carregar references.bin", "err", err)
		os.Exit(1)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /ready", handleReady)
	mux.HandleFunc("POST /fraud-score", handleFraudScore(idx, norm, mcc))
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
			slog.Error("erro ao iniciar o servidor", "err", err)
			os.Exit(1)
		}
	}()
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("erro ao desligar o servidor", "err", err)
		os.Exit(1)
	}
}

func handleFraudScore(idx *index.Index, norm vector.Normalization, mcc config.MccRisk) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		var req vector.Request
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			slog.Error("decode falhou", "err", err)
			return
		}
		vec, err := vector.Vectorize(req, mcc, norm)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			slog.Error("vectorize falhou", "err", err)
			return
		}
		slog.Info("Requisição vetorizada", "req", req, "vec", vec)
		score := knn.Search(idx, vec, 5)
		res := vector.Response{Approved: score < 0.5, FraudScore: score}
		_ = json.NewEncoder(w).Encode(&res)
	}
}

func handleReady(w http.ResponseWriter, r *http.Request) {
	if !ready.Load() {
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}
	w.WriteHeader(http.StatusOK)
}
