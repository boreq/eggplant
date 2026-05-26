package hls

import (
	"fmt"
	"math"
	"time"

	"github.com/boreq/errors"
)

type Playlist struct {
	version        Version
	targetDuration TargetDuration
	mediaSequence  MediaSequence
	playlistType   PlaylistType
	mapURI         MapURI
	segments       []Segment
	complete       bool
}

func NewPlaylist(
	version Version,
	targetDuration TargetDuration,
	mediaSequence MediaSequence,
	playlistType PlaylistType,
	mapURI MapURI,
	segments []Segment,
	complete bool,
) (Playlist, error) {
	targetSecs := targetDuration.Duration().Seconds()
	for i, seg := range segments {
		rounded := math.Round(seg.Duration().Duration().Seconds())
		if rounded > targetSecs {
			return Playlist{}, fmt.Errorf(
				"segment %d duration %s exceeds target duration %s when rounded",
				i, seg.Duration().Duration(), targetDuration.Duration(),
			)
		}
	}
	return Playlist{
		version:        version,
		targetDuration: targetDuration,
		mediaSequence:  mediaSequence,
		playlistType:   playlistType,
		mapURI:         mapURI,
		segments:       segments,
		complete:       complete,
	}, nil
}

func (p Playlist) Version() Version {
	return p.version
}

func (p Playlist) TargetDuration() TargetDuration {
	return p.targetDuration
}

func (p Playlist) MediaSequence() MediaSequence {
	return p.mediaSequence
}

func (p Playlist) PlaylistType() PlaylistType {
	return p.playlistType
}

func (p Playlist) MapURI() MapURI {
	return p.mapURI
}

func (p Playlist) Segments() []Segment {
	return p.segments
}

func (p Playlist) Complete() bool {
	return p.complete
}

func (p Playlist) String() string {
	first, last := "", ""
	if len(p.segments) > 0 {
		first = p.segments[0].uri.String()
		last = p.segments[len(p.segments)-1].uri.String()
	}
	return fmt.Sprintf(
		"Playlist{mediaSequence=%d, segments=%d, first=%q, last=%q, complete=%t}",
		p.mediaSequence.Int(), len(p.segments), first, last, p.complete,
	)
}

type Segment struct {
	duration SegmentDuration
	uri      SegmentURI
}

func NewSegment(duration SegmentDuration, uri SegmentURI) Segment {
	return Segment{
		duration: duration,
		uri:      uri,
	}
}

func (s Segment) Duration() SegmentDuration {
	return s.duration
}

func (s Segment) URI() SegmentURI {
	return s.uri
}

type Version struct {
	value int
}

func NewVersion(n int) (Version, error) {
	if n < 1 {
		return Version{}, errors.New("version must be >= 1")
	}
	return Version{value: n}, nil
}

func (v Version) Int() int {
	return v.value
}

type TargetDuration struct {
	value time.Duration
}

func NewTargetDuration(d time.Duration) (TargetDuration, error) {
	if d <= 0 {
		return TargetDuration{}, errors.New("target duration must be positive")
	}
	if d%time.Second != 0 {
		return TargetDuration{}, errors.New("target duration must be a whole number of seconds")
	}
	return TargetDuration{value: d}, nil
}

func (td TargetDuration) Duration() time.Duration {
	return td.value
}

type MediaSequence struct {
	value int
}

func NewMediaSequence(n int) (MediaSequence, error) {
	if n < 0 {
		return MediaSequence{}, errors.New("media sequence must not be negative")
	}
	return MediaSequence{value: n}, nil
}

func (ms MediaSequence) Int() int {
	return ms.value
}

type PlaylistType struct {
	value string
}

var (
	PlaylistTypeEvent = PlaylistType{value: "EVENT"}
	PlaylistTypeVOD   = PlaylistType{value: "VOD"}
)

func (pt PlaylistType) String() string {
	return pt.value
}

type MapURI struct {
	value string
}

func NewMapURI(s string) (MapURI, error) {
	if s == "" {
		return MapURI{}, errors.New("map uri must not be empty")
	}
	return MapURI{value: s}, nil
}

func (m MapURI) String() string {
	return m.value
}

type SegmentURI struct {
	value string
}

func NewSegmentURI(s string) (SegmentURI, error) {
	if s == "" {
		return SegmentURI{}, errors.New("segment uri must not be empty")
	}
	return SegmentURI{value: s}, nil
}

func (s SegmentURI) String() string {
	return s.value
}

type SegmentDuration struct {
	value time.Duration
}

func NewSegmentDuration(d time.Duration) (SegmentDuration, error) {
	if d <= 0 {
		return SegmentDuration{}, errors.New("segment duration must be positive")
	}
	return SegmentDuration{value: d}, nil
}

func (sd SegmentDuration) Duration() time.Duration {
	return sd.value
}
