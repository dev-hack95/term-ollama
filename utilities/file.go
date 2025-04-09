package utilities

import (
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type clientConfig struct {
	Model  string `json:"model"`
	Url    string `json:"url"`
	Role   string `json:"role"`
	Stream bool   `json:"stream"`
}

func CheckFileExist(filePath string) (clientConfig, error) {
	var data clientConfig

	homeDir, err1 := os.UserHomeDir()
	if err1 != nil {
		return data, err1
	}

	filePath = homeDir + filePath
	_, err := os.Stat(filePath)

	if err != nil {

		dirPath := filepath.Join(homeDir, ".term-ollama")

		err2 := os.MkdirAll(dirPath, 0755)

		if err2 != nil {
			return data, err2
		}

		file_path := filepath.Join(dirPath, "model.yaml")

		clientConfig := clientConfig{
			Model:  "granite3.1-moe:latest",
			Url:    "http://127.0.0.1:11434/api/chat",
			Role:   "user",
			Stream: false,
		}

		yamlData, err3 := yaml.Marshal(&clientConfig)

		if err3 != nil {
			return data, err3
		}

		err4 := os.WriteFile(file_path, yamlData, 0755)

		if err4 != nil {
			return data, err4
		}
	}

	yamlFile, err := os.ReadFile(filePath)

	if err != nil {
		fmt.Println("debug2")
		return data, err
	}

	err = yaml.Unmarshal(yamlFile, &data)

	if err != nil {
		return data, err
	}

	return data, nil
}

func SessionDB(filePath string) (*sql.DB, error) {
	var db *sql.DB

	homeDir, err1 := os.UserHomeDir()
	if err1 != nil {
		return nil, err1
	}

	filePath = homeDir + filePath
	_, err := os.Stat(filePath)

	if err != nil {

		dirPath := filepath.Join(homeDir, ".term-ollama/db")

		err2 := os.MkdirAll(dirPath, 0755)

		if err2 != nil {
			return nil, err2
		}

		file_path := filepath.Join(dirPath, "sessions.db")

		file, err := os.Create(file_path)

		if err != nil {
			return nil, err
		}

		defer file.Close()

	}

	db, err = sql.Open("sqlite3", filePath)

	if err != nil {
		return nil, err
	}

	_, err = db.Exec("CREATE TABLE IF NOT EXISTS sessions (id INTEGER PRIMARY KEY AUTOINCREMENT, data JSON, created_at TIMESTAMP DEFAULT (datetime('now', 'localtime')), active BOOLEAN DEFAULT TRUE)")

	if err != nil {
		return nil, err
	}

	return db, nil
}
