package hls_test

import (
	"testing"
	"time"

	"github.com/boreq/eggplant/domain/music/hls"
	"github.com/stretchr/testify/require"
)

func TestNewVersionRejectsNonPositive(t *testing.T) {
	for _, n := range []int{-1, 0} {
		_, err := hls.NewVersion(n)
		require.Error(t, err)
	}
	v, err := hls.NewVersion(1)
	require.NoError(t, err)
	require.Equal(t, 1, v.Int())
}

func TestNewMediaSequenceRejectsNegative(t *testing.T) {
	_, err := hls.NewMediaSequence(-1)
	require.Error(t, err)
	ms, err := hls.NewMediaSequence(0)
	require.NoError(t, err)
	require.Equal(t, 0, ms.Int())
}

func TestNewTargetDurationRejectsNonSecond(t *testing.T) {
	_, err := hls.NewTargetDuration(500 * time.Millisecond)
	require.Error(t, err)
	_, err = hls.NewTargetDuration(0)
	require.Error(t, err)
	td, err := hls.NewTargetDuration(4 * time.Second)
	require.NoError(t, err)
	require.Equal(t, 4*time.Second, td.Duration())
}

func TestNewSegmentDurationRejectsNonPositive(t *testing.T) {
	_, err := hls.NewSegmentDuration(0)
	require.Error(t, err)
	_, err = hls.NewSegmentDuration(-1)
	require.Error(t, err)
	sd, err := hls.NewSegmentDuration(2909375 * time.Microsecond)
	require.NoError(t, err)
	require.Equal(t, time.Duration(2909375000), sd.Duration())
}

func TestNewMapURIRejectsEmpty(t *testing.T) {
	_, err := hls.NewMapURI("")
	require.Error(t, err)
	m, err := hls.NewMapURI("init.mp4")
	require.NoError(t, err)
	require.Equal(t, "init.mp4", m.String())
}

func TestNewSegmentURIRejectsEmpty(t *testing.T) {
	_, err := hls.NewSegmentURI("")
	require.Error(t, err)
	u, err := hls.NewSegmentURI("seg.m4s")
	require.NoError(t, err)
	require.Equal(t, "seg.m4s", u.String())
}

func TestPlaylistTypeConstants(t *testing.T) {
	require.Equal(t, "EVENT", hls.PlaylistTypeEvent.String())
	require.Equal(t, "VOD", hls.PlaylistTypeVOD.String())
}

func TestNewPlaylistRejectsSegmentExceedingTargetDuration(t *testing.T) {
	version, err := hls.NewVersion(7)
	require.NoError(t, err)
	target, err := hls.NewTargetDuration(4 * time.Second)
	require.NoError(t, err)
	ms, err := hls.NewMediaSequence(0)
	require.NoError(t, err)
	mapURI, err := hls.NewMapURI("init.mp4")
	require.NoError(t, err)
	uri, err := hls.NewSegmentURI("seg.m4s")
	require.NoError(t, err)

	mk := func(d time.Duration) []hls.Segment {
		sd, err := hls.NewSegmentDuration(d)
		require.NoError(t, err)
		return []hls.Segment{hls.NewSegment(sd, uri)}
	}

	_, err = hls.NewPlaylist(version, target, ms, hls.PlaylistTypeVOD, mapURI, mk(4400*time.Millisecond), true)
	require.NoError(t, err, "4.4s rounds to 4 — within target")

	_, err = hls.NewPlaylist(version, target, ms, hls.PlaylistTypeVOD, mapURI, mk(4500*time.Millisecond), true)
	require.Error(t, err, "4.5s rounds to 5 — exceeds target")

	_, err = hls.NewPlaylist(version, target, ms, hls.PlaylistTypeVOD, mapURI, mk(10*time.Second), true)
	require.Error(t, err)
}
