package domain

import (
	"time"

	"github.com/boreq/errors"
)

type Track struct {
	id       TrackId
	fileId   FileId
	title    TrackTitle
	duration TrackDuration
}

func NewTrack(id TrackId, fileId FileId, title TrackTitle, duration TrackDuration) Track {
	return Track{
		id:       id,
		fileId:   fileId,
		title:    title,
		duration: duration,
	}
}

func (t Track) Id() TrackId {
	return t.id
}

func (t Track) FileId() FileId {
	return t.fileId
}

func (t Track) Title() TrackTitle {
	return t.title
}

func (t Track) Duration() TrackDuration {
	return t.duration
}

type TrackId struct {
	id idForHumans
}

func NewTrackId(parents []AlbumId, title TrackTitle) (TrackId, error) {
	return TrackId{id: newIdForHumans(parents, title)}, nil
}

func NewTrackIdFromString(s string) (TrackId, error) {
	id, err := newIdForHumansFromString(s)
	if err != nil {
		return TrackId{}, errors.Wrap(err, "invalid track id")
	}
	return TrackId{id: id}, nil
}

func (t TrackId) String() string {
	return t.id.String()
}

type TrackTitle struct {
	value string
}

func NewTrackTitle(s string) (TrackTitle, error) {
	if s == "" {
		return TrackTitle{}, errors.New("track title must not be empty")
	}
	return TrackTitle{value: s}, nil
}

func (t TrackTitle) String() string {
	return t.value
}

type TrackDuration struct {
	value time.Duration
}

func NewTrackDuration(d time.Duration) (TrackDuration, error) {
	if d <= 0 {
		return TrackDuration{}, errors.New("track duration must be positive")
	}
	return TrackDuration{value: d}, nil
}

func (d TrackDuration) Seconds() float64 {
	return d.value.Seconds()
}

type TrackFragmentId struct {
	value int
}

func NewTrackFragmentId(n int) (TrackFragmentId, error) {
	if n < 0 {
		return TrackFragmentId{}, errors.New("track fragment id must not be negative")
	}
	return TrackFragmentId{value: n}, nil
}

func MustNewTrackFragmentId(n int) TrackFragmentId {
	v, err := NewTrackFragmentId(n)
	if err != nil {
		panic(err)
	}
	return v
}

func (t TrackFragmentId) Int() int {
	return t.value
}
