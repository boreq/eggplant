package music

import (
	"github.com/boreq/eggplant/domain/music"
	"github.com/boreq/eggplant/domain/music/library"
	"github.com/boreq/errors"
)

type StreamFragment struct {
	trackId    music.TrackId
	streamId   music.StreamId
	fragmentId music.FragmentId
}

func NewStreamFragment(trackId music.TrackId, streamId music.StreamId, fragmentId music.FragmentId) StreamFragment {
	return StreamFragment{
		trackId:    trackId,
		streamId:   streamId,
		fragmentId: fragmentId,
	}
}

func (s StreamFragment) TrackId() music.TrackId {
	return s.trackId
}

func (s StreamFragment) StreamId() music.StreamId {
	return s.streamId
}

func (s StreamFragment) FragmentId() music.FragmentId {
	return s.fragmentId
}

type StreamFragmentHandler struct {
	libraryRepository LibraryRepository
	trackConverter    TrackConverter
}

func NewStreamFragmentHandler(libraryRepository LibraryRepository, trackConverter TrackConverter) *StreamFragmentHandler {
	return &StreamFragmentHandler{
		libraryRepository: libraryRepository,
		trackConverter:    trackConverter,
	}
}

func (h *StreamFragmentHandler) Execute(accessCtx library.AccessContext, cmd StreamFragment) (ConvertedFile, error) {
	lib, err := h.libraryRepository.Get()
	if err != nil {
		return ConvertedFile{}, errors.Wrap(err, "could not get the library")
	}

	track, err := lib.GetTrack(accessCtx, cmd.TrackId())
	if err != nil {
		return ConvertedFile{}, errors.Wrap(err, "could not get the track")
	}

	fileId, ok := track.FileId()
	if !ok {
		return ConvertedFile{}, errors.New("track has no file id")
	}

	cf, err := h.trackConverter.GetFragment(fileId, cmd.StreamId(), cmd.FragmentId())
	if err != nil {
		return ConvertedFile{}, errors.Wrap(err, "could not get the fragment")
	}
	return cf, nil
}
