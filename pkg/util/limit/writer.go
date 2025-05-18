package limit

import (
	"context"
	"golang.org/x/time/rate"
	"io"
	"sync"
)

type Writer struct {
	w       io.Writer
	limiter *rate.Limiter
	ctx     context.Context
	mux     sync.Mutex
}

// NewWriter returns a writer that implements io.Writer with rate limiting.
func NewWriter(w io.Writer, limiter *rate.Limiter) *Writer {
	return &Writer{
		w:       w,
		limiter: limiter,
		ctx:     context.Background(),
		mux:     sync.Mutex{},
	}
}

// Write writes bytes from p.
func (s *Writer) Write(p []byte) (int, error) {
	s.mux.Lock()
	defer s.mux.Unlock()

	if s.limiter == nil {
		return s.w.Write(p)
	}
	n, err := s.w.Write(p)
	if err != nil {
		return n, err
	}
	if err := s.limiter.WaitN(s.ctx, n); err != nil {
		return n, err
	}
	return n, err
}
