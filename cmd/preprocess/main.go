package main

import (
	"bufio"
	"compress/gzip"
	"encoding/binary"
	"fmt"
	"os"
	"time"
	"unsafe"

	"encoding/json"

	"github.com/brunobortolucci/rinha-de-backend-2026/internal/ivf"
)

const (
	inputPath  = "resources/references.json.gz"
	outputPath = "resources/references.bin"
	magic      = "RINHA004"
	version    = uint32(4)
	dims       = ivf.Dims
	nlist      = 1024
	labelLegit = byte(0)
	labelFraud = byte(1)
)

type reference struct {
	Vector [dims]float64 `json:"vector"`
	Label  string        `json:"label"`
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "preprocess:", err)
		os.Exit(1)
	}
}

func run() error {
	in, err := os.Open(inputPath)
	if err != nil {
		return fmt.Errorf("abrir %s: %w", inputPath, err)
	}
	defer in.Close()

	gz, err := gzip.NewReader(in)
	if err != nil {
		return fmt.Errorf("gzip: %w", err)
	}
	defer gz.Close()

	vectors := make([]int16, 0, 3_000_000*dims)
	labels := make([]byte, 0, 3_000_000)

	dec := json.NewDecoder(gz)
	tok, err := dec.Token()
	if err != nil {
		return fmt.Errorf("token inicial: %w", err)
	}
	if delim, ok := tok.(json.Delim); !ok || delim != '[' {
		return fmt.Errorf("esperava '[' no início, recebi %v", tok)
	}

	var count uint32
	for dec.More() {
		var ref reference
		if err := dec.Decode(&ref); err != nil {
			return fmt.Errorf("decode no item %d: %w", count, err)
		}

		for i := 0; i < dims; i++ {
			vectors = append(vectors, ivf.Quantize(ref.Vector[i]))
		}

		switch ref.Label {
		case "legit":
			labels = append(labels, labelLegit)
		case "fraud":
			labels = append(labels, labelFraud)
		default:
			return fmt.Errorf("label desconhecido no item %d: %q", count, ref.Label)
		}

		count++
	}

	fmt.Printf("preprocess: %d vetores lidos, rodando k-means (nlist=%d)...\n", count, nlist)
	t0 := time.Now()
	centroids, offsets, perm := ivf.Build(vectors, nlist)
	fmt.Printf("preprocess: IVF construído em %s (%d clusters)\n", time.Since(t0).Round(time.Second), len(offsets)-1)

	out, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("criar %s: %w", outputPath, err)
	}
	defer out.Close()
	w := bufio.NewWriterSize(out, 1<<20)

	header := make([]byte, 20)
	copy(header[0:8], magic)
	binary.LittleEndian.PutUint32(header[8:12], version)
	binary.LittleEndian.PutUint32(header[12:16], count)
	binary.LittleEndian.PutUint32(header[16:20], uint32(len(offsets)-1))
	if _, err := w.Write(header); err != nil {
		return fmt.Errorf("escrever header: %w", err)
	}

	if err := writeInt16s(w, centroids); err != nil {
		return fmt.Errorf("escrever centróides: %w", err)
	}

	var buf4 [4]byte
	for _, o := range offsets {
		binary.LittleEndian.PutUint32(buf4[:], o)
		if _, err := w.Write(buf4[:]); err != nil {
			return fmt.Errorf("escrever offsets: %w", err)
		}
	}

	for _, orig := range perm {
		base := int(orig) * dims
		if err := writeInt16s(w, vectors[base:base+dims]); err != nil {
			return fmt.Errorf("escrever vetores: %w", err)
		}
	}
	for _, orig := range perm {
		if err := w.WriteByte(labels[orig]); err != nil {
			return fmt.Errorf("escrever labels: %w", err)
		}
	}

	if err := w.Flush(); err != nil {
		return fmt.Errorf("flush: %w", err)
	}
	if err := out.Sync(); err != nil {
		return fmt.Errorf("sync: %w", err)
	}

	info, err := out.Stat()
	if err != nil {
		return fmt.Errorf("stat: %w", err)
	}
	fmt.Printf("preprocess: %d bytes em %s (count=%d nlist=%d)\n",
		info.Size(), outputPath, count, len(offsets)-1)
	return nil
}

func writeInt16s(w *bufio.Writer, v []int16) error {
	b := unsafe.Slice((*byte)(unsafe.Pointer(&v[0])), len(v)*2)
	_, err := w.Write(b)
	return err
}
