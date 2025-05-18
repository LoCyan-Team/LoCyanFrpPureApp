package limit

import (
	"context"
	"golang.org/x/time/rate"
	"io"
	"sync"
)

type Reader struct {
	r       io.Reader
	limiter *rate.Limiter
	ctx     context.Context
	mux     sync.Mutex
}

// NewReader returns a reader that implements io.Reader with rate limiting.
func NewReader(r io.Reader, limiter *rate.Limiter) *Reader {
	return &Reader{
		r:       r,
		limiter: limiter,
		ctx:     context.Background(),
		mux:     sync.Mutex{},
	}
}

// Read reads bytes into p.
func (s *Reader) Read(p []byte) (int, error) {
	s.mux.Lock()
	defer s.mux.Unlock()

	if s.limiter == nil {
		return s.r.Read(p)
	}
	n, err := s.r.Read(p)
	if err != nil {
		return n, err
	}
	if err := s.limiter.WaitN(s.ctx, n); err != nil {
		return n, err
	}
	return n, nil
}
