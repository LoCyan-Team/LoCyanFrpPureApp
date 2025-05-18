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

const BurstLimit = 1024 * 1024 * 1024

type LimitConn struct {
	net.Conn

	lr io.Reader
	lw io.Writer
}

func NewLimitConn(maxRead, maxWrite uint64, c net.Conn) LimitConn {
	// 这里不知道为什么要 49 才能对的上真实速度
	// 49 是根据 wget 速度来取的，测试了 512、1024、2048、4096、8192 等多种速度下都很准确
	lr := rate.NewLimiter(rate.Limit(maxRead*49), BurstLimit)
	lw := rate.NewLimiter(rate.Limit(maxWrite*49), BurstLimit)
	return LimitConn{
		lr:   NewReader(c, lr),
		lw:   NewWriter(c, lw),
		Conn: c,
	}
}

func (c LimitConn) Read(p []byte) (n int, err error) {
	return c.lr.Read(p)
}

func (c LimitConn) Write(p []byte) (n int, err error) {
	return c.lw.Write(p)
}
