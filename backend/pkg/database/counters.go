package database

import "gorm.io/gorm"

// Counter mirrors Mongo's `counters` collection (backend/database.py's
// get_next_sequence): a single ever-incrementing counter per name, used to
// generate order/invoice progressivi. It never resets (not per-year).
type Counter struct {
	Name string `gorm:"primaryKey" json:"name"`
	Seq  int64  `gorm:"not null;default:0" json:"seq"`
}

// NextSequence atomically increments and returns the counter for name,
// creating it at 1 if it doesn't exist yet — mirrors get_next_sequence's
// find_one_and_update(upsert=True) semantics.
func NextSequence(db *gorm.DB, name string) (int64, error) {
	var seq int64
	err := db.Raw(`
		INSERT INTO counters (name, seq) VALUES (?, 1)
		ON CONFLICT (name) DO UPDATE SET seq = counters.seq + 1
		RETURNING seq
	`, name).Scan(&seq).Error
	return seq, err
}
