package common

import (
	"math"
	"sync"
)

type LevelPool struct {
	size int
	pool sync.Pool
}

func newLevelPool(size int) *LevelPool {
	return &LevelPool{
		size: size,
		pool: sync.Pool{
			New: func() interface{} {
				data := make([]byte, size)
				return data
			},
		},
	}
}

type LimitedPool struct {
	minSize int
	maxSize int
	pools   []*LevelPool
}

func NewLimitedPool(minSize, maxSize int) *LimitedPool {
	if maxSize < minSize {
		panic("maxSize can't be less than minSize")
	}
	const multiplier = 2
	var pools []*LevelPool
	curSize := minSize
	for curSize < maxSize {
		pools = append(pools, newLevelPool(curSize))
		curSize *= multiplier
	}
	pools = append(pools, newLevelPool(maxSize))
	return &LimitedPool{
		minSize: minSize,
		maxSize: maxSize,
		pools:   pools,
	}
}

func (p *LimitedPool) findPool(size int) *LevelPool {
	if size > p.maxSize {
		return nil
	}
	idx := int(math.Ceil(math.Log2(float64(size) / float64(p.minSize))))
	if idx < 0 {
		idx = 0
	}
	if idx > len(p.pools)-1 {
		return nil
	}
	return p.pools[idx]
}

func (p *LimitedPool) findPutPool(size int) *LevelPool {
	if size > p.maxSize {
		return nil
	}
	if size < p.minSize {
		return nil
	}

	idx := int(math.Floor(math.Log2(float64(size) / float64(p.minSize))))
	if idx < 0 {
		idx = 0
	}
	if idx > len(p.pools)-1 {
		return nil
	}
	return p.pools[idx]
}

func (p *LimitedPool) Get(size int) []byte {
	sp := p.findPool(size)
	if sp == nil {
		data := make([]byte, size)
		return data
	}
	buf := sp.pool.Get().([]byte)
	buf = (buf)[:size]
	return buf
}

func (p *LimitedPool) Put(b []byte) {
	sp := p.findPutPool(cap(b))
	if sp == nil {
		return
	}
	b = b[:cap(b)]
	sp.pool.Put(b)
}
