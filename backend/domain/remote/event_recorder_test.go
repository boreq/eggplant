package remote_test

import (
	"testing"

	remotedomain "github.com/boreq/eggplant/domain/remote"
	"github.com/stretchr/testify/require"
)

func TestEventRecorder_NewRecorderHasNoEvents(t *testing.T) {
	recorder := remotedomain.NewEventRecorder()

	require.Empty(t, recorder.PopEvents())
}

func TestEventRecorder_PopEventsReturnsEventsInRecordedOrder(t *testing.T) {
	recorder := remotedomain.NewEventRecorder()

	recorder.RecordEvent(testEvent{name: "first"})
	recorder.RecordEvent(testEvent{name: "second"})

	require.Equal(t, []remotedomain.Event{
		testEvent{name: "first"},
		testEvent{name: "second"},
	}, recorder.PopEvents())
}

func TestEventRecorder_PopEventsClearsEvents(t *testing.T) {
	recorder := remotedomain.NewEventRecorder()
	recorder.RecordEvent(testEvent{name: "event"})

	require.Len(t, recorder.PopEvents(), 1)
	require.Empty(t, recorder.PopEvents())
}

func TestEventRecorder_RecordingAfterPopStartsAgain(t *testing.T) {
	recorder := remotedomain.NewEventRecorder()

	recorder.RecordEvent(testEvent{name: "first"})
	recorder.PopEvents()
	recorder.RecordEvent(testEvent{name: "second"})

	require.Equal(t, []remotedomain.Event{testEvent{name: "second"}}, recorder.PopEvents())
}

type testEvent struct {
	name string
}

func (e testEvent) EventName() string {
	return e.name
}
