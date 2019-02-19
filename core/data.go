package core

import (
	"crypto"

	"github.com/boreq/goaccess/logging"
	"github.com/boreq/goaccess/parser"
)

func NewData() *Data {
	return &Data{
		ByReferer: make(map[string]*ByRefererData),
		ByUri:     make(map[string]*ByUriData),
		log:       logging.New("data"),
	}
}

type ByRefererData struct {
	Visits *Set `json:"visits"`
	Hits   int  `json:"hits"`
}

func (b *ByRefererData) Insert(entry *parser.Entry) error {
	visit := createVisitHash(entry)
	b.Hits++
	b.Visits.Add(visit)
	return nil
}

type ByUriData struct {
	Visits   *Set           `json:"visits"`
	ByStatus map[string]int `json:"by_status"`
}

func (b *ByUriData) Insert(entry *parser.Entry) error {
	visit := createVisitHash(entry)
	b.Visits.Add(visit)

	value, ok := b.ByStatus[entry.Status]
	if !ok {
		value = 0
	}
	b.ByStatus[entry.Status] = value + 1

	return nil
}

type Data struct {
	ByReferer map[string]*ByRefererData `json:"by_referer"`
	ByUri     map[string]*ByUriData     `json:"by_uri"`
	log       logging.Logger
}

func (d *Data) Insert(entry *parser.Entry) error {
	// By referer
	byRefererData, ok := d.ByReferer[entry.Referer]
	if !ok {
		byRefererData = &ByRefererData{
			Visits: NewSet(),
			Hits:   0,
		}
		d.ByReferer[entry.Referer] = byRefererData
	}
	if err := byRefererData.Insert(entry); err != nil {
		return err
	}

	// By URI
	byUriData, ok := d.ByUri[entry.HttpRequestURI]
	if !ok {
		byUriData = &ByUriData{
			Visits:   NewSet(),
			ByStatus: make(map[string]int),
		}
		d.ByUri[entry.HttpRequestURI] = byUriData
	}
	if err := byUriData.Insert(entry); err != nil {
		return err
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
