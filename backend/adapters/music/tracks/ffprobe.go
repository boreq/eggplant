package tracks

import (
	"bytes"
	"context"
	"os/exec"
	"strings"
	"time"

	"github.com/boreq/eggplant/domain/music"
	"github.com/boreq/errors"
)

type FFProbe struct{}

func NewFFProbe() (*FFProbe, error) {
	if _, err := exec.LookPath("ffprobe"); err != nil {
		return nil, errors.Wrap(err, "ffprobe not found in PATH")
	}
	return &FFProbe{}, nil
}

func (p *FFProbe) GetDuration(ctx context.Context, path string) (music.TrackDuration, error) {
	args := []string{
		"-v", "error",
		"-show_entries", "format=duration",
		"-of", "default=noprint_wrappers=1:nokey=1",
		path,
	}
	cmd := exec.CommandContext(ctx, "ffprobe", args...)
	bufErr := &bytes.Buffer{}
	cmd.Stderr = bufErr
	output, err := cmd.Output()
	if err != nil {
		return music.TrackDuration{}, errors.Wrapf(err, "ffprobe execution failed: %s", strings.TrimSpace(bufErr.String()))
	}

	d, err := time.ParseDuration(strings.TrimSpace(string(output)) + "s")
	if err != nil {
		return music.TrackDuration{}, errors.Wrap(err, "could not parse ffprobe duration")
	}

	td, err := music.NewTrackDuration(d)
	if err != nil {
		return music.TrackDuration{}, errors.Wrap(err, "invalid track duration")
	}
	return td, nil
}
