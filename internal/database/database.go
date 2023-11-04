package database

import (
	"encoding/json"
	"errors"
	"flag"
	"log"
	"os"
	"sort"
	"sync"
)

const dbFile = "database.json"

type DB struct {
	mu *sync.RWMutex
}

type Chirp struct {
	ID   int    `json:"id"`
	Body string `json:"body"`
}

type User struct {
	ID    int    `json:"id"`
	Email string `json:"email"`
}

type Schema struct {
	Chirps map[int]Chirp `json:"chirps"`
	Users  map[int]User  `json:"users"`
}

func NewDB() DB {
	dbg := flag.Bool("debug", false, "Enable debug mode")
	flag.Parse()

	// Clear db if exists
	if *dbg {
		os.Remove(dbFile)
	}

	return DB{
		mu: &sync.RWMutex{},
	}
}

func (db *DB) ListChirps() ([]Chirp, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	schema, err := readDB()
	if err != nil {
		log.Println(err)
		return []Chirp{}, err
	}

	chirpList := []Chirp{}
	for _, chirp := range schema.Chirps {
		chirpList = append(chirpList, chirp)
	}

	sort.Slice(chirpList, func(i, j int) bool {
		return chirpList[i].ID < chirpList[j].ID
	})

	return chirpList, nil
}

func (db *DB) ReadChirp(ID int) (Chirp, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	schema, err := readDB()
	if err != nil {
		log.Println(err)
		return Chirp{}, err
	}

	chirp, ok := schema.Chirps[ID]
	if !ok {
		return Chirp{}, nil
	}

	return chirp, nil
}

func (db *DB) CreateChirp(body string) (Chirp, error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	schema, err := readDB()
	if err != nil {
		log.Println(err)
		return Chirp{}, err
	}

	chirp := Chirp{
		ID:   len(schema.Chirps) + 1,
		Body: body,
	}

	schema.Chirps[chirp.ID] = chirp

	err = saveDB(schema)
	if err != nil {
		log.Println(err)
		return Chirp{}, err
	}

	return chirp, nil
}

func (db *DB) CreateUser(email string) (User, error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	schema, err := readDB()
	if err != nil {
		log.Println(err)
		return User{}, err
	}

	user := User{
		ID:    len(schema.Users) + 1,
		Email: email,
	}

	schema.Users[user.ID] = user

	err = saveDB(schema)
	if err != nil {
		log.Println(err)
		return User{}, err
	}

	return user, nil
}

func readDB() (Schema, error) {
	_, err := os.Stat(dbFile)
	if err != nil {
		return Schema{
			Chirps: make(map[int]Chirp),
			Users:  make(map[int]User),
		}, nil
	}

	data, err := os.ReadFile(dbFile)
	if err != nil {
		return Schema{}, errors.New("could not read database")
	}

	schema := Schema{}
	err = json.Unmarshal(data, &schema)
	if err != nil {
		return schema, errors.New("could not parse json")
	}

	return schema, nil
}

func saveDB(schema Schema) error {
	file, err := os.Create(dbFile)
	if err != nil {
		return errors.New("could not create the database file")
	}
	defer file.Close()

	encoder := json.NewEncoder(file)

	err = encoder.Encode(schema)
	if err != nil {
		return errors.New("there was a problem saving the list")
	}

	return nil
}
