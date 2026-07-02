// Package ivf implementa um índice IVF (Inverted File): os vetores são
// agrupados em clusters por k-means no pré-processamento e, em runtime,
// a busca faz brute force apenas dentro dos nprobe clusters mais próximos
// da query. Busca aproximada: recall controlado por nprobe.
package ivf

const (
	Dims  = 14
	Scale = 4095.0
)

func Quantize(x float64) int16 {
	v := x * Scale
	if v > Scale {
		v = Scale
	}
	if v < -Scale {
		v = -Scale
	}
	if v >= 0 {
		return int16(v + 0.5)
	}
	return int16(v - 0.5)
}

func distSquared(a []int16, baseA int, b []int16, baseB int) int32 {
	d0 := int32(a[baseA+0]) - int32(b[baseB+0])
	d1 := int32(a[baseA+1]) - int32(b[baseB+1])
	d2 := int32(a[baseA+2]) - int32(b[baseB+2])
	d3 := int32(a[baseA+3]) - int32(b[baseB+3])
	d4 := int32(a[baseA+4]) - int32(b[baseB+4])
	d5 := int32(a[baseA+5]) - int32(b[baseB+5])
	d6 := int32(a[baseA+6]) - int32(b[baseB+6])
	d7 := int32(a[baseA+7]) - int32(b[baseB+7])
	d8 := int32(a[baseA+8]) - int32(b[baseB+8])
	d9 := int32(a[baseA+9]) - int32(b[baseB+9])
	d10 := int32(a[baseA+10]) - int32(b[baseB+10])
	d11 := int32(a[baseA+11]) - int32(b[baseB+11])
	d12 := int32(a[baseA+12]) - int32(b[baseB+12])
	d13 := int32(a[baseA+13]) - int32(b[baseB+13])
	return d0*d0 + d1*d1 + d2*d2 + d3*d3 + d4*d4 + d5*d5 + d6*d6 +
		d7*d7 + d8*d8 + d9*d9 + d10*d10 + d11*d11 + d12*d12 + d13*d13
}
