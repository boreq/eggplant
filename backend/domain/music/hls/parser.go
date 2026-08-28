package hls

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/boreq/errors"
)

func Parse(r io.Reader) (Playlist, error) {
	sc := bufio.NewScanner(r)
	sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	var (
		seenExtm3u        bool
		sawVersion        bool
		version           Version
		sawTargetDuration bool
		targetDuration    TargetDuration
		sawMediaSequence  bool
		mediaSequence     MediaSequence
		sawPlaylistType   bool
		playlistType      PlaylistType
		sawMap            bool
		mapURI            MapURI
		segments          []Segment
		complete          bool
		pendingInf        *extinf
	)

	for sc.Scan() {
		line := strings.TrimRight(sc.Text(), "\r")
		if line == "" {
			continue
		}
		if !strings.HasPrefix(line, "#") {
			if pendingInf == nil {
				return Playlist{}, errors.New("encountered uri without preceding EXTINF")
			}
			uri, err := NewSegmentURI(line)
			if err != nil {
				return Playlist{}, errors.Wrap(err, "invalid segment uri")
			}
			segments = append(segments, NewSegment(pendingInf.duration, uri))
			pendingInf = nil
			continue
		}
		if !strings.HasPrefix(line, "#EXT") {
			continue
		}
		name, value := splitTag(line)
		if !seenExtm3u && name != "EXTM3U" {
			return Playlist{}, errors.New("playlist must start with #EXTM3U")
		}
		switch name {
		case "EXTM3U":
			seenExtm3u = true
		case "EXT-X-VERSION":
			n, err := strconv.Atoi(value)
			if err != nil {
				return Playlist{}, errors.Wrap(err, "invalid version")
			}
			v, err := NewVersion(n)
			if err != nil {
				return Playlist{}, err
			}
			version = v
			sawVersion = true
		case "EXT-X-TARGETDURATION":
			n, err := strconv.Atoi(value)
			if err != nil {
				return Playlist{}, errors.Wrap(err, "invalid target duration")
			}
			td, err := NewTargetDuration(time.Duration(n) * time.Second)
			if err != nil {
				return Playlist{}, err
			}
			targetDuration = td
			sawTargetDuration = true
		case "EXT-X-MEDIA-SEQUENCE":
			n, err := strconv.Atoi(value)
			if err != nil {
				return Playlist{}, errors.Wrap(err, "invalid media sequence")
			}
			ms, err := NewMediaSequence(n)
			if err != nil {
				return Playlist{}, err
			}
			mediaSequence = ms
			sawMediaSequence = true
		case "EXT-X-PLAYLIST-TYPE":
			switch value {
			case PlaylistTypeEvent.String():
				playlistType = PlaylistTypeEvent
			case PlaylistTypeVOD.String():
				playlistType = PlaylistTypeVOD
			default:
				return Playlist{}, fmt.Errorf("unknown playlist type: %q", value)
			}
			sawPlaylistType = true
		case "EXT-X-MAP":
			raw, err := parseMapURI(value)
			if err != nil {
				return Playlist{}, errors.Wrap(err, "invalid map")
			}
			m, err := NewMapURI(raw)
			if err != nil {
				return Playlist{}, err
			}
			mapURI = m
			sawMap = true
		case "EXTINF":
			inf, err := parseExtinf(value)
			if err != nil {
				return Playlist{}, errors.Wrap(err, "invalid extinf")
			}
			pendingInf = &inf
		case "EXT-X-ENDLIST":
			complete = true
		default:
			return Playlist{}, fmt.Errorf("unsupported tag: %s", name)
		}
	}
	if err := sc.Err(); err != nil {
		return Playlist{}, errors.Wrap(err, "scanner error")
	}
	if pendingInf != nil {
		return Playlist{}, errors.New("dangling EXTINF without uri")
	}
	if !sawVersion {
		return Playlist{}, errors.New("missing EXT-X-VERSION")
	}
	if !sawTargetDuration {
		return Playlist{}, errors.New("missing EXT-X-TARGETDURATION")
	}
	if !sawMediaSequence {
		return Playlist{}, errors.New("missing EXT-X-MEDIA-SEQUENCE")
	}
	if !sawPlaylistType {
		return Playlist{}, errors.New("missing EXT-X-PLAYLIST-TYPE")
	}
	if !sawMap {
		return Playlist{}, errors.New("missing EXT-X-MAP")
	}
	return NewPlaylist(version, targetDuration, mediaSequence, playlistType, mapURI, segments, complete)
}

type extinf struct {
	duration SegmentDuration
}

func parseExtinf(value string) (extinf, error) {
	before, _, ok := strings.Cut(value, ",")
	if !ok {
		return extinf{}, errors.New("missing comma in EXTINF value")
	}
	secs, err := strconv.ParseFloat(before, 64)
	if err != nil {
		return extinf{}, errors.Wrap(err, "invalid duration")
	}
	d, err := NewSegmentDuration(time.Duration(secs * float64(time.Second)))
	if err != nil {
		return extinf{}, err
	}
	return extinf{
		duration: d,
	}, nil
}

func parseMapURI(value string) (string, error) {
	attrs, err := parseAttrList(value)
	if err != nil {
		return "", err
	}
	uri, ok := attrs["URI"]
	if !ok {
		return "", errors.New("EXT-X-MAP missing URI attribute")
	}
	return uri, nil
}

func parseAttrList(s string) (map[string]string, error) {
	m := make(map[string]string)
	i := 0
	for i < len(s) {
		eq := strings.IndexByte(s[i:], '=')
		if eq < 0 {
			return nil, errors.New("attribute missing =")
		}
		key := s[i : i+eq]
		i += eq + 1
		if i < len(s) && s[i] == '"' {
			end := strings.IndexByte(s[i+1:], '"')
			if end < 0 {
				return nil, errors.New("unterminated quoted attribute value")
			}
			m[key] = s[i+1 : i+1+end]
			i += end + 2
		} else {
			comma := strings.IndexByte(s[i:], ',')
			if comma < 0 {
				m[key] = s[i:]
				i = len(s)
			} else {
				m[key] = s[i : i+comma]
				i += comma
			}
		}
		if i < len(s) && s[i] == ',' {
			i++
		}
	}
	return m, nil
}

func splitTag(line string) (name, value string) {
	body := line[1:]
	before, after, ok := strings.Cut(body, ":")
	if !ok {
		return body, ""
	}
	return before, after
}
