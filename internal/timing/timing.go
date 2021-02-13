package timing

import (
	"fmt"
	"strings"
	"time"
)

type Timer struct {
	name     string
	started  time.Time
	stopped  time.Time
	children []*Timer
}

func NewTimer(name string) *Timer {
	return &Timer{
		name:    name,
		started: time.Now(),
	}
}

func (t *Timer) Stop() {
	t.stopped = time.Now()
}

func (t *Timer) Duration() (time.Duration, bool) {
	if t.stopped.IsZero() {
		return 0, false
	}
	return t.stopped.Sub(t.started), true
}

func (t *Timer) New(name string) *Timer {
	n := NewTimer(name)
	t.children = append(t.children, n)
	return n
}

func (t *Timer) StopAndPrint() {
	t.Stop()
	t.print(0, 0)
}

func (t *Timer) print(i int, total time.Duration) {
	indent := strings.Repeat(" ", i)

	d, ok := t.Duration()
	if ok {
		var percentage float64
		if total == 0 {
			percentage = 100
		} else {
			percentage = d.Seconds() / total.Seconds() * 100
		}

		fmt.Printf("%s%s %6s (%.0f%%)\n", indent, t.name, d, percentage)
	} else {
		fmt.Printf("%s%s still running\n", indent, t.name)
	}

	for _, child := range t.children {
		child.print(i+1, d)
	}
}
