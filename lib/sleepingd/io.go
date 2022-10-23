package sleepingd

import (
	"io"
	"net"
	"sync"
)

type closedConn struct {
	readError  error
	writeError error
}

func (bp *closedConn) Read(p []byte) (int, error) {
	return 0, bp.readError
}

func (bp *closedConn) Write(p []byte) (int, error) {
	return 0, bp.writeError
}

func (bp *closedConn) Close() error {
	return nil
}

type simpleConn interface {
	Read(b []byte) (int, error)
	Write(b []byte) (int, error)
	Close() error
}

type lazyConn struct {
	connGetter  func() (net.Conn, error)
	openOnRead  bool
	openOnWrite bool

	opened sync.WaitGroup
	lock   sync.Mutex
	conn   simpleConn
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

func (lc *lazyConn) ensureOpen() error {
	lc.lock.Lock()
	defer lc.lock.Unlock()
	if lc.conn == nil {
		conn, err := lc.connGetter()
		if err != nil {
			return err
		}
		lc.conn = conn
		lc.opened.Done()
	}
	return nil
}

func (lc *lazyConn) Read(p []byte) (int, error) {
	if !lc.openOnRead {
		lc.opened.Wait()
	}
	err := lc.ensureOpen()
	if err != nil {
		return 0, err
	}
	return lc.conn.Read(p)
}

func (lc *lazyConn) Write(p []byte) (int, error) {
	if !lc.openOnWrite {
		lc.opened.Wait()
	}
	err := lc.ensureOpen()
	if err != nil {
		return 0, err
	}
	return lc.conn.Write(p)
}

func (lc *lazyConn) Close() error {
	lc.lock.Lock()
	if lc.conn == nil {
		lc.conn = &closedConn{}
		lc.opened.Done()
	}
	lc.lock.Unlock()
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
