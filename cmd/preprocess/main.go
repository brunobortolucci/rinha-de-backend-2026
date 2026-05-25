package main

import (
	"bufio"
	"compress/gzip"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"os"
)

const (
	inputPath  = "resources/references.json.gz"
	outputPath = "resources/references.bin"
	magic      = "RINHA001"
	version    = uint32(1)
	dims       = 14
	labelLegit = byte(0)
	labelFraud = byte(1)
)

type reference struct {
	Vector [dims]float32 `json:"vector"`
	Label  string        `json:"label"`
}

func main() {
	if err := run(); err != nil {
		_, err := fmt.Fprintln(os.Stderr, "preprocess:", err)
		if err != nil {
			return
		}
		os.Exit(1)
	}
}

func run() error {
	in, err := os.Open(inputPath)
	if err != nil {
		return fmt.Errorf("abrir %s: %w", inputPath, err)
	}
	defer func(in *os.File) {
		err := in.Close()
		if err != nil {

		}
	}(in)

	gz, err := gzip.NewReader(in)
	if err != nil {
		return fmt.Errorf("gzip: %w", err)
	}
	defer func(gz *gzip.Reader) {
		err := gz.Close()
		if err != nil {

		}
	}(gz)

	out, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("criar %s: %w", outputPath, err)
	}
	defer func(out *os.File) {
		err := out.Close()
		if err != nil {

		}
	}(out)

	bw := bufio.NewWriterSize(out, 1<<20)

	if _, err := bw.Write(make([]byte, 16)); err != nil {
		return fmt.Errorf("reservar header: %w", err)
	}

	dec := json.NewDecoder(gz)

	tok, err := dec.Token()
	if err != nil {
		return fmt.Errorf("token inicial: %w", err)
	}
	if delim, ok := tok.(json.Delim); !ok || delim != '[' {
		return fmt.Errorf("esperava '[' no início, recebi %v", tok)
	}

	labels := make([]byte, 0, 3_000_000)
	vecBuf := make([]byte, dims*4)
	var count uint32

	for dec.More() {
		var ref reference
		if err := dec.Decode(&ref); err != nil {
			return fmt.Errorf("decode no item %d: %w", count, err)
		}

		for i := 0; i < dims; i++ {
			bits := math.Float32bits(ref.Vector[i])
			binary.LittleEndian.PutUint32(vecBuf[i*4:], bits)
		}
		if _, err := bw.Write(vecBuf); err != nil {
			return fmt.Errorf("escrever vetor %d: %w", count, err)
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

	if _, err := bw.Write(labels); err != nil {
		return fmt.Errorf("escrever labels: %w", err)
	}

	if err := bw.Flush(); err != nil {
		return fmt.Errorf("flush: %w", err)
	}

	header := make([]byte, 16)
	copy(header[0:8], magic)
	binary.LittleEndian.PutUint32(header[8:12], version)
	binary.LittleEndian.PutUint32(header[12:16], count)
	if _, err := out.WriteAt(header, 0); err != nil {
		return fmt.Errorf("escrever header: %w", err)
	}

	if err := out.Sync(); err != nil {
		return fmt.Errorf("sync: %w", err)
	}

	info, err := out.Stat()
	if err != nil {
		return fmt.Errorf("stat: %w", err)
	}
	fmt.Printf("preprocess: %d vetores, %d bytes em %s\n", count, info.Size(), outputPath)

	if _, err := io.Copy(io.Discard, gz); err != nil {
		return fmt.Errorf("drenar gzip: %w", err)
	}
	return nil
}
