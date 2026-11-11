package remote

type EventSource interface {
	PopEvents() []Event
}

type EventRecorder struct {
	recordedEvents []Event
}

func NewEventRecorder() EventRecorder {
	return EventRecorder{recordedEvents: nil}
}

func (r *EventRecorder) RecordEvent(event Event) {
	r.recordedEvents = append(r.recordedEvents, event)
}

func (r *EventRecorder) PopEvents() []Event {
	events := r.recordedEvents
	r.recordedEvents = nil
	return events
}
