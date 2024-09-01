package main

import (
	"encoding/json"
	"errors"
	"io/fs"
	"log"
	"os"
	"sync"
)

type DB struct {
	path   string
	mux    *sync.RWMutex
	chirps int
}

type DBStructure struct {
	Chirps map[int]Chirp `json:"chirps"`
}

func NewDB(path string) (*DB, error) {
	db := DB{
		path: path,
		mux:  &sync.RWMutex{},
	}
	err := db.ensureDB()
	if err != nil {
		log.Fatal("DB load failed")
	}

	return &db, nil
}

func (db *DB) ensureDB() error {
	_, err := os.ReadFile(db.path)
	if errors.Is(err, fs.ErrNotExist) {
		_, err := os.Create(db.path)
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	}

	return nil
}

func (db *DB) loadDB() (DBStructure, error) {
	data, err := os.ReadFile(db.path)
	if err != nil {
		return DBStructure{}, err
	}
	if len(data) == 0 {
		return DBStructure{
			Chirps: map[int]Chirp{},
		}, nil
	}
	dbData := DBStructure{}
	err = json.Unmarshal(data, &dbData)
	if err != nil {
		log.Printf("Error decoding db file: %s", err)
		return DBStructure{}, err
	}

	return dbData, nil
}

func (db *DB) writeDB(dbStructure DBStructure) error {
	data, err := json.Marshal(dbStructure)
	if err != nil {
		log.Printf("Error marshalling: %s", err)
		return err
	}

	err = os.WriteFile(db.path, data, 0666)
	if err != nil {
		log.Printf("Error writing file: %s", err)
		return err
	}
	return nil
}
