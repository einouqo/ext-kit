package util

import "testing"

func TestOnce_Do(t *testing.T) {
	var (
		i    int
		once = NewOnce(func() { i++ })
	)
	once.Exec()
	if i != 1 {
		t.Errorf("i = %d; want 1", i)
	}
	once.Exec()
	if i != 1 {
		t.Errorf("i = %d; want 1", i)
	}
}
