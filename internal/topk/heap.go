package topk

type Neighbor struct {
	Dist  int32
	Label byte
}

type Heap struct {
	items []Neighbor
	k     int
	size  int
}

func New(k int) *Heap {
	return &Heap{items: make([]Neighbor, k), k: k}
}

func (h *Heap) Worst() int32 {
	return h.items[0].Dist
}

func (h *Heap) Full() bool {
	return h.size == h.k
}

func (h *Heap) Push(n Neighbor) {
	if !h.Full() {
		h.items[h.size] = n
		h.size++
		h.siftUp(h.size - 1)
		return
	}
	if n.Dist >= h.items[0].Dist {
		return
	}
	h.items[0] = n
	h.siftDown(0)
}

func (h *Heap) siftUp(i int) {
	for i > 0 {
		parent := (i - 1) / 2
		if h.items[i].Dist <= h.items[parent].Dist {
			return
		}
		h.items[i], h.items[parent] = h.items[parent], h.items[i]
		i = parent
	}
}

func (h *Heap) siftDown(i int) {
	n := h.size
	for {
		left := 2*i + 1
		right := 2*i + 2
		largest := i
		if left < n && h.items[left].Dist > h.items[largest].Dist {
			largest = left
		}
		if right < n && h.items[right].Dist > h.items[largest].Dist {
			largest = right
		}
		if largest == i {
			return
		}
		h.items[i], h.items[largest] = h.items[largest], h.items[i]
		i = largest
	}
}

func (h *Heap) Drain() []Neighbor {
	return h.items[:h.size]
}

func (h *Heap) Reset() {
	h.size = 0
}
