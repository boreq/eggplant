package music

import (
	"strconv"
	"time"

	"github.com/boreq/errors"
	"github.com/oklog/ulid/v2"
)

type StreamId struct {
	value string
}

func NewStreamId() (StreamId, error) {
	id, err := ulid.New(ulid.Now(), ulid.DefaultEntropy())
	if err != nil {
		return StreamId{}, errors.Wrap(err, "could not generate ulid")
	}
	return StreamId{value: id.String()}, nil
}

func NewStreamIdFromString(s string) (StreamId, error) {
	if s == "" {
		return StreamId{}, errors.New("stream id must not be empty")
	}
	if _, err := ulid.Parse(s); err != nil {
		return StreamId{}, errors.Wrap(err, "stream id must be a ulid")
	}
	return StreamId{value: s}, nil
}

func (s StreamId) String() string {
	return s.value
}

type RequestedSeekPosition struct {
	value time.Duration
}

func NewRequestedSeekPosition(d time.Duration) (RequestedSeekPosition, error) {
	if d <= 0 {
		return RequestedSeekPosition{}, errors.New("seek position must be positive")
	}
	return RequestedSeekPosition{value: d}, nil
}

type SeekPosition struct {
	value time.Duration
}

func NewSeekPosition(requested RequestedSeekPosition, track Track) (SeekPosition, error) {
	if requested.value >= track.Duration().Duration() {
		return SeekPosition{}, errors.New("seek position must be before the end of the track")
	}
	return SeekPosition(requested), nil
}

func (s SeekPosition) Duration() time.Duration {
	return s.value
}

type FragmentId struct {
	value int
}

func NewFragmentId(n int) (FragmentId, error) {
	if n < 0 {
		return FragmentId{}, errors.New("fragment id must not be negative")
	}
	return FragmentId{value: n}, nil
}

func MustNewFragmentId(n int) FragmentId {
	v, err := NewFragmentId(n)
	if err != nil {
		panic(err)
	}
	return v
}

func (f FragmentId) Int() int {
	return f.value
}

func (f FragmentId) String() string {
	return strconv.Itoa(f.value)
}
