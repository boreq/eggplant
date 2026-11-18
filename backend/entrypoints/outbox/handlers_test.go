package outbox_test

import (
	"context"
	"testing"

	remoteadapter "github.com/boreq/eggplant/adapters/remote"
	remoteapp "github.com/boreq/eggplant/application/remote"
	remotedomain "github.com/boreq/eggplant/domain/remote"
	"github.com/boreq/eggplant/entrypoints/outbox"
	"github.com/boreq/errors"
	"github.com/stretchr/testify/require"
)

func TestRemotePairedHandler(t *testing.T) {
	testCases := []struct {
		Name          string
		Err           error
		ExpectedError bool
	}{
		{
			Name:          "success",
			Err:           nil,
			ExpectedError: false,
		},
		{
			Name:          "removed_instance",
			Err:           errors.Wrap(remoteapp.ErrNotFound, "could not get the remote instance"),
			ExpectedError: false,
		},
		{
			Name:          "other_error",
			Err:           errors.New("some error"),
			ExpectedError: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			id, err := remotedomain.NewRemoteInstanceID()
			require.NoError(t, err)

			msg, err := remoteadapter.MarshalRemotePaired(remotedomain.NewRemotePaired(id))
			require.NoError(t, err)

			checkRemote := &checkRemoteHandlerMock{err: testCase.Err}
			handler := outbox.NewRemotePairedHandler(checkRemote)

			err = handler.Handle(msg)
			if testCase.ExpectedError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			require.Equal(t, []remoteapp.CheckRemote{remoteapp.NewCheckRemote(id)}, checkRemote.calls)
		})
	}
}

func TestRemotePairingTokenSetHandler(t *testing.T) {
	testCases := []struct {
		Name          string
		Err           error
		ExpectedError bool
	}{
		{
			Name:          "success",
			Err:           nil,
			ExpectedError: false,
		},
		{
			Name:          "removed_instance",
			Err:           errors.Wrap(remoteapp.ErrNotFound, "could not get the remote instance"),
			ExpectedError: false,
		},
		{
			Name:          "other_error",
			Err:           errors.New("some error"),
			ExpectedError: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			id, err := remotedomain.NewRemoteInstanceID()
			require.NoError(t, err)

			msg, err := remoteadapter.MarshalRemotePairingTokenSet(remotedomain.NewRemotePairingTokenSet(id))
			require.NoError(t, err)

			sendLocalAuthToken := &sendLocalAuthTokenHandlerMock{err: testCase.Err}
			handler := outbox.NewRemotePairingTokenSetHandler(sendLocalAuthToken)

			err = handler.Handle(msg)
			if testCase.ExpectedError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			require.Equal(t, []remoteapp.SendLocalAuthToken{remoteapp.NewSendLocalAuthToken(id)}, sendLocalAuthToken.calls)
		})
	}
}

type checkRemoteHandlerMock struct {
	err   error
	calls []remoteapp.CheckRemote
}

func (m *checkRemoteHandlerMock) Execute(ctx context.Context, cmd remoteapp.CheckRemote) error {
	m.calls = append(m.calls, cmd)
	return m.err
}

type sendLocalAuthTokenHandlerMock struct {
	err   error
	calls []remoteapp.SendLocalAuthToken
}

func (m *sendLocalAuthTokenHandlerMock) Execute(ctx context.Context, cmd remoteapp.SendLocalAuthToken) error {
	m.calls = append(m.calls, cmd)
	return m.err
}
