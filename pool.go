package gopetri

import (
	"time"
)

// DefaultPoolTimeout is default timeout for poll get.
const DefaultPoolTimeout time.Duration = 30 * time.Second

// Pool is pool of net.
type Pool struct {
	size       int
	getTimeout time.Duration
	pool       chan PoolNet
}

// NewPool init pool.
func NewPool(size int, getTimeout time.Duration) *Pool {
	return &Pool{size: size, getTimeout: getTimeout}
}

// Init pool.
func (p *Pool) Init(cfg Cfg) error {
	if p.pool != nil {
		return NewError(ErrCodePoolAlreadyInit, "Pool of net already init")
	}
	p.pool = make(chan PoolNet, p.size)
	for i := 0; i < p.size; i++ {
		net, err := BuildFromCfg(cfg)
		if err != nil {
			return err
		}
		p.pool <- PoolNet{Net: net, releaseNet: p.releaseNet}
	}
	return nil
}

func (p *Pool) releaseNet(net *PoolNet) {
	p.pool <- *net
}

// Get return net.
func (p *Pool) Get() (*PoolNet, error) {
	select {
	case net := <-p.pool:
		return &net, nil
	case <-time.After(p.getTimeout):
		return nil, NewError(ErrCodeWaitingForNetFromPoolTooLong, "Waiting for net from pool too long")
	}
}

// PoolNet is wrapper of net for pool.
type PoolNet struct {
	*Net
	releaseNet func(net *PoolNet)
}

// Close current net and return it to pool.
func (n *PoolNet) Close() {
	n.FullReset()
	n.releaseNet(n)
}
