package core

import (
	"crypto"

	"github.com/boreq/plum/parser"
)

func NewData() *Data {
	return &Data{
		Referers: make(map[string]*RefererData),
		Uris:     make(map[string]*UriData),
		Visits:   NewSet(),
	}
}

type Data struct {
	Referers map[string]*RefererData `json:"referers"`
	Uris     map[string]*UriData     `json:"uris"`
	Visits   Set                     `json:"visits"`
}

func (d *Data) Insert(entry *parser.Entry) error {
	visit := createVisitHash(entry)

	d.InsertVisit(visit)

	refererData := d.GetOrCreateRefererData(entry.Referer)
	refererData.InsertHits(1)
	refererData.InsertVisit(visit)

	uriData := d.GetOrCreateUriData(entry.HttpRequestURI)
	uriData.InsertVisit(visit)
	uriData.AddBodyBytesSent(entry.BodyBytesSent)
	statusData := uriData.GetOrCreateStatusData(entry.Status)
	statusData.InsertHits(1)

	return nil
}

func (d *Data) GetOrCreateRefererData(referer string) *RefererData {
	refererData, ok := d.Referers[referer]
	if !ok {
		refererData = &RefererData{
			Visits: NewSet(),
			Hits:   0,
		}
		d.Referers[referer] = refererData
	}
	return refererData
}

func (d *Data) InsertVisit(visit string) {
	d.Visits.Add(visit)
}

func (d *Data) GetOrCreateUriData(uri string) *UriData {
	uriData, ok := d.Uris[uri]
	if !ok {
		uriData = &UriData{
			Visits:   NewSet(),
			Statuses: make(map[string]*StatusData),
		}
		d.Uris[uri] = uriData
	}
	return uriData
}

type RefererData struct {
	Visits Set `json:"visits"`
	Hits   int `json:"hits"`
}

func (b *RefererData) InsertHits(hits int) {
	b.Hits += hits
}

func (b *RefererData) InsertVisit(visit string) {
	b.Visits.Add(visit)
}

type UriData struct {
	Visits        Set                    `json:"visits"`
	BodyBytesSent int                    `json:"bytes"`
	Statuses      map[string]*StatusData `json:"statuses"`
}

func (b *UriData) InsertVisit(visit string) {
	b.Visits.Add(visit)
}

func (b *UriData) AddBodyBytesSent(bytes int) {
	b.BodyBytesSent += bytes
}

func (b *UriData) GetOrCreateStatusData(status string) *StatusData {
	statusData, ok := b.Statuses[status]
	if !ok {
		statusData = &StatusData{
			Hits: 0,
		}
		b.Statuses[status] = statusData
	}
	return statusData
}

type StatusData struct {
	Hits int `json:"hits"`
}

func (b *StatusData) InsertHits(hits int) {
	b.Hits += hits
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
