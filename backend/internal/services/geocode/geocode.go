// Package geocode exposes forward geocoding (address text -> coordinates)
// for anagrafica forms (Destination, Garage, WashStation) that let a user
// type an address and place the map marker instead of only clicking the map.
package geocode

import (
	"context"

	"fratelli-feccia/internal/dto"
	"fratelli-feccia/internal/services/geo"
)

type GeocodeService struct {
	geo *geo.GeoService
}

// NewGeocodeService takes no *gorm.DB: geocoding never touches the database
// (unlike route computation, which caches in RouteCache), so the shared
// geo.GeoService is constructed with a nil db — safe here since
// GeocodeSearch never dereferences it.
func NewGeocodeService(orsApiKey, orsBaseURL string) *GeocodeService {
	return &GeocodeService{geo: geo.NewGeoService(nil, orsApiKey, orsBaseURL)}
}

func (s *GeocodeService) Search(ctx context.Context, query string) ([]dto.GeocodeResultDTO, error) {
	results := s.geo.GeocodeSearch(ctx, query, 5)
	out := make([]dto.GeocodeResultDTO, len(results))
	for i, r := range results {
		out[i] = dto.GeocodeResultDTO{
			Label: r.Label, Indirizzo: r.Label, Citta: r.Locality,
			Cap: r.Postcode, Provincia: r.ProvinceA, Nazione: r.Country,
			Lat: r.Lat, Lng: r.Lng,
		}
	}
	return out, nil
}
