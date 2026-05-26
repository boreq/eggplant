package hls_test

import (
	"strings"
	"testing"
	"time"

	"github.com/boreq/eggplant/domain/hls"
	"github.com/stretchr/testify/require"
)

func TestParseRoundTrip(t *testing.T) {
	input := strings.Join([]string{
		"#EXTM3U",
		"#EXT-X-VERSION:7",
		"#EXT-X-TARGETDURATION:4",
		"#EXT-X-MEDIA-SEQUENCE:0",
		"#EXT-X-PLAYLIST-TYPE:EVENT",
		`#EXT-X-MAP:URI="init.mp4"`,
		"#EXTINF:4.000000,",
		"fragment/seg_000.m4s",
		"#EXTINF:4.000000,",
		"fragment/seg_001.m4s",
		"#EXTINF:2.909375,",
		"fragment/seg_002.m4s",
		"#EXT-X-ENDLIST",
		"",
	}, "\n")

	p, err := hls.Parse(strings.NewReader(input))
	require.NoError(t, err)

	require.Equal(t, 7, p.Version().Int())
	require.Equal(t, 4*time.Second, p.TargetDuration().Duration())
	require.Equal(t, 0, p.MediaSequence().Int())
	require.Equal(t, hls.PlaylistTypeEvent, p.PlaylistType())
	require.Equal(t, "init.mp4", p.MapURI().String())
	require.True(t, p.Complete())

	require.Len(t, p.Segments(), 3)
	require.Equal(t, "fragment/seg_000.m4s", p.Segments()[0].URI().String())
	require.Equal(t, 4*time.Second, p.Segments()[0].Duration().Duration())
	require.Equal(t, time.Duration(2909375000), p.Segments()[2].Duration().Duration())

	require.Equal(t, input, string(p.Bytes()))
}

func TestParseErrors(t *testing.T) {
	cases := []struct {
		name  string
		input []string
	}{
		{
			name: "missing_extm3u",
			input: []string{
				"#EXT-X-VERSION:7",
				"#EXTINF:1,",
				"seg.m4s",
				"",
			},
		},
		{
			name: "uri_before_extinf",
			input: []string{
				"#EXTM3U",
				"seg.m4s",
				"",
			},
		},
		{
			name: "dangling_extinf",
			input: []string{
				"#EXTM3U",
				"#EXT-X-VERSION:7",
				"#EXT-X-TARGETDURATION:4",
				"#EXT-X-MEDIA-SEQUENCE:0",
				"#EXT-X-PLAYLIST-TYPE:VOD",
				`#EXT-X-MAP:URI="init.mp4"`,
				"#EXTINF:1.0,",
				"",
			},
		},
		{
			name: "unknown_tag",
			input: []string{
				"#EXTM3U",
				"#EXT-X-VERSION:7",
				"#EXT-X-WHATEVER:1",
				"",
			},
		},
		{
			name: "no_version",
			input: []string{
				"#EXTM3U",
				"#EXT-X-TARGETDURATION:1",
				"#EXT-X-MEDIA-SEQUENCE:0",
				"#EXT-X-PLAYLIST-TYPE:VOD",
				`#EXT-X-MAP:URI="init.mp4"`,
				"#EXT-X-ENDLIST",
				"",
			},
		},
		{
			name: "no_target_duration",
			input: []string{
				"#EXTM3U",
				"#EXT-X-VERSION:7",
				"#EXT-X-MEDIA-SEQUENCE:0",
				"#EXT-X-PLAYLIST-TYPE:VOD",
				`#EXT-X-MAP:URI="init.mp4"`,
				"#EXT-X-ENDLIST",
				"",
			},
		},
		{
			name: "no_media_sequence",
			input: []string{
				"#EXTM3U",
				"#EXT-X-VERSION:7",
				"#EXT-X-TARGETDURATION:1",
				"#EXT-X-PLAYLIST-TYPE:VOD",
				`#EXT-X-MAP:URI="init.mp4"`,
				"#EXT-X-ENDLIST",
				"",
			},
		},
		{
			name: "no_playlist_type",
			input: []string{
				"#EXTM3U",
				"#EXT-X-VERSION:7",
				"#EXT-X-TARGETDURATION:1",
				"#EXT-X-MEDIA-SEQUENCE:0",
				`#EXT-X-MAP:URI="init.mp4"`,
				"#EXT-X-ENDLIST",
				"",
			},
		},
		{
			name: "no_map",
			input: []string{
				"#EXTM3U",
				"#EXT-X-VERSION:7",
				"#EXT-X-TARGETDURATION:1",
				"#EXT-X-MEDIA-SEQUENCE:0",
				"#EXT-X-PLAYLIST-TYPE:VOD",
				"#EXT-X-ENDLIST",
				"",
			},
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, err := hls.Parse(strings.NewReader(strings.Join(c.input, "\n")))
			require.Error(t, err)
		})
	}
}

func TestParseSkipsCommentsAndBlankLines(t *testing.T) {
	input := strings.Join([]string{
		"#EXTM3U",
		"# this is a comment",
		"",
		"#EXT-X-VERSION:7",
		"#EXT-X-TARGETDURATION:1",
		"#EXT-X-MEDIA-SEQUENCE:0",
		"#EXT-X-PLAYLIST-TYPE:VOD",
		`#EXT-X-MAP:URI="init.mp4"`,
		"#EXTINF:1.0,",
		"seg.m4s",
		"#EXT-X-ENDLIST",
		"",
	}, "\n")

	p, err := hls.Parse(strings.NewReader(input))
	require.NoError(t, err)
	require.Len(t, p.Segments(), 1)
	require.Equal(t, "seg.m4s", p.Segments()[0].URI().String())
	require.Equal(t, hls.PlaylistTypeVOD, p.PlaylistType())
}
