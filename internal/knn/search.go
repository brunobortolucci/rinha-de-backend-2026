package knn

import (
	"github.com/brunobortolucci/rinha-de-backend-2026/internal/index"
	"github.com/brunobortolucci/rinha-de-backend-2026/internal/ivf"
	"github.com/brunobortolucci/rinha-de-backend-2026/internal/topk"
)

// DefaultNProbe: quantos clusters IVF visitar por busca. Calibrado offline
// com cmd/validate — sobe recall (menos FP/FN), custa latência linearmente.
// 8 → 0.12% de erro | 16 → 0.08% | 32 → 0.06% (teto do int16)
const DefaultNProbe = 8

func Search(idx *index.Index, query [ivf.Dims]float64, k, nprobe int) float64 {
	var q [ivf.Dims]int16
	for i := 0; i < ivf.Dims; i++ {
		q[i] = ivf.Quantize(query[i])
	}

	h := ivf.Search(idx.Centroids, idx.Offsets, idx.Vectors, idx.Labels, q, k, nprobe)
	return FraudRate(h)
}

func FraudRate(h *topk.Heap) float64 {
	winners := h.Drain()
	var fraud int
	for _, n := range winners {
		if n.Label == 1 {
			fraud++
		}
	}
	return float64(fraud) / float64(len(winners))
}
