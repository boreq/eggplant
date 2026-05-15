package domain

import (
	"encoding/hex"
	"errors"
	"time"
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
	value string
}

func NewTrackIdFromString(s string) (TrackId, error) {
	if s == "" {
		return TrackId{}, errors.New("track id must not be empty")
	}
	if _, err := hex.DecodeString(s); err != nil {
		return TrackId{}, errors.New("track id must be a hex string")
	}
	return TrackId{value: s}, nil
}

func NewTrackId(parents []AlbumId, title TrackTitle) (TrackId, error) {
	return NewTrackIdFromString(hash(parentsAsString(parents) + title.value))
}

func (id TrackId) String() string {
	return id.value
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
