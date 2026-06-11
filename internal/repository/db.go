package repository

import (
    "database/sql"
    _ "github.com/mattn/go-sqlite3"
)

type DB struct {
    Conn *sql.DB
}

func NewDB(dataSource string) (*DB, error) {
    conn, err := sql.Open("sqlite3", dataSource)
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
