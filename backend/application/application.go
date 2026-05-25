package application

import (
	"github.com/boreq/eggplant/application/auth"
	"github.com/boreq/eggplant/application/music"
	"github.com/boreq/eggplant/application/queries"
)

type Application struct {
	Auth    auth.Auth
	Music   Music
	Queries Queries
}

type Music struct {
	Thumbnail      *music.ThumbnailHandler
	StartStreaming *music.StartStreamingHandler
	StreamPlaylist *music.StreamPlaylistHandler
	StreamInit     *music.StreamInitHandler
	StreamFragment *music.StreamFragmentHandler
	GetRootAlbum   *music.GetRootAlbumHandler
	GetAlbum       *music.GetAlbumHandler
	Search         *music.SearchHandler
	LoadLibrary    *music.LoadLibraryHandler
}

type Queries struct {
	Stats *queries.StatsHandler
}
