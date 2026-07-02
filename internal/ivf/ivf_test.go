package ivf

import (
	"encoding/json"
	"math/rand"
	"os"
	"path/filepath"
	"sort"
	"testing"

	"github.com/brunobortolucci/rinha-de-backend-2026/internal/brute"
	"github.com/brunobortolucci/rinha-de-backend-2026/internal/topk"
)

type sampleRef struct {
	Vector [Dims]float64 `json:"vector"`
	Label  string        `json:"label"`
}

func loadSampleVectors(t *testing.T) ([]int16, []byte) {
	t.Helper()
	path := filepath.Join("..", "..", "resources", "example-references.json")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ler %s: %v", path, err)
	}
	var refs []sampleRef
	if err := json.Unmarshal(data, &refs); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	vectors := make([]int16, 0, len(refs)*Dims)
	labels := make([]byte, 0, len(refs))
	for _, r := range refs {
		for i := 0; i < Dims; i++ {
			vectors = append(vectors, Quantize(r.Vector[i]))
		}
		switch r.Label {
		case "legit":
			labels = append(labels, 0)
		case "fraud":
			labels = append(labels, 1)
		default:
			t.Fatalf("label inesperado: %q", r.Label)
		}
	}
	return vectors, labels
}

// reordena vetores e labels pelo perm, como o preprocess faz ao gravar
func reorder(vectors []int16, labels []byte, perm []uint32) ([]int16, []byte) {
	outV := make([]int16, len(vectors))
	outL := make([]byte, len(labels))
	for pos, orig := range perm {
		copy(outV[pos*Dims:(pos+1)*Dims], vectors[int(orig)*Dims:int(orig+1)*Dims])
		outL[pos] = labels[orig]
	}
	return outV, outL
}

func TestBuild_PermIsPermutation(t *testing.T) {
	vectors, _ := loadSampleVectors(t)
	count := len(vectors) / Dims

	const nlist = 16
	centroids, offsets, perm := Build(vectors, nlist)

	if len(centroids) != nlist*Dims {
		t.Fatalf("esperava %d valores de centróide, obteve %d", nlist*Dims, len(centroids))
	}
	if len(offsets) != nlist+1 {
		t.Fatalf("esperava %d offsets, obteve %d", nlist+1, len(offsets))
	}
	if offsets[0] != 0 || int(offsets[nlist]) != count {
		t.Fatalf("offsets não cobrem [0,%d]: início=%d fim=%d", count, offsets[0], offsets[nlist])
	}
	for c := 0; c < nlist; c++ {
		if offsets[c] > offsets[c+1] {
			t.Fatalf("offsets não-monotônicos no cluster %d", c)
		}
	}

	seen := make([]bool, count)
	for _, orig := range perm {
		if int(orig) >= count {
			t.Fatalf("perm fora do range: %d", orig)
		}
		if seen[orig] {
			t.Fatalf("índice %d aparece duas vezes no perm", orig)
		}
		seen[orig] = true
	}
}

func TestBuild_EachVectorInNearestCluster(t *testing.T) {
	vectors, _ := loadSampleVectors(t)

	const nlist = 16
	centroids, offsets, perm := Build(vectors, nlist)

	for c := 0; c < nlist; c++ {
		for p := offsets[c]; p < offsets[c+1]; p++ {
			base := int(perm[p]) * Dims
			var q [Dims]int16
			copy(q[:], vectors[base:base+Dims])
			myDist := distToQuery(centroids, c*Dims, q)
			for other := 0; other < nlist; other++ {
				if d := distToQuery(centroids, other*Dims, q); d < myDist {
					t.Fatalf("vetor %d atribuído ao cluster %d (dist=%d) mas cluster %d está mais perto (dist=%d)",
						perm[p], c, myDist, other, d)
				}
			}
		}
	}
}

func TestSearch_AllProbesMatchesBruteForce(t *testing.T) {
	vectors, labels := loadSampleVectors(t)

	const nlist = 16
	centroids, offsets, perm := Build(vectors, nlist)
	ordV, ordL := reorder(vectors, labels, perm)

	rng := rand.New(rand.NewSource(42))
	const queries = 200
	const k = 5

	for qi := 0; qi < queries; qi++ {
		var query [Dims]int16
		for i := 0; i < Dims; i++ {
			query[i] = int16(rng.Intn(2*4095+1) - 4095)
		}

		got := Search(centroids, offsets, ordV, ordL, query, k, nlist)
		want := brute.Scan16(ordV, ordL, query, k)

		gotDists := extractDists(got)
		wantDists := extractDists(want)
		for i := range gotDists {
			if gotDists[i] != wantDists[i] {
				t.Fatalf("query %d: dists divergem\n  ivf:   %v\n  brute: %v", qi, gotDists, wantDists)
			}
		}
	}
}

func extractDists(h *topk.Heap) []int32 {
	winners := h.Drain()
	out := make([]int32, len(winners))
	for i, n := range winners {
		out[i] = n.Dist
	}
	sort.Slice(out, func(i, j int) bool { return out[i] < out[j] })
	return out
}
