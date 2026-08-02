package tracker

import (
	"testing"
	"time"
)

func TestSnapshotPreStart(t *testing.T) {
	tr := NewTracker(TrackerOpts{RingSize: 10}, nil)

	done := make(chan []float64, 1)
	go func() { done <- tr.Snapshot("BTCUSDT") }()

	select {
	case got := <-done:
		t.Fatalf("Snapshot returned %v before Start; want block", got)
	case <-time.After(50 * time.Millisecond):
	}

	tr.Stop()

	select {
	case got := <-done:
		if got != nil {
			t.Fatalf("Snapshot = %v after Stop; want nil", got)
		}
	case <-time.After(time.Second):
		t.Fatal("Snapshot still blocked after Stop")
	}
}
