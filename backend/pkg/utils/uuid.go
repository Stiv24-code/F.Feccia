package utils

import "github.com/google/uuid"

// ParseUUID parses a required UUID reference coming from a request DTO
// (e.g. OrderRequest.ClienteID), returning a 400 APIError on malformed input
// instead of a raw parse error.
func ParseUUID(s string) (uuid.UUID, error) {
	id, err := uuid.Parse(s)
	if err != nil {
		return uuid.UUID{}, NewAPIError(400, "ID non valido: "+s)
	}
	return id, nil
}

// ParseOptionalUUID parses an optional UUID reference (e.g. OrderRequest.GarageID,
// legitimately empty until assigned) — nil, nil on an empty string, a 400
// APIError on malformed non-empty input.
func ParseOptionalUUID(s string) (*uuid.UUID, error) {
	if s == "" {
		return nil, nil
	}
	id, err := ParseUUID(s)
	if err != nil {
		return nil, err
	}
	return &id, nil
}
