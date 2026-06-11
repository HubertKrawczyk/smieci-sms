package repository

import (
	"database/sql"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type DB struct {
	Conn *sql.DB
}

func NewDB(dataSource string) (*DB, error) {
	conn, err := sql.Open("pgx", dataSource)
	if err != nil {
		return nil, err
	}
	if err := conn.Ping(); err != nil {
		return nil, err
	}
	return &DB{Conn: conn}, nil
}

func (db *DB) Close() error {
	return db.Conn.Close()
}
