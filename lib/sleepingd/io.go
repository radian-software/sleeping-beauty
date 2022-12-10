package sleepingd

import (
	"fmt"
	"io"
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

type SimpleConn interface {
	Read(b []byte) (int, error)
	Write(b []byte) (int, error)
	Close() error
}

type lazyConn struct {
	connGetter  func() (SimpleConn, error)
	openOnRead  bool
	openOnWrite bool

	opened sync.WaitGroup
	lock   sync.Mutex
	conn   SimpleConn
}

type LazyConn interface {
	Read(p []byte) (int, error)
	Write(p []byte) (int, error)
	Close() error
}

func NewLazyConn(connGetter func() (SimpleConn, error), openOnRead bool, openOnWrite bool) LazyConn {
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
			lc.conn = &closedConn{
				readError:  err,
				writeError: err,
			}
			lc.opened.Done()
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
		lc.conn = &closedConn{
			readError:  fmt.Errorf("use of lazy connection that was closed before being initialized"),
			writeError: fmt.Errorf("use of lazy connection that was closed before being initialized"),
		}
		lc.opened.Done()
		lc.lock.Unlock()
	} else {
		// This clause is not strictly necessary; we could
		// instead continue to proxy Read and Write calls down
		// to the underlying connection, which is required by
		// contract to return errors if it has been closed
		// already. The only thing that will change is the
		// returned error types. However, if we explicitly
		// return an error rather than calling down, it makes
		// mocking a little easier for unit tests, so we do
		// that.
		defer func() {
			lc.conn = &closedConn{
				readError:  fmt.Errorf("use of lazy connection that was closed by caller already"),
				writeError: fmt.Errorf("use of lazy connection that was closed by caller already"),
			}
			lc.lock.Unlock()
		}()
	}
	return lc.conn.Close()
}

// CopyWithActivity copies all the data from src to dst, and sends a
// signal to activityCh each time any amount of data is copied.
// Warning: please ensure that both src.Read and dst.Write will both
// return eventually (either with EOF or an error), because the memory
// allocated by CopyWithActivity will not be freed until that happens.
//
// This function does not return until all data is copied. The
// returned error is nil if all data was copied, non-nil otherwise
// (either due to a read error or a write error).
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
