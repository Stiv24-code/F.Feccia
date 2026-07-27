package seeddemo

import (
	"embed"
	"encoding/csv"
	"fmt"
	"io"
	"strconv"

	"github.com/google/uuid"

	"fratelli-feccia/internal/models"
)

//go:embed data/destinations_poi.csv data/wash_stations_poi.csv
var poiFS embed.FS

// loadDestinationsPOI carica le destinazioni carico/scarico da un export reale
// Visirun (backend/internal/seeddemo/data/destinations_poi.csv, generato una
// tantum da "Report POI Visirun.xls" — righe con Categoria del P.O.I. ==
// "Carico-Scarico"), al posto della manciata di città curate a mano di prima.
func loadDestinationsPOI() ([]models.Destination, error) {
	rows, err := readPOICSV("data/destinations_poi.csv")
	if err != nil {
		return nil, err
	}
	out := make([]models.Destination, 0, len(rows))
	for _, r := range rows {
		lat, lng, err := parseLatLng(r["lat"], r["lng"])
		if err != nil {
			return nil, fmt.Errorf("destinations_poi.csv: %w", err)
		}
		out = append(out, models.Destination{
			ID:        uuid.New(),
			Nome:      r["nome"],
			Indirizzo: r["indirizzo"],
			Citta:     r["citta"],
			Provincia: r["provincia"],
			Nazione:   r["nazione"],
			Lat:       lat,
			Lng:       lng,
			Active:    true,
		})
	}
	return out, nil
}

// loadWashStationsPOI carica i punti di lavaggio dallo stesso export Visirun
// (data/wash_stations_poi.csv, righe con Categoria del P.O.I. == "Lavaggio").
// Il campo Tipo non è nell'export originale (nessuna colonna strutturata per
// il tipo di lavaggio) e resta vuoto — compilabile a mano da anagrafica.
func loadWashStationsPOI() ([]models.WashStation, error) {
	rows, err := readPOICSV("data/wash_stations_poi.csv")
	if err != nil {
		return nil, err
	}
	out := make([]models.WashStation, 0, len(rows))
	for _, r := range rows {
		lat, lng, err := parseLatLng(r["lat"], r["lng"])
		if err != nil {
			return nil, fmt.Errorf("wash_stations_poi.csv: %w", err)
		}
		out = append(out, models.WashStation{
			ID:        uuid.New(),
			Nome:      r["nome"],
			Indirizzo: r["indirizzo"],
			Citta:     r["citta"],
			Lat:       lat,
			Lng:       lng,
			Active:    true,
		})
	}
	return out, nil
}

func parseLatLng(latStr, lngStr string) (*float64, *float64, error) {
	lat, err := strconv.ParseFloat(latStr, 64)
	if err != nil {
		return nil, nil, fmt.Errorf("lat %q: %w", latStr, err)
	}
	lng, err := strconv.ParseFloat(lngStr, 64)
	if err != nil {
		return nil, nil, fmt.Errorf("lng %q: %w", lngStr, err)
	}
	return &lat, &lng, nil
}

// readPOICSV legge un CSV con intestazione e restituisce ogni riga come mappa
// colonna→valore, per non legare i chiamanti all'ordine delle colonne.
func readPOICSV(path string) ([]map[string]string, error) {
	f, err := poiFS.Open(path)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	defer f.Close()

	r := csv.NewReader(f)
	header, err := r.Read()
	if err != nil {
		return nil, fmt.Errorf("%s: header: %w", path, err)
	}

	var out []map[string]string
	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("%s: %w", path, err)
		}
		row := make(map[string]string, len(header))
		for i, col := range header {
			row[col] = record[i]
		}
		out = append(out, row)
	}
	return out, nil
}
