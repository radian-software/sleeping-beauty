package sleepingd

import (
	"fmt"
	"io"
	"net"
	"sync"
)

type lazyConn struct {
	connGetter  func() (net.Conn, error)
	openOnRead  bool
	openOnWrite bool

	opened sync.WaitGroup
	closed bool
	lock   sync.Mutex
	conn   net.Conn
}

type LazyConn interface {
	Read(p []byte) (int, error)
	Write(p []byte) (int, error)
	Close() error
}

func NewLazyConn(connGetter func() (net.Conn, error), openOnRead bool, openOnWrite bool) LazyConn {
	lc := lazyConn{
		connGetter:  connGetter,
		openOnRead:  openOnRead,
		openOnWrite: openOnWrite,
	}
	lc.opened.Add(1)
	return &lc
}

func (lc *lazyConn) Read(p []byte) (int, error) {
	if lc.conn == nil {
		lc.lock.Lock()
		if lc.closed {
			return 0, fmt.Errorf("use of closed connection")
		}
		if !lc.openOnRead {
			lc.opened.Wait()
		}
		conn, err := lc.connGetter()
		if err != nil {
			lc.lock.Unlock()
			return 0, err
		}
		lc.conn = conn
		lc.opened.Done()
		lc.lock.Unlock()
	}
	return lc.conn.Read(p)
}

func (lc *lazyConn) Write(p []byte) (int, error) {
	if lc.conn == nil {
		lc.lock.Lock()
		if lc.closed {
			return 0, fmt.Errorf("use of closed connection")
		}
		if !lc.openOnWrite {
			lc.opened.Wait()
		}
		conn, err := lc.connGetter()
		if err != nil {
			lc.lock.Unlock()
			return 0, err
		}
		lc.conn = conn
		lc.opened.Done()
		lc.lock.Unlock()
	}
	return lc.conn.Write(p)
}

func (lc *lazyConn) Close() error {
	lc.lock.Lock()
	lc.closed = true
	lc.lock.Unlock()
	if lc.conn == nil {
		return nil
	}
	return lc.conn.Close()
}

func CopyWithActivity(dst io.Writer, src io.Reader, activityCh chan<- struct{}) error {
	buf := make([]byte, 32*1024)
	// Implementation based on copyBuffer in io from stdlib
	for {
		nr, err := src.Read(buf)
		if err == io.EOF {
			return nil
		} else if err != nil {
			return err
		} else if nr == 0 {
			continue
		}
		activityCh <- struct{}{}
		_, err = dst.Write(buf[0:nr])
		if err != nil {
			return err
		}
		activityCh <- struct{}{}
	}
}
