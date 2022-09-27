package sleepingd

import "io"

func CopyWithActivity(dst io.Writer, src io.Reader, activityCh chan<- struct{}) error {
	buf := make([]byte, 32*1024)
	// Implementation baesd on copyBuffer in io from stdlib
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
