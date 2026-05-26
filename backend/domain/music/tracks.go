package music

import (
	"sort"
)

type Tracks struct {
	items []Track
}

func NewTracks(tracks []Track) Tracks {
	sorted := append([]Track(nil), tracks...)
	sortTracks(sorted)
	return Tracks{items: sorted}
}

func (t Tracks) Items() []Track {
	return t.items
}

func sortTracks(tracks []Track) {
	sort.Slice(tracks, func(i, j int) bool {
		ni := tracks[i].Number()
		nj := tracks[j].Number()

		if ni != nil && nj != nil {
			if ni.Int() != nj.Int() {
				return ni.Int() < nj.Int()
			}
		} else if ni != nil {
			return true
		} else if nj != nil {
			return false
		}

		return tracks[i].Title().String() < tracks[j].Title().String()
	})
}
