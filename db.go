package main

import (
	"encoding/json"
	"errors"
	"io/fs"
	"log"
	"os"
	"sort"
	"sync"
)

type DB struct {
	path string
	mux  *sync.RWMutex
}

type DBStructure struct {
	Chirps        map[int]Chirp           `json:"chirps"`
	Users         map[int]User            `json:"users"`
	Emails        map[string]int          `json:"emails"`
	ChirpId       int                     `json:"chirpId"`
	UserId        int                     `json:"userId"`
	RefreshTokens map[string]RefreshToken `json:"refreshTokens"`
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
			Chirps:        map[int]Chirp{},
			Users:         map[int]User{},
			Emails:        map[string]int{},
			ChirpId:       0,
			UserId:        0,
			RefreshTokens: map[string]RefreshToken{},
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

func (db *DB) GetChirps() ([]Chirp, error) {
	db.mux.RLock()
	defer db.mux.RUnlock()

	chirps, err := db.loadDB()
	if err != nil {
		log.Printf("failed to get chirps: %s", err)
		return []Chirp{}, err
	}
	keys := []int{}
	for k := range chirps.Chirps {
		keys = append(keys, k)
	}
	sort.Ints(keys)

	chirpsSlice := []Chirp{}
	for _, k := range keys {
		chirpsSlice = append(chirpsSlice, chirps.Chirps[k])
	}

	return chirpsSlice, nil
}
