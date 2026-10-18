package remote

import (
	"time"

	"github.com/boreq/errors"
)

type HealthcheckStatus struct {
	s string
}

var (
	HealthcheckStatusAlive = HealthcheckStatus{"ALIVE"}
	HealthcheckStatusDead  = HealthcheckStatus{"DEAD"}
)

func (s HealthcheckStatus) String() string {
	return s.s
}

type Healthcheck struct {
	status HealthcheckStatus
	at     time.Time
}

func NewHealthcheck(status HealthcheckStatus, at time.Time) (Healthcheck, error) {
	if at.IsZero() {
		return Healthcheck{}, errors.New("time must be set")
	}
	return Healthcheck{status: status, at: at}, nil
}

func (h Healthcheck) Status() HealthcheckStatus {
	return h.status
}

func (h Healthcheck) At() time.Time {
	return h.at
}
