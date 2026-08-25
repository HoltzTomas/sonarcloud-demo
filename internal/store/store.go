package store

import (
	"database/sql"

	// Blank import: registers the pure-Go "sqlite" driver with database/sql so
	// that sql.Open("sqlite", ...) below can resolve it.
	_ "modernc.org/sqlite"
)

// Open creates the in-memory demo database and seeds it with contracts.
func Open() (*sql.DB, error) {
	db, err := sql.Open("sqlite", "file:contracts?mode=memory&cache=shared")
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)

	schema := `
CREATE TABLE contracts (
	id       TEXT PRIMARY KEY,
	customer TEXT NOT NULL,
	region   TEXT NOT NULL,
	amount   INTEGER NOT NULL,
	owner    TEXT NOT NULL,
	notes    TEXT NOT NULL
);
CREATE TABLE api_keys (
	name  TEXT PRIMARY KEY,
	token TEXT NOT NULL
);`
	if _, err := db.Exec(schema); err != nil {
		return nil, err
	}

	contracts := [][]interface{}{
		{"C-1001", "Iberdrola", "ES-01", 480000, "ana.gomez@example.com", "Suministro trimestral"},
		{"C-1002", "Naturgy", "ES-02", 1250000, "ana.gomez@example.com", "Renovacion anual"},
		{"C-1003", "Cepsa", "PT-01", 95000, "luis.martin@example.com", "Piloto logistico"},
		{"C-1004", "Endesa", "ES-01", 2100000, "luis.martin@example.com", "Contrato marco"},
	}
	for _, c := range contracts {
		if _, err := db.Exec("INSERT INTO contracts (id, customer, region, amount, owner, notes) VALUES (?, ?, ?, ?, ?, ?)", c...); err != nil {
			return nil, err
		}
	}

	if _, err := db.Exec("INSERT INTO api_keys (name, token) VALUES (?, ?)", "gis-integration", "gis_live_9f2b41c7de6a4f0e"); err != nil {
		return nil, err
	}
	return db, nil
}
