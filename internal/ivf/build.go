package ivf

import (
	"math/rand"
	"runtime"
	"sync"
)

const (
	trainSample = 200_000
	trainIters  = 20
	trainSeed   = 42
)

// Build roda k-means sobre uma amostra, atribui todos os vetores ao
// centróide mais próximo e devolve:
//   - centroids: nlist × Dims em int16
//   - offsets: nlist+1 posições; o cluster c ocupa perm[offsets[c]:offsets[c+1]]
//   - perm: índices originais reordenados por cluster
func Build(vectors []int16, nlist int) (centroids []int16, offsets []uint32, perm []uint32) {
	count := len(vectors) / Dims
	if count == 0 {
		return nil, make([]uint32, 1), nil
	}
	if nlist > count {
		nlist = count
	}

	centroids = train(vectors, count, nlist)

	assign := make([]uint16, count)
	parallelFor(count, func(lo, hi int) {
		for i := lo; i < hi; i++ {
			base := i * Dims
			best, bestDist := 0, int32(1<<31-1)
			for c := 0; c < nlist; c++ {
				d := distSquared(vectors, base, centroids, c*Dims)
				if d < bestDist {
					best, bestDist = c, d
				}
			}
			assign[i] = uint16(best)
		}
	})

	sizes := make([]uint32, nlist)
	for _, c := range assign {
		sizes[c]++
	}
	offsets = make([]uint32, nlist+1)
	for c := 0; c < nlist; c++ {
		offsets[c+1] = offsets[c] + sizes[c]
	}

	perm = make([]uint32, count)
	cursor := make([]uint32, nlist)
	copy(cursor, offsets[:nlist])
	for i := 0; i < count; i++ {
		c := assign[i]
		perm[cursor[c]] = uint32(i)
		cursor[c]++
	}
	return centroids, offsets, perm
}

// train: k-means de Lloyd sobre uma amostra por stride, com centróides
// re-quantizados para int16 a cada iteração (mantém a atribuição em
// aritmética inteira, a mesma usada em runtime).
func train(vectors []int16, count, nlist int) []int16 {
	stride := count / trainSample
	if stride < 1 {
		stride = 1
	}
	var sample []uint32
	for i := 0; i < count; i += stride {
		sample = append(sample, uint32(i))
	}

	rng := rand.New(rand.NewSource(trainSeed))
	centroids := make([]int16, nlist*Dims)
	for c, si := range rng.Perm(len(sample))[:nlist] {
		base := int(sample[si]) * Dims
		copy(centroids[c*Dims:(c+1)*Dims], vectors[base:base+Dims])
	}

	workers := runtime.NumCPU()
	for iter := 0; iter < trainIters; iter++ {
		sums := make([][]int64, workers)
		counts := make([][]int64, workers)
		for w := range sums {
			sums[w] = make([]int64, nlist*Dims)
			counts[w] = make([]int64, nlist)
		}

		var wg sync.WaitGroup
		chunk := (len(sample) + workers - 1) / workers
		for w := 0; w < workers; w++ {
			lo, hi := w*chunk, min((w+1)*chunk, len(sample))
			if lo >= hi {
				continue
			}
			wg.Add(1)
			go func(w, lo, hi int) {
				defer wg.Done()
				s, ct := sums[w], counts[w]
				for _, idx := range sample[lo:hi] {
					base := int(idx) * Dims
					best, bestDist := 0, int32(1<<31-1)
					for c := 0; c < nlist; c++ {
						d := distSquared(vectors, base, centroids, c*Dims)
						if d < bestDist {
							best, bestDist = c, d
						}
					}
					for d := 0; d < Dims; d++ {
						s[best*Dims+d] += int64(vectors[base+d])
					}
					ct[best]++
				}
			}(w, lo, hi)
		}
		wg.Wait()

		for w := 1; w < workers; w++ {
			for i := range sums[0] {
				sums[0][i] += sums[w][i]
			}
			for i := range counts[0] {
				counts[0][i] += counts[w][i]
			}
		}
		for c := 0; c < nlist; c++ {
			n := counts[0][c]
			if n == 0 {
				continue
			}
			for d := 0; d < Dims; d++ {
				centroids[c*Dims+d] = int16(sums[0][c*Dims+d] / n)
			}
		}
	}
	return centroids
}

func parallelFor(n int, fn func(lo, hi int)) {
	workers := runtime.NumCPU()
	chunk := (n + workers - 1) / workers
	var wg sync.WaitGroup
	for w := 0; w < workers; w++ {
		lo, hi := w*chunk, min((w+1)*chunk, n)
		if lo >= hi {
			continue
		}
		wg.Add(1)
		go func(lo, hi int) {
			defer wg.Done()
			fn(lo, hi)
		}(lo, hi)
	}
	wg.Wait()
}
