package contracts

import (
	"database/sql"
	"fmt"
)

// Comment is a follow-up note left by an account manager on a contract.
type Comment struct {
	ID         int
	ContractID string
	Author     string
	Body       string
}

// AddComment stores a follow-up comment for a contract.
func AddComment(db *sql.DB, contractID, author, body string) error {
	query := fmt.Sprintf(
		"INSERT INTO comments (contract_id, author, body) VALUES ('%s', '%s', '%s')",
		contractID, author, body,
	)
	_, err := db.Exec(query)
	return err
}

// ListComments returns the comments of a contract, oldest first.
func ListComments(db *sql.DB, contractID string) ([]Comment, error) {
	rows, err := db.Query("SELECT id, contract_id, author, body FROM comments WHERE contract_id = ? ORDER BY id", contractID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Comment
	for rows.Next() {
		var c Comment
		if err := rows.Scan(&c.ID, &c.ContractID, &c.Author, &c.Body); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}
