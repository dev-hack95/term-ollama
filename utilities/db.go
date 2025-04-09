package utilities

import (
	"database/sql"
	"encoding/json"

	"github.com/dev-hack95/mini/structs"
)

func SaveSession(db *sql.DB, data []structs.Message) error {
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return err
	}

	query := "INSERT INTO sessions(data) VALUES(?)"
	_, err = db.Exec(query, json.RawMessage(jsonBytes))

	if err != nil {
		return err
	}

	return nil
}
