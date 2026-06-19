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

func NewHealthcheckStatusFromString(s string) (HealthcheckStatus, error) {
	switch s {
	case HealthcheckStatusAlive.s:
		return HealthcheckStatusAlive, nil
	case HealthcheckStatusDead.s:
		return HealthcheckStatusDead, nil
	default:
		return HealthcheckStatus{}, errors.New("unknown healthcheck status")
	}
}

func (s HealthcheckStatus) IsZero() bool {
	return s == HealthcheckStatus{}
}

func (s HealthcheckStatus) String() string {
	return s.s
}

type Healthcheck struct {
	status HealthcheckStatus
	at     time.Time
}

func NewHealthcheck(status HealthcheckStatus, at time.Time) (Healthcheck, error) {
	if status.IsZero() {
		return Healthcheck{}, errors.New("status must be set")
	}
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
