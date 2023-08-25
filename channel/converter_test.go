package channel

import (
	"fmt"
	"testing"
	"time"
)

func TestConvert(t *testing.T) {
	aC := make(chan int)
	bC, errC := Convert[int, string](aC, func(a int) (b string, err error) {
		return fmt.Sprintf("%+v", a), nil
	})
	wait := time.Second
	timeout := time.NewTimer(wait)
	defer timeout.Stop()
	for _, val := range []int{-10, -1, 0, 1, 10} {
		aC <- val
		select {
		case err := <-errC:
			t.Errorf("unexpected error %v", err)
		case bval, ok := <-bC:
			if !ok {
				t.Errorf("unexpected close closed")
			}
			if bval != fmt.Sprintf("%+v", val) {
				t.Errorf("unexpected value %v", bval)
			}
		case <-timeout.C:
			t.Errorf("timeout")
		}
		timeout.Reset(wait)
	}
	close(aC)
	_, ok := <-bC
	if ok {
		t.Errorf("b channel must be closed")
	}
	_, ok = <-errC
	if ok {
		t.Errorf("error channel must be closed")
	}
}
