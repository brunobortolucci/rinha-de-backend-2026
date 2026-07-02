package brute

import (
	"github.com/brunobortolucci/rinha-de-backend-2026/internal/topk"
)

const Dims = 14

// Scan16 faz a varredura linear sobre vetores int16.
// Usado como oráculo de exatidão nos testes e diagnósticos do IVF.
func Scan16(vectors []int16, labels []byte, q [Dims]int16, k int) *topk.Heap {
	h := topk.New(k)
	count := len(vectors) / Dims
	for i := 0; i < count; i++ {
		base := i * Dims
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
		dist := d0*d0 + d1*d1 + d2*d2 + d3*d3 + d4*d4 + d5*d5 + d6*d6 +
			d7*d7 + d8*d8 + d9*d9 + d10*d10 + d11*d11 + d12*d12 + d13*d13

		if h.Full() && dist >= h.Worst() {
			continue
		}
		h.Push(topk.Neighbor{Dist: dist, Label: labels[i]})
	}
	return h
}
