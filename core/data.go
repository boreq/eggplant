package core

import (
	"crypto"

	"github.com/boreq/goaccess/parser"
)

func NewData() *Data {
	return &Data{
		ByReferer: []ByRefererData{},
		ByUri:     []ByUriData{},
	}
}

type ByRefererData struct {
	Referer string `json:"referer"`
	Visits  Set    `json:"visits"`
	Hits    int    `json:"hits"`
}

func (b *ByRefererData) InsertHits(hits int) {
	b.Hits += hits
}

func (b *ByRefererData) InsertVisit(visit string) {
	b.Visits.Add(visit)
}

type ByUriData struct {
	Uri      string         `json:"uri"`
	Visits   Set            `json:"visits"`
	ByStatus []ByStatusData `json:"statuses"`
}

func (b *ByUriData) InsertVisit(visit string) {
	b.Visits.Add(visit)
}

func (b *ByUriData) GetOrCreateByStatus(status string) *ByStatusData {
	byStatusData := b.findByStatus(status)
	if byStatusData == nil {
		b.ByStatus = cheapAppendByStatus(b.ByStatus, ByStatusData{
			Status: status,
			Hits:   0,
		})
		byStatusData = &b.ByStatus[len(b.ByStatus)-1]
	}
	return byStatusData
}

func (b *ByUriData) findByStatus(status string) *ByStatusData {
	for i := range b.ByStatus {
		if b.ByStatus[i].Status == status {
			return &b.ByStatus[i]
		}
	}
	return nil
}

type ByStatusData struct {
	Status string `json:"status"`
	Hits   int    `json:"hits"`
}

func (b *ByStatusData) InsertHits(hits int) {
	b.Hits += hits
}

type Data struct {
	ByReferer []ByRefererData `json:"referers"`
	ByUri     []ByUriData     `json:"uris"`
}

func (d *Data) Insert(entry *parser.Entry) error {
	visit := createVisitHash(entry)

	// By referer
	byRefererData := d.GetOrCreateByReferer(entry.Referer)
	byRefererData.InsertHits(1)
	byRefererData.InsertVisit(visit)

	// By URI
	byUriData := d.GetOrCreateByUri(entry.HttpRequestURI)
	byUriData.InsertVisit(visit)
	byStatusData := byUriData.GetOrCreateByStatus(entry.Status)
	byStatusData.InsertHits(1)

	return nil
}

func (d *Data) GetOrCreateByReferer(referer string) *ByRefererData {
	byRefererData := d.findByReferer(referer)
	if byRefererData == nil {
		d.ByReferer = cheapAppendByReferer(d.ByReferer, ByRefererData{
			Referer: referer,
			Visits:  NewSet(),
			Hits:    0,
		})
		byRefererData = &d.ByReferer[len(d.ByReferer)-1]
	}
	return byRefererData
}

func (d *Data) GetOrCreateByUri(uri string) *ByUriData {
	byUriData := d.findByUri(uri)
	if byUriData == nil {
		d.ByUri = cheapAppendByUri(d.ByUri, ByUriData{
			Uri:      uri,
			Visits:   NewSet(),
			ByStatus: []ByStatusData{},
		})
		byUriData = &d.ByUri[len(d.ByUri)-1]
	}
	return byUriData
}

func (d *Data) findByReferer(referer string) *ByRefererData {
	for i := range d.ByReferer {
		if d.ByReferer[i].Referer == referer {
			return &d.ByReferer[i]
		}
	}
	return nil
}

func (d *Data) findByUri(uri string) *ByUriData {
	for i := range d.ByUri {
		if d.ByUri[i].Uri == uri {
			return &d.ByUri[i]
		}
	}
	return nil
}

var visitHash = crypto.SHA512_256

const retainHashBytes = 8

func createVisitHash(entry *parser.Entry) string {
	h := visitHash.New()
	h.Write([]byte(entry.RemoteAddress))
	h.Write([]byte(entry.UserAgent))
	sum := h.Sum(nil)
	return string(sum)[:retainHashBytes]
}

func cheapAppendByReferer(slice []ByRefererData, value ByRefererData) []ByRefererData {
	newSlice := make([]ByRefererData, len(slice)+1, len(slice)+1)
	copy(newSlice, slice)
	newSlice[len(newSlice)-1] = value
	return newSlice
}

func cheapAppendByUri(slice []ByUriData, value ByUriData) []ByUriData {
	newSlice := make([]ByUriData, len(slice)+1, len(slice)+1)
	copy(newSlice, slice)
	newSlice[len(newSlice)-1] = value
	return newSlice
}

func cheapAppendByStatus(slice []ByStatusData, value ByStatusData) []ByStatusData {
	newSlice := make([]ByStatusData, len(slice)+1, len(slice)+1)
	copy(newSlice, slice)
	newSlice[len(newSlice)-1] = value
	return newSlice
}
