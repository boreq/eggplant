package tracks

import (
	"bytes"
	"os/exec"
	"strings"
	"time"

	"github.com/boreq/eggplant/domain"
	"github.com/boreq/errors"
)

type FFProbe struct{}

func NewFFProbe() *FFProbe {
	return &FFProbe{}
}

func (p *FFProbe) GetDuration(path string) (domain.TrackDuration, error) {
	args := []string{
		"-v", "error",
		"-show_entries", "format=duration",
		"-of", "default=noprint_wrappers=1:nokey=1",
		path,
	}
	cmd := exec.Command("ffprobe", args...)
	bufErr := &bytes.Buffer{}
	cmd.Stderr = bufErr
	output, err := cmd.Output()
	if err != nil {
		return domain.TrackDuration{}, errors.Wrapf(err, "ffprobe execution failed: %s", strings.TrimSpace(bufErr.String()))
	}

	d, err := time.ParseDuration(strings.TrimSpace(string(output)) + "s")
	if err != nil {
		return domain.TrackDuration{}, errors.Wrap(err, "could not parse ffprobe duration")
	}

	td, err := domain.NewTrackDuration(d)
	if err != nil {
		return domain.TrackDuration{}, errors.Wrap(err, "invalid track duration")
	}
	return td, nil
}
