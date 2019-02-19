package server

import (
	"github.com/boreq/flightradar-backend/storage"
	"testing"
	"time"
)

func BenchmarkPolar(b *testing.B) {
	fakeStoredData := createFakeStoredData()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		toPolar(fakeStoredData)
	}
}

func createFakeStoredData() []storage.StoredData {
	rv := make([]storage.StoredData, 0, 1000000)
	for lat := 49.0; lat < 50.0; lat += 0.01 {
		for lon := 19.0; lon < 21.0; lon += 0.01 {
			for _, icao := range []string{"aaaaaa", "bbbbbb", "cccccc"} {
				for _, flightNumber := range []string{"aaaaaa", "bbbbbb", "cccccc"} {
					for _, altitude := range []int{2000, 8000, 30000} {
						data := storage.Data{
							Icao:         new(string),
							FlightNumber: new(string),
							Altitude:     new(int),
							Latitude:     new(float64),
							Longitude:    new(float64),
						}
						*data.Icao = icao
						*data.FlightNumber = flightNumber
						*data.Altitude = altitude
						*data.Latitude = lat
						*data.Longitude = lon
						storedData := storage.StoredData{
							Time: time.Now(),
							Data: data,
						}
						rv = append(rv, storedData)
					}
				}
			}
		}
	}
	return rv
}
