package remote

import (
	"net/http"
	"net/http/httputil"
	"net/url"

	appremote "github.com/boreq/eggplant/application/remote"
	remotedomain "github.com/boreq/eggplant/domain/remote"
	"github.com/boreq/eggplant/internal/logging"
	"github.com/boreq/errors"
)

type StreamProxy struct {
	transactionProvider appremote.TransactionProvider
	log                 logging.Logger
}

func NewStreamProxy(transactionProvider appremote.TransactionProvider) *StreamProxy {
	return &StreamProxy{
		transactionProvider: transactionProvider,
		log:                 logging.New("remote.StreamProxy"),
	}
}

// Proxy forwards the request to <instance>/api/track<suffix>. The suffix is the
// catch-all part of the local route and already starts with a slash, e.g.
// "/<trackId>/stream/<streamId>/playlist".
func (p *StreamProxy) Proxy(w http.ResponseWriter, r *http.Request, instanceId remotedomain.RemoteInstanceID, suffix string) error {
	target, err := p.targetByID(instanceId)
	if err != nil {
		return err
	}

	base, err := url.Parse(target.address.String())
	if err != nil {
		return errors.Wrap(err, "could not parse the remote address")
	}

	proxy := &httputil.ReverseProxy{
		Director: func(req *http.Request) {
			req.URL.Scheme = base.Scheme
			req.URL.Host = base.Host
			req.Host = base.Host
			req.URL.Path = "/api/track" + suffix
			SetAuthToken(req, target.token)
		},
		ErrorHandler: func(w http.ResponseWriter, r *http.Request, err error) {
			p.log.Error("remote stream proxy failed", "instance", instanceId.String(), "err", err)
			w.WriteHeader(http.StatusBadGateway)
		},
	}
	proxy.ServeHTTP(w, r)
	return nil
}

func (p *StreamProxy) targetByID(id remotedomain.RemoteInstanceID) (remoteTarget, error) {
	var target remoteTarget
	found := false
	if err := p.transactionProvider.Read(func(r *appremote.TransactableRepositories) error {
		instance, err := r.RemoteInstances.GetByID(id)
		if err != nil {
			if errors.Is(err, appremote.ErrNotFound) {
				return nil
			}
			return errors.Wrap(err, "could not get the remote instance")
		}
		token, ok := instance.RemoteAuthToken()
		if !ok {
			return nil
		}
		target = remoteTarget{
			id:      instance.Id(),
			address: instance.Address(),
			token:   token,
		}
		found = true
		return nil
	}); err != nil {
		return remoteTarget{}, errors.Wrap(err, "transaction failed")
	}
	if !found {
		return remoteTarget{}, appremote.ErrNotFound
	}
	return target, nil
}
