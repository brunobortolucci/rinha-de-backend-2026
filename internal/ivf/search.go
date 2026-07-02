package ivf

import (
	"github.com/brunobortolucci/rinha-de-backend-2026/internal/topk"
)

const maxProbe = 64

// Search acha os nprobe centróides mais próximos da query e faz brute
// force só dentro desses clusters. vectors e labels devem estar
// reordenados por cluster (layout produzido por Build).
func Search(centroids []int16, offsets []uint32, vectors []int16, labels []byte, q [Dims]int16, k, nprobe int) *topk.Heap {
	h := topk.New(k)
	nlist := len(offsets) - 1
	if nlist <= 0 {
		return h
	}
	if nprobe > nlist {
		nprobe = nlist
	}
	if nprobe > maxProbe {
		nprobe = maxProbe
	}

	var probeDist [maxProbe]int32
	var probeID [maxProbe]int32
	for i := 0; i < nprobe; i++ {
		probeDist[i] = 1<<31 - 1
	}
	for c := 0; c < nlist; c++ {
		d := distToQuery(centroids, c*Dims, q)
		if d >= probeDist[nprobe-1] {
			continue
		}
		j := nprobe - 1
		for j > 0 && probeDist[j-1] > d {
			probeDist[j] = probeDist[j-1]
			probeID[j] = probeID[j-1]
			j--
		}
		probeDist[j] = d
		probeID[j] = int32(c)
	}

	for p := 0; p < nprobe; p++ {
		c := probeID[p]
		start, end := int(offsets[c]), int(offsets[c+1])
		for i := start; i < end; i++ {
			d := distToQuery(vectors, i*Dims, q)
			if h.Full() && d >= h.Worst() {
				continue
			}
			h.Push(topk.Neighbor{Dist: d, Label: labels[i]})
		}
	}
	return h
}

func distToQuery(vectors []int16, base int, q [Dims]int16) int32 {
	d0 := int32(q[0]) - int32(vectors[base+0])
	d1 := int32(q[1]) - int32(vectors[base+1])
	d2 := int32(q[2]) - int32(vectors[base+2])
	d3 := int32(q[3]) - int32(vectors[base+3])
	d4 := int32(q[4]) - int32(vectors[base+4])
	d5 := int32(q[5]) - int32(vectors[base+5])
	d6 := int32(q[6]) - int32(vectors[base+6])
	d7 := int32(q[7]) - int32(vectors[base+7])
	d8 := int32(q[8]) - int32(vectors[base+8])
	d9 := int32(q[9]) - int32(vectors[base+9])
	d10 := int32(q[10]) - int32(vectors[base+10])
	d11 := int32(q[11]) - int32(vectors[base+11])
	d12 := int32(q[12]) - int32(vectors[base+12])
	d13 := int32(q[13]) - int32(vectors[base+13])
	return d0*d0 + d1*d1 + d2*d2 + d3*d3 + d4*d4 + d5*d5 + d6*d6 +
		d7*d7 + d8*d8 + d9*d9 + d10*d10 + d11*d11 + d12*d12 + d13*d13
}
