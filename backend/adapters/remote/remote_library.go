package remote

import (
	"context"
	"io"
	"time"

	"github.com/boreq/eggplant/adapters/openapi"
	appremote "github.com/boreq/eggplant/application/remote"
	musicdomain "github.com/boreq/eggplant/domain/music"
	"github.com/boreq/eggplant/domain/music/library"
	remotedomain "github.com/boreq/eggplant/domain/remote"
	"github.com/boreq/eggplant/internal/logging"
	"github.com/boreq/errors"
)

type RemoteLibrary struct {
	transactionProvider appremote.TransactionProvider
	client              *RemoteClient
	log                 logging.Logger
}

func NewRemoteLibrary(transactionProvider appremote.TransactionProvider, client *RemoteClient) *RemoteLibrary {
	return &RemoteLibrary{
		transactionProvider: transactionProvider,
		client:              client,
		log:                 logging.New("remote.RemoteLibrary"),
	}
}

type remoteTarget struct {
	id      remotedomain.RemoteInstanceID
	address remotedomain.RemoteInstanceAddress
	token   remotedomain.AuthToken
}

func (l *RemoteLibrary) GetRootAlbum(ctx context.Context, instanceId remotedomain.RemoteInstanceID) (musicdomain.RootAlbum, error) {
	target, err := l.targetByID(instanceId)
	if err != nil {
		return musicdomain.RootAlbum{}, err
	}

	resp, err := l.client.GetRootAlbum(ctx, target.address, target.token)
	if err != nil {
		return musicdomain.RootAlbum{}, errors.Wrap(err, "could not get the remote root album")
	}

	albums, err := toRemotePartialAlbums(resp.Albums, instanceId)
	if err != nil {
		return musicdomain.RootAlbum{}, errors.Wrap(err, "could not convert the child albums")
	}

	tracks := make([]musicdomain.Track, 0, len(resp.Tracks))
	for _, t := range resp.Tracks {
		track, err := toRemoteTrack(t, instanceId)
		if err != nil {
			return musicdomain.RootAlbum{}, errors.Wrap(err, "could not convert a track")
		}
		tracks = append(tracks, track)
	}

	rootAlbum, err := musicdomain.NewRootAlbum(nil, albums, tracks)
	if err != nil {
		return musicdomain.RootAlbum{}, errors.Wrap(err, "could not build the remote root album")
	}

	return rootAlbum, nil
}

func (l *RemoteLibrary) GetThumbnail(ctx context.Context, instanceId remotedomain.RemoteInstanceID, thumbnailId musicdomain.ThumbnailId) (io.ReadCloser, error) {
	target, err := l.targetByID(instanceId)
	if err != nil {
		return nil, err
	}

	body, err := l.client.GetThumbnail(ctx, target.address, target.token, thumbnailId.String())
	if err != nil {
		return nil, errors.Wrap(err, "could not get the remote thumbnail")
	}

	return body, nil
}

func (l *RemoteLibrary) GetTrackDuration(ctx context.Context, instanceId remotedomain.RemoteInstanceID, trackId musicdomain.TrackId) (musicdomain.TrackDuration, error) {
	target, err := l.targetByID(instanceId)
	if err != nil {
		return musicdomain.TrackDuration{}, err
	}

	resp, err := l.client.GetTrackDuration(ctx, target.address, target.token, trackId.String())
	if err != nil {
		return musicdomain.TrackDuration{}, errors.Wrap(err, "could not get the remote track duration")
	}

	duration, err := musicdomain.NewTrackDuration(time.Duration(resp.Duration * float64(time.Second)))
	if err != nil {
		return musicdomain.TrackDuration{}, errors.Wrap(err, "invalid duration")
	}

	return duration, nil
}

func (l *RemoteLibrary) StartTrackStream(ctx context.Context, instanceId remotedomain.RemoteInstanceID, trackId musicdomain.TrackId, seekPosition *musicdomain.RequestedSeekPosition) (musicdomain.StreamId, error) {
	target, err := l.targetByID(instanceId)
	if err != nil {
		return musicdomain.StreamId{}, err
	}

	var seekSeconds *float64
	if seekPosition != nil {
		s := musicdomain.NewSeekPosition(*seekPosition).Duration().Seconds()
		seekSeconds = &s
	}

	resp, err := l.client.StartTrackStream(ctx, target.address, target.token, trackId.String(), seekSeconds)
	if err != nil {
		return musicdomain.StreamId{}, errors.Wrap(err, "could not start the remote track stream")
	}

	streamId, err := musicdomain.NewStreamIdFromString(resp.StreamId)
	if err != nil {
		return musicdomain.StreamId{}, errors.Wrap(err, "invalid stream id")
	}

	return streamId, nil
}

