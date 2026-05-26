package hls

import (
	"fmt"
	"strings"
)

func (p Playlist) Bytes() []byte {
	var sb strings.Builder
	sb.WriteString("#EXTM3U\n")
	fmt.Fprintf(&sb, "#EXT-X-VERSION:%d\n", p.version.Int())
	fmt.Fprintf(&sb, "#EXT-X-TARGETDURATION:%d\n", int(p.targetDuration.Duration().Seconds()))
	fmt.Fprintf(&sb, "#EXT-X-MEDIA-SEQUENCE:%d\n", p.mediaSequence.Int())
	fmt.Fprintf(&sb, "#EXT-X-PLAYLIST-TYPE:%s\n", p.playlistType.String())
	fmt.Fprintf(&sb, "#EXT-X-MAP:URI=%q\n", p.mapURI.String())
	for _, seg := range p.segments {
		fmt.Fprintf(&sb, "#EXTINF:%f,\n", seg.duration.Duration().Seconds())
		sb.WriteString(seg.uri.String())
		sb.WriteByte('\n')
	}
	if p.complete {
		sb.WriteString("#EXT-X-ENDLIST\n")
	}
	return []byte(sb.String())
}
