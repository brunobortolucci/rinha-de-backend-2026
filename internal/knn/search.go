package knn

import (
	"runtime"
	"sync"

	"github.com/brunobortolucci/rinha-de-backend-2026/internal/index"
)

const dims = 14

func Search(idx *index.Index, query [dims]float64, k int) float64 {
	var q [dims]float32
	for i := 0; i < dims; i++ {
		q[i] = float32(query[i])
	}

	workers := runtime.NumCPU()
	if workers > idx.Count {
		workers = idx.Count
	}
	if workers < 1 {
		workers = 1
	}

	chunk := idx.Count / workers
	heaps := make([]*topK, workers)

	var wg sync.WaitGroup
	wg.Add(workers)
	for w := 0; w < workers; w++ {
		start := w * chunk
		end := start + chunk
		if w == workers-1 {
			end = idx.Count
		}
		go func(w, start, end int) {
			defer wg.Done()
			heaps[w] = scanChunk(idx.Vectors, idx.Labels, q, start, end, k)
		}(w, start, end)
	}
	wg.Wait()

	final := newTopK(k)
	for _, h := range heaps {
		for _, n := range h.drain() {
			final.push(n)
		}
	}

	var fraud int
	winners := final.drain()
	for _, n := range winners {
		if n.label == 1 {
			fraud++
		}
	}
	return float64(fraud) / float64(len(winners))
}

func scanChunk(vectors []float32, labels []byte, q [dims]float32, start, end, k int) *topK {
	h := newTopK(k)
	for i := start; i < end; i++ {
		base := i * dims
		d0 := q[0] - vectors[base+0]
		d1 := q[1] - vectors[base+1]
		d2 := q[2] - vectors[base+2]
		d3 := q[3] - vectors[base+3]
		d4 := q[4] - vectors[base+4]
		d5 := q[5] - vectors[base+5]
		d6 := q[6] - vectors[base+6]
		d7 := q[7] - vectors[base+7]
		d8 := q[8] - vectors[base+8]
		d9 := q[9] - vectors[base+9]
		d10 := q[10] - vectors[base+10]
		d11 := q[11] - vectors[base+11]
		d12 := q[12] - vectors[base+12]
		d13 := q[13] - vectors[base+13]
		dist := d0*d0 + d1*d1 + d2*d2 + d3*d3 + d4*d4 + d5*d5 + d6*d6 +
			d7*d7 + d8*d8 + d9*d9 + d10*d10 + d11*d11 + d12*d12 + d13*d13

		if h.full() && dist >= h.worst() {
			continue
		}
		h.push(neighbor{dist: dist, label: labels[i]})
	}
	return h
}