func (l *RemoteLibrary) GetStreamPlaylist(ctx context.Context, instanceId remotedomain.RemoteInstanceID, trackId musicdomain.TrackId, streamId musicdomain.StreamId) (io.ReadCloser, error) {
	target, err := l.targetByID(instanceId)
	if err != nil {
		return nil, err
	}

	body, err := l.client.GetStreamPlaylist(ctx, target.address, target.token, trackId.String(), streamId.String())
	if err != nil {
		return nil, errors.Wrap(err, "could not get the remote stream playlist")
	}

	return body, nil
}

func (l *RemoteLibrary) GetStreamInit(ctx context.Context, instanceId remotedomain.RemoteInstanceID, trackId musicdomain.TrackId, streamId musicdomain.StreamId) (io.ReadCloser, error) {
	target, err := l.targetByID(instanceId)
	if err != nil {
		return nil, err
	}

	body, err := l.client.GetStreamInit(ctx, target.address, target.token, trackId.String(), streamId.String())
	if err != nil {
		return nil, errors.Wrap(err, "could not get the remote stream init")
	}

	return body, nil
}

func (l *RemoteLibrary) GetStreamFragment(ctx context.Context, instanceId remotedomain.RemoteInstanceID, trackId musicdomain.TrackId, streamId musicdomain.StreamId, fragmentId musicdomain.FragmentId) (io.ReadCloser, error) {
	target, err := l.targetByID(instanceId)
	if err != nil {
		return nil, err
	}

	body, err := l.client.GetStreamFragment(ctx, target.address, target.token, trackId.String(), streamId.String(), fragmentId.Int())
	if err != nil {
		return nil, errors.Wrap(err, "could not get the remote stream fragment")
	}

	return body, nil
}

func (l *RemoteLibrary) KeepAliveStream(ctx context.Context, instanceId remotedomain.RemoteInstanceID, trackId musicdomain.TrackId, streamId musicdomain.StreamId) error {
	target, err := l.targetByID(instanceId)
	if err != nil {
		return err
	}

	if err := l.client.KeepAliveStream(ctx, target.address, target.token, trackId.String(), streamId.String()); err != nil {
		return errors.Wrap(err, "could not keep alive the remote stream")
	}

	return nil
}

func (l *RemoteLibrary) GetAlbum(ctx context.Context, instanceId remotedomain.RemoteInstanceID, albumId musicdomain.AlbumId) (musicdomain.Album, error) {
	target, err := l.targetByID(instanceId)
	if err != nil {
		return musicdomain.Album{}, err
	}

	resp, err := l.client.GetAlbum(ctx, target.address, target.token, albumId.String())
	if err != nil {
		return musicdomain.Album{}, errors.Wrap(err, "could not get the remote album")
	}

	album, err := toRemoteAlbum(resp, instanceId)
	if err != nil {
		return musicdomain.Album{}, errors.Wrap(err, "could not convert the remote album")
	}

	return album, nil
}

func (l *RemoteLibrary) targetByID(id remotedomain.RemoteInstanceID) (remoteTarget, error) {
	targets, err := l.pairedTargets()
	if err != nil {
		return remoteTarget{}, errors.Wrap(err, "could not get the paired targets")
	}
	for _, t := range targets {
		if t.id == id {
			return t, nil
		}
	}
	return remoteTarget{}, appremote.ErrNotFound
}

func (l *RemoteLibrary) pairedTargets() ([]remoteTarget, error) {
	var targets []remoteTarget
	if err := l.transactionProvider.Read(func(r *appremote.TransactableRepositories) error {
		instances, err := r.RemoteInstances.GetAll()
		if err != nil {
			return errors.Wrap(err, "could not get the remote instances")
		}
		for _, instance := range instances {
			token, ok := instance.RemoteAuthToken()
			if !ok {
				continue
			}
			targets = append(targets, remoteTarget{
				id:      instance.Id(),
				address: instance.Address(),
				token:   token,
			})
		}
		return nil
	}); err != nil {
		return nil, errors.Wrap(err, "transaction failed")
	}
	return targets, nil
}

func (l *RemoteLibrary) Search(ctx context.Context, query string) (library.SearchResults, error) {
	targets, err := l.pairedTargets()
	if err != nil {
		return library.SearchResults{}, errors.Wrap(err, "could not get the paired targets")
	}

	var allFoundAlbums []library.FoundAlbum
	var allFoundTracks []library.FoundTrack

	for _, t := range targets {
		resp, err := l.client.Search(ctx, t.address, t.token, query)
		if err != nil {
			l.log.Error("could not search remote", "instance", t.id.String(), "err", err)
			continue
		}

		for _, a := range resp.Albums {
			album, err := toRemotePartialAlbum(a.Album, t.id)
			if err != nil {
				l.log.Error("could not convert remote search album", "instance", t.id.String(), "err", err)
				continue
			}
			allFoundAlbums = append(allFoundAlbums, library.FoundAlbum{Album: album, Dist: a.Score})
		}

		for _, tw := range resp.Tracks {
			track, err := toRemoteTrack(tw.Track, t.id)
			if err != nil {
				l.log.Error("could not convert remote search track", "instance", t.id.String(), "err", err)
				continue
			}
			var trackWithAlbum library.TrackWithAlbum
			if tw.Album != nil {
				album, err := toRemotePartialAlbum(*tw.Album, t.id)
				if err != nil {
					l.log.Error("could not convert remote search track album", "instance", t.id.String(), "err", err)
					continue
				}
				trackWithAlbum = library.NewTrackWithAlbum(track, album)
			} else {
				trackWithAlbum = library.NewRootTrackWithAlbum(track)
			}
			allFoundTracks = append(allFoundTracks, library.FoundTrack{Track: trackWithAlbum, Dist: tw.Score})
		}
	}

	return library.NewSearchResults(allFoundAlbums, allFoundTracks), nil
}

