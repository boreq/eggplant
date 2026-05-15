package domain

import (
	"sort"
	"strconv"
	"strings"
	"unicode"
)

type Tracks []Track

func NewTracks(tracks []Track) Tracks {
	sorted := append([]Track(nil), tracks...)
	sortTracks(sorted)
	return Tracks(sorted)
}

func sortTracks(tracks []Track) {
	sort.Slice(tracks, func(i, j int) bool {
		titleI := tracks[i].Title().String()
		titleJ := tracks[j].Title().String()

		fieldsI := strings.Fields(titleI)
		fieldsJ := strings.Fields(titleJ)

		if len(fieldsI) > 0 && len(fieldsJ) > 0 {
			f := func(r rune) bool { return !unicode.IsNumber(r) }
			numI, errI := strconv.Atoi(strings.TrimFunc(fieldsI[0], f))
			numJ, errJ := strconv.Atoi(strings.TrimFunc(fieldsJ[0], f))
			if errI == nil && errJ == nil {
				if numI == numJ {
					return strings.Join(fieldsI[1:], "") < strings.Join(fieldsJ[1:], "")
				}
				return numI < numJ
			}
		}
		return titleI < titleJ
	})
}
