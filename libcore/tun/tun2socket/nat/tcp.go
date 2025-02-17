package nat

import (
	"libcore/tun"
	"net"
	"syscall"
	"time"
)

type TCP struct {
	listener *net.TCPListener
	table    *table
}

type conn struct {
	net.Conn

	tuple tuple
}

func (t *TCP) Accept() (net.Conn, error) {
	c, err := t.listener.AcceptTCP()
	if err != nil {
		return nil, err
	}

	addr := c.RemoteAddr().(*net.TCPAddr)
	tup := t.table.tupleOf(uint16(addr.Port))
	if (!addr.IP.Equal(tun.PRIVATE_VLAN4_ROUTER_IP) && !addr.IP.Equal(tun.PRIVATE_VLAN6_ROUTER_IP)) || tup == zeroTuple {
		_ = c.Close()

		return nil, net.InvalidAddrError("unknown remote addr")
	}

	_ = c.SetKeepAlive(false)

	sys, err := c.SyscallConn()
	if err == nil {
		_ = sys.Control(func(fd uintptr) {
			_ = syscall.SetsockoptInt(int(fd), syscall.SOL_SOCKET, syscall.SO_NO_CHECK, 1)
		})
	}

	return &conn{
		Conn:  c,
		tuple: tup,
	}, nil
}

func (t *TCP) Close() error {
	return t.listener.Close()
}

func (t *TCP) Addr() net.Addr {
	return t.listener.Addr()
}

func (t *TCP) SetDeadline(time time.Time) error {
	return t.listener.SetDeadline(time)
}

func (c *conn) LocalAddr() net.Addr {
	return &net.TCPAddr{
		IP:   c.tuple.SourceIP[:],
		Port: int(c.tuple.SourcePort),
	}
}

func (c *conn) RemoteAddr() net.Addr {
	return &net.TCPAddr{
		IP:   c.tuple.DestinationIP[:],
		Port: int(c.tuple.DestinationPort),
	}
}
