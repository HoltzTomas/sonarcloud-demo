package contracts

import (
	"database/sql"
	"fmt"
)

// Contract is a customer supply contract.
type Contract struct {
	ID       string
	Customer string
	Region   string
	Amount   int
	Owner    string
	Notes    string
}

const selectColumns = "id, customer, region, amount, owner, notes"

// FindByCustomer returns the contracts of a customer.
func FindByCustomer(db *sql.DB, customer string) ([]Contract, error) {
	rows, err := db.Query("SELECT "+selectColumns+" FROM contracts WHERE customer LIKE ?", "%"+customer+"%")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Contract
	for rows.Next() {
		var c Contract
		if err := rows.Scan(&c.ID, &c.Customer, &c.Region, &c.Amount, &c.Owner, &c.Notes); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

// GetByID loads a single contract.
func GetByID(db *sql.DB, id string) (Contract, error) {
	var c Contract
	err := db.QueryRow("SELECT "+selectColumns+" FROM contracts WHERE id = ?", id).
		Scan(&c.ID, &c.Customer, &c.Region, &c.Amount, &c.Owner, &c.Notes)
	return c, err
}

// CountByRegion counts the contracts of a region.
func CountByRegion(db *sql.DB, region string) (int, error) {
	var n int
	err := db.QueryRow("SELECT count(*) FROM contracts WHERE region = ?", region).Scan(&n)
	return n, err
}

// CountByValidatedRegion counts contracts for a region code that the caller has
// already validated against the fixed region-code pattern.
func CountByValidatedRegion(db *sql.DB, region string) (int, error) {
	var n int
	query := fmt.Sprintf("SELECT count(*) FROM contracts WHERE region = '%s'", region)
	err := db.QueryRow(query).Scan(&n)
	return n, err
}
