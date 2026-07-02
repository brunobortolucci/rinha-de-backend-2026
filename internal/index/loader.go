package index

import (
	"encoding/binary"
	"fmt"
	"os"
	"syscall"
	"unsafe"
)

// Formato v4 (RINHA004), little-endian, vetores int16 reordenados por cluster IVF:
//
//	[0:8]   magic "RINHA004"
//	[8:12]  version = 4
//	[12:16] count
//	[16:20] nlist
//	centroids  nlist × dims × int16
//	offsets    (nlist+1) × uint32
//	vectors    count × dims × int16
//	labels     count × byte
const (
	magic       = "RINHA004"
	version     = uint32(4)
	dims        = 14
	headerBytes = 20
)

type Index struct {
	Centroids []int16
	Offsets   []uint32
	Vectors   []int16
	Labels    []byte
	Count     int
	Nlist     int

	raw []byte
}

func Load(path string) (*Index, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("abrir %s: %w", path, err)
	}
	defer f.Close()

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
	nlist := int(binary.LittleEndian.Uint32(data[16:20]))

	centroidBytes := nlist * dims * 2
	offsetBytes := (nlist + 1) * 4
	vecBytes := count * dims * 2
	expected := headerBytes + centroidBytes + offsetBytes + vecBytes + count
	if size != expected {
		_ = syscall.Munmap(data)
		return nil, fmt.Errorf("tamanho inválido: %d (esperava %d para count=%d nlist=%d)", size, expected, count, nlist)
	}

	centroidOffset := headerBytes
	offsetOffset := centroidOffset + centroidBytes
	vecOffset := offsetOffset + offsetBytes
	labelOffset := vecOffset + vecBytes

	centroids := unsafe.Slice((*int16)(unsafe.Pointer(&data[centroidOffset])), nlist*dims)
	offsets := unsafe.Slice((*uint32)(unsafe.Pointer(&data[offsetOffset])), nlist+1)
	vectors := unsafe.Slice((*int16)(unsafe.Pointer(&data[vecOffset])), count*dims)
	labels := data[labelOffset : labelOffset+count]

	return &Index{
		Centroids: centroids,
		Offsets:   offsets,
		Vectors:   vectors,
		Labels:    labels,
		Count:     count,
		Nlist:     nlist,
		raw:       data,
	}, nil
}

func (i *Index) Warmup() {
	if i.raw == nil {
		return
	}
	var sum byte
	for _, b := range i.raw {
		sum ^= b
	}
	_ = sum
}

func (i *Index) Close() error {
	if i.raw == nil {
		return nil
	}
	err := syscall.Munmap(i.raw)
	i.raw = nil
	i.Centroids = nil
	i.Offsets = nil
	i.Vectors = nil
	i.Labels = nil
	return err
}