func toRemoteAlbum(a openapi.Album, instanceID remotedomain.RemoteInstanceID) (musicdomain.Album, error) {
	id, err := musicdomain.NewAlbumIdFromString(a.Id)
	if err != nil {
		return musicdomain.Album{}, errors.Wrap(err, "invalid album id")
	}
	title, err := musicdomain.NewAlbumTitle(a.Title)
	if err != nil {
		return musicdomain.Album{}, errors.Wrap(err, "invalid album title")
	}

	parents, err := toRemotePartialAlbums(a.Parents, instanceID)
	if err != nil {
		return musicdomain.Album{}, errors.Wrap(err, "could not convert the parents")
	}
	albums, err := toRemotePartialAlbums(a.Albums, instanceID)
	if err != nil {
		return musicdomain.Album{}, errors.Wrap(err, "could not convert the child albums")
	}

	tracks := make([]musicdomain.Track, 0, len(a.Tracks))
	for _, t := range a.Tracks {
		track, err := toRemoteTrack(t, instanceID)
		if err != nil {
			return musicdomain.Album{}, errors.Wrap(err, "could not convert a track")
		}
		tracks = append(tracks, track)
	}

	thumbnail, err := toRemoteThumbnail(a.Thumbnail)
	if err != nil {
		return musicdomain.Album{}, errors.Wrap(err, "could not convert the thumbnail")
	}

	return musicdomain.NewRemoteAlbum(id, title, thumbnail, parents, albums, tracks, instanceID)
}

func toRemotePartialAlbums(albums []openapi.PartialAlbum, instanceID remotedomain.RemoteInstanceID) ([]musicdomain.PartialAlbum, error) {
	out := make([]musicdomain.PartialAlbum, 0, len(albums))
	for _, a := range albums {
		album, err := toRemotePartialAlbum(a, instanceID)
		if err != nil {
			return nil, err
		}
		out = append(out, album)
	}
	return out, nil
}

func toRemotePartialAlbum(a openapi.PartialAlbum, instanceID remotedomain.RemoteInstanceID) (musicdomain.PartialAlbum, error) {
	id, err := musicdomain.NewAlbumIdFromString(a.Id)
	if err != nil {
		return musicdomain.PartialAlbum{}, errors.Wrap(err, "invalid album id")
	}
	title, err := musicdomain.NewAlbumTitle(a.Title)
	if err != nil {
		return musicdomain.PartialAlbum{}, errors.Wrap(err, "invalid album title")
	}
	thumbnail, err := toRemoteThumbnail(a.Thumbnail)
	if err != nil {
		return musicdomain.PartialAlbum{}, errors.Wrap(err, "could not convert the thumbnail")
	}

	return musicdomain.NewRemotePartialAlbum(id, title, thumbnail, instanceID), nil
}

func toRemoteThumbnail(t *openapi.Thumbnail) (*musicdomain.Thumbnail, error) {
	if t == nil {
		return nil, nil
	}
	id, err := musicdomain.NewThumbnailIdFromString(t.Id)
	if err != nil {
		return nil, errors.Wrap(err, "invalid thumbnail id")
	}
	thumb := musicdomain.NewThumbnail(id, musicdomain.FileId{})
	return &thumb, nil
}

// toRemoteTrack builds a remote track. The peer does not expose the internal
// file id (it is not part of the public album representation), so remote tracks
// carry a zero file id and cannot be streamed through this path yet.
func toRemoteTrack(t openapi.Track, instanceID remotedomain.RemoteInstanceID) (musicdomain.Track, error) {
	id, err := musicdomain.NewTrackIdFromString(t.Id)
	if err != nil {
		return musicdomain.Track{}, errors.Wrap(err, "invalid track id")
	}
	title, err := musicdomain.NewTrackTitle(t.Title)
	if err != nil {
		return musicdomain.Track{}, errors.Wrap(err, "invalid track title")
	}
	var number *musicdomain.TrackNumber
	if t.Number != nil {
		n, err := musicdomain.NewTrackNumber(*t.Number)
		if err != nil {
			return musicdomain.Track{}, errors.Wrap(err, "invalid track number")
		}
		number = &n
	}
	return musicdomain.NewRemoteTrack(id, musicdomain.FileId{}, number, title, instanceID), nil
}
