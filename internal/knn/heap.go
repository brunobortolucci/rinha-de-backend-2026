package knn

type neighbor struct {
	dist  float32
	label byte
}

type topK struct {
	items []neighbor
	k     int
	size  int
}

func newTopK(k int) *topK {
	return &topK{items: make([]neighbor, k), k: k}
}

func (h *topK) worst() float32 {
	return h.items[0].dist
}

func (h *topK) full() bool {
	return h.size == h.k
}

func (h *topK) push(n neighbor) {
	if !h.full() {
		h.items[h.size] = n
		h.size++
		h.siftUp(h.size - 1)
		return
	}
	if n.dist >= h.items[0].dist {
		return
	}
	h.items[0] = n
	h.siftDown(0)
}

func (h *topK) siftUp(i int) {
	for i > 0 {
		parent := (i - 1) / 2
		if h.items[i].dist <= h.items[parent].dist {
			return
		}
		h.items[i], h.items[parent] = h.items[parent], h.items[i]
		i = parent
	}
}

func (h *topK) siftDown(i int) {
	n := h.size
	for {
		left := 2*i + 1
		right := 2*i + 2
		largest := i
		if left < n && h.items[left].dist > h.items[largest].dist {
			largest = left
		}
		if right < n && h.items[right].dist > h.items[largest].dist {
			largest = right
		}
		if largest == i {
			return
		}
		h.items[i], h.items[largest] = h.items[largest], h.items[i]
		i = largest
	}
}

func (h *topK) drain() []neighbor {
	return h.items[:h.size]
}
