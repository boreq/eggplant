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
	Thumbnail    *music.ThumbnailHandler
	Track        *music.TrackHandler
	GetRootAlbum *music.GetRootAlbumHandler
	GetAlbum     *music.GetAlbumHandler
	Search       *music.SearchHandler
	BuildLibrary *music.BuildLibraryHandler
}

type Queries struct {
	Stats *queries.StatsHandler
}
