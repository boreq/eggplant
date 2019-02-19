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

func (b *ByRefererData) Insert(entry *parser.Entry) error {
	visit := createVisitHash(entry)
	b.Hits++
	b.Visits.Add(visit)
	return nil
}

type ByUriData struct {
	Uri      string         `json:"uri"`
	Visits   Set            `json:"visits"`
	ByStatus []ByStatusData `json:"by_status"`
}

func (b *ByUriData) Insert(entry *parser.Entry) error {
	visit := createVisitHash(entry)
	b.Visits.Add(visit)

	// By status
	byStatusData := b.findByStatus(entry.Status)
	if byStatusData == nil {
		b.ByStatus = cheapAppendByStatus(b.ByStatus, ByStatusData{
			Status: entry.Status,
			Hits:   0,
		})
		byStatusData = &b.ByStatus[len(b.ByStatus)-1]
	}
	byStatusData.Hits++
	return nil
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

type Data struct {
	ByReferer []ByRefererData `json:"by_referer"`
	ByUri     []ByUriData     `json:"by_uri"`
}

func (d *Data) Insert(entry *parser.Entry) error {
	// By referer
	byRefererData := d.findByReferer(entry.Referer)
	if byRefererData == nil {
		d.ByReferer = cheapAppendByReferer(d.ByReferer, ByRefererData{
			Referer: entry.Referer,
			Visits:  NewSet(),
			Hits:    0,
		})
		byRefererData = &d.ByReferer[len(d.ByReferer)-1]
	}
	if err := byRefererData.Insert(entry); err != nil {
		return err
	}

	// By URI
	byUriData := d.findByUri(entry.HttpRequestURI)
	if byUriData == nil {
		d.ByUri = cheapAppendByUri(d.ByUri, ByUriData{
			Uri:      entry.HttpRequestURI,
			Visits:   NewSet(),
			ByStatus: []ByStatusData{},
		})
		byUriData = &d.ByUri[len(d.ByUri)-1]
	}
	if err := byUriData.Insert(entry); err != nil {
		return err
	}

	return nil
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
