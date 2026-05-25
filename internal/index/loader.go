package index

import (
	"encoding/binary"
	"fmt"
	"os"
	"syscall"
	"unsafe"
)

const (
	magic       = "RINHA001"
	version     = uint32(1)
	dims        = 14
	headerBytes = 16
	floatBytes  = 4
)

type Index struct {
	Vectors []float32
	Labels  []byte
	Count   int

	raw []byte
}

func Load(path string) (*Index, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("abrir %s: %w", path, err)
	}
	defer func(f *os.File) {
		err := f.Close()
		if err != nil {

		}
	}(f)

	info, err := f.Stat()
	if err != nil {
		return nil, fmt.Errorf("stat %s: %w", path, err)
	}
	size := int(info.Size())
	if size < headerBytes {
		return nil, fmt.Errorf("arquivo menor que o header (%d bytes)", size)
	}

	data, err := syscall.Mmap(int(f.Fd()), 0, size, syscall.PROT_READ, syscall.MAP_SHARED)
	if err != nil {
		return nil, fmt.Errorf("mmap: %w", err)
	}

	if string(data[0:8]) != magic {
		_ = syscall.Munmap(data)
		return nil, fmt.Errorf("magic inválido: %q", string(data[0:8]))
	}
	gotVersion := binary.LittleEndian.Uint32(data[8:12])
	if gotVersion != version {
		_ = syscall.Munmap(data)
		return nil, fmt.Errorf("versão inválida: %d (esperava %d)", gotVersion, version)
	}
	count := int(binary.LittleEndian.Uint32(data[12:16]))

	vecBytes := count * dims * floatBytes
	expected := headerBytes + vecBytes + count
	if size != expected {
		_ = syscall.Munmap(data)
		return nil, fmt.Errorf("tamanho inválido: %d (esperava %d para count=%d)", size, expected, count)
	}

	vectors := unsafe.Slice((*float32)(unsafe.Pointer(&data[headerBytes])), count*dims)
	labels := data[headerBytes+vecBytes:]

	return &Index{
		Vectors: vectors,
		Labels:  labels,
		Count:   count,
		raw:     data,
	}, nil
}

func (i *Index) Close() error {
	if i.raw == nil {
		return nil
	}
	err := syscall.Munmap(i.raw)
	i.raw = nil
	i.Vectors = nil
	i.Labels = nil
	return err
}
