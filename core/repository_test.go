package core

import (
	"testing"
	"time"
)

func TestIterateDay(t *testing.T) {
	r := iterateDay(2019, time.February, 28)
	if len(r) != 24 {
		t.Fatalf("error: %v", r)
	}
	for i := 0; i < 24; i++ {
		if !r[i].Equal(time.Date(2019, time.February, 28, i, 0, 0, 0, time.UTC)) {
			t.Fatalf("error [%d]: %v", i, r[i])
		}
	}

func TestIterateMonth(t *testing.T) {
	r := iterateDay(2019, time.February)
	if != 24 {
		t.Fatalf("error: %v", r)
	}
	for i := 0; i < 24; i++ {
		if !r[i].Equal(time.Date(2019, time.February, 28, i, 0, 0, 0, time.UTC)) {
			t.Fatalf("error [%d]: %v", i, r[i])
		}
	}
}
}
