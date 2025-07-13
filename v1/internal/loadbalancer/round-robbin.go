package loadbalancer

import "sync"

type RoundRobbin struct {
	idx  int
	size int

	mu sync.Mutex
}

func (r *RoundRobbin) Next() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.idx >= r.size-1 {
		r.idx = 0
	} else {
		r.idx++
	}
	return r.idx
}

func (r *RoundRobbin) SetSize(size int) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.size = size
}

func (r *RoundRobbin) GetSize() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.size
}
