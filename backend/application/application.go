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
	Thumbnail       *music.LoggingThumbnailHandler
	StartStreaming  *music.LoggingStartStreamingHandler
	StreamPlaylist  *music.LoggingStreamPlaylistHandler
	StreamInit      *music.LoggingStreamInitHandler
	StreamFragment  *music.LoggingStreamFragmentHandler
	KeepAliveStream *music.LoggingKeepAliveStreamHandler
	GetRootAlbum    *music.LoggingGetRootAlbumHandler
	GetAlbum        *music.LoggingGetAlbumHandler
	Search          *music.LoggingSearchHandler
	LoadLibrary     *music.LoggingLoadLibraryHandler
}

type Queries struct {
	Stats *queries.StatsHandler
}
