package limit

import (
	"golang.org/x/time/rate"
	"io"
	"net"
)

const (
	B uint64 = 1 << (10 * (iota))
	KB
	MB
	GB
	TB
	PB
	EB
)

type LimitConn struct {
	net.Conn

	lr *rate.Limiter
	lw *rate.Limiter
}

func NewLimitConn(maxRead, maxWrite int64, c net.Conn) LimitConn {
	// 此处抄袭官方库
	lr := rate.NewLimiter(rate.Limit(float64(maxRead)), int(maxRead))
	lw := rate.NewLimiter(rate.Limit(float64(maxWrite)), int(maxWrite))
	return LimitConn{
		lr:   lr,
		lw:   lw,
		Conn: c,
	}
}

func (c LimitConn) Read(p []byte) (n int, err error) {
	return c.lr.Read(p)
}

func (c LimitConn) Write(p []byte) (n int, err error) {
	return c.lw.Write(p)
}
