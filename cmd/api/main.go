package main

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/brunobortolucci/rinha-de-backend-2026/internal/config"
	"github.com/brunobortolucci/rinha-de-backend-2026/internal/index"
	"github.com/brunobortolucci/rinha-de-backend-2026/internal/knn"
	"github.com/brunobortolucci/rinha-de-backend-2026/internal/vector"
	"github.com/brunobortolucci/rinha-de-backend-2026/resources"
)

var (
	ready       atomic.Bool
	requestPool = sync.Pool{New: func() any { return new(vector.Request) }}
)

func main() {
	opts := &slog.HandlerOptions{Level: slog.LevelError}
	logger := slog.New(slog.NewTextHandler(os.Stderr, opts))
	slog.SetDefault(logger)

	norm, err := config.LoadNormalization(resources.Normalization)
	if err != nil {
		slog.Error("Erro ao carregar arquivo normalization.json", "err", err)
		os.Exit(1)
	}

	mcc, err := config.LoadMcc(resources.MccRisk)
	if err != nil {
		slog.Error("Erro ao carregar arquivo mcc_risk.json", "err", err)
		os.Exit(1)
	}

	binPath := os.Getenv("REFERENCES_BIN_PATH")
	if binPath == "" {
		binPath = "/references.bin"
	}
	idx, err := index.Load(binPath)
	if err != nil {
		slog.Error("erro ao carregar references.bin", "err", err)
		os.Exit(1)
	}
	idx.Warmup()

	nprobe := knn.DefaultNProbe
	if v, err := strconv.Atoi(os.Getenv("NPROBE")); err == nil && v > 0 {
		nprobe = v
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /ready", handleReady)
	mux.HandleFunc("POST /fraud-score", handleFraudScore(idx, norm, mcc, nprobe))
	srv := &http.Server{
		Addr:         ":9999",
		Handler:      mux,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	ln, err := net.Listen("tcp", srv.Addr)
	if err != nil {
		slog.Error("erro ao escutar", "err", err)
		os.Exit(1)
	}
	ready.Store(true)

	go func() {
		if err := srv.Serve(ln); err != nil && !errors.Is(err, http.ErrServerClosed) {
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

const fallbackBudgetMs = 1500

var fallbackResponse = []byte(`{"approved":true,"fraud_score":0.0000}`)

func handleFraudScore(idx *index.Index, norm vector.Normalization, mcc config.MccRisk, nprobe int) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if isOverBudget(r.Header.Get("X-Req-Start")) {
			_, _ = w.Write(fallbackResponse)
			return
		}

		req := requestPool.Get().(*vector.Request)
		*req = vector.Request{}
		defer requestPool.Put(req)

		if err := json.NewDecoder(r.Body).Decode(req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		vec, err := vector.Vectorize(*req, mcc, norm)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		score := knn.Search(idx, vec, 5, nprobe)

		buf := make([]byte, 0, 48)
		buf = append(buf, `{"approved":`...)
		buf = strconv.AppendBool(buf, score < 0.5)
		buf = append(buf, `,"fraud_score":`...)
		buf = strconv.AppendFloat(buf, score, 'f', 4, 64)
		buf = append(buf, '}')
		_, _ = w.Write(buf)
	}
}

func isOverBudget(reqStart string) bool {
	if reqStart == "" {
		return false
	}
	dot := -1
	for i := 0; i < len(reqStart); i++ {
		if reqStart[i] == '.' {
			dot = i
			break
		}
	}
	if dot < 0 {
		return false
	}
	secs, err := strconv.ParseInt(reqStart[:dot], 10, 64)
	if err != nil {
		return false
	}
	ms, err := strconv.ParseInt(reqStart[dot+1:], 10, 64)
	if err != nil {
		return false
	}
	startMs := secs*1000 + ms
	nowMs := time.Now().UnixMilli()
	return nowMs-startMs > fallbackBudgetMs
}

func handleReady(w http.ResponseWriter, r *http.Request) {
	if !ready.Load() {
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}
	w.WriteHeader(http.StatusOK)
}
