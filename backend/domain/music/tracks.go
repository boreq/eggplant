package music

import (
	"slices"
	"strings"
)

type Tracks struct {
	items []Track
}

func NewTracks(tracks []Track) Tracks {
	sorted := slices.Clone(tracks)
	sortTracks(sorted)
	return Tracks{items: sorted}
}

func (t Tracks) Items() []Track {
	return t.items
}

func sortTracks(tracks []Track) {
	slices.SortFunc(tracks, func(a, b Track) int {
		na := a.Number()
		nb := b.Number()

		if na != nil && nb != nil {
			if na.Int() != nb.Int() {
				if na.Int() < nb.Int() {
					return -1
				}
				return 1
			}
		} else if na != nil {
			return -1
		} else if nb != nil {
			return 1
		}

		return strings.Compare(a.Title().String(), b.Title().String())
	})
}
