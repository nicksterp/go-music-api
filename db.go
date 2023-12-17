package main

import (
	"database/sql"
	"errors"
	"fmt"
	"os"

	_ "github.com/lib/pq"
)

func connectToDB() (*sql.DB, error) {
	connStr, err := createDBStr()
	if err != nil {
		return nil, err
	}

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	return db, nil
}

type DBVars struct {
	host string
	user string
	pass string
	name string
	env  string
}

func createDBStr() (string, error) {
	vars := DBVars{
		host: os.Getenv("DB_HOST"),
		user: os.Getenv("DB_USER"),
		pass: os.Getenv("DB_PASS"),
		name: os.Getenv("DB_NAME"),
		env:  os.Getenv("ENV"),
	}

	// If any of the variables are empty, return an error
	if vars.host == "" || vars.user == "" || vars.pass == "" || vars.name == "" || vars.env == "" {
		return "", errors.New("missing DB variable")
	}

	connStr := fmt.Sprintf("postgres://%s:%s@%s/%s", vars.user, vars.pass, vars.host, vars.name)

	// Default to dev environment, append ?sslmode=disable
	if vars.env != "prod" {
		connStr += "?sslmode=disable"
	}

	return connStr, nil
}
