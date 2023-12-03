package database

import (
	"encoding/json"
	"errors"
	"flag"
	"log"
	"os"
	"sort"
	"sync"

	"golang.org/x/crypto/bcrypt"
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
	ID       int    `json:"id"`
	Email    string `json:"email"`
	Password string `json:"password,omitempty"`
}

type Schema struct {
	Chirps        map[int]Chirp   `json:"chirps"`
	Users         map[int]User    `json:"users"`
	RefreshTokens map[string]bool `json:"refresh_tokens"`
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

func (db *DB) CreateUser(email string, password string) (User, error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	schema, err := readDB()
	if err != nil {
		log.Println(err)
		return User{}, err
	}

	user, err := findUserByEmail(schema.Users, email)
	if err == nil {
		return User{}, errors.New("user email already exists")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), 0)
	if err != nil {
		return User{}, errors.New("problem saving password")
	}

	user = User{
		ID:       len(schema.Users) + 1,
		Email:    email,
		Password: string(hash),
	}

	schema.Users[user.ID] = user

	err = saveDB(schema)
	if err != nil {
		log.Println(err)
		return User{}, err
	}

	user.Password = ""
	return user, nil
}

func (db *DB) UpdateUser(userId int, email string, password string) (User, error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	schema, err := readDB()
	if err != nil {
		log.Println(err)
		return User{}, err
	}

	user, err := findUserById(schema.Users, userId)
	if err != nil {
		return User{}, errors.New("user does not exist")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), 0)
	if err != nil {
		return User{}, errors.New("problem saving password")
	}

	user = User{
		ID:       user.ID,
		Email:    email,
		Password: string(hash),
	}

	schema.Users[user.ID] = user

	err = saveDB(schema)
	if err != nil {
		log.Println(err)
		return User{}, err
	}

	user.Password = ""
	return user, nil
}

func (db *DB) Login(email string, password string) (User, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	schema, err := readDB()
	if err != nil {
		log.Println(err)
		return User{}, err
	}

	user, err := findUserByEmail(schema.Users, email)
	if err != nil {
		return User{}, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return User{}, errors.New("incorrect credentials")
	}

	user.Password = ""
	return user, nil
}

func (db *DB) SaveRefreshToken(token string) error {
	db.mu.RLock()
	defer db.mu.RUnlock()

	schema, err := readDB()
	if err != nil {
		log.Println(err)
		return err
	}

	schema.RefreshTokens[token] = false

	err = saveDB(schema)
	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}

func (db *DB) CheckRefreshToken(token string) bool {
	db.mu.RLock()
	defer db.mu.RUnlock()

	schema, err := readDB()
	if err != nil {
		log.Println(err)
		return false
	}

	revoked, ok := schema.RefreshTokens[token]
	if revoked || !ok {
		return false
	} else {
		return true
	}
}

func (db *DB) RevokeRefreshToken(token string) error {
	db.mu.RLock()
	defer db.mu.RUnlock()

	schema, err := readDB()
	if err != nil {
		log.Println(err)
		return err
	}

	schema.RefreshTokens[token] = true

	err = saveDB(schema)
	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}

func findUserByEmail(users map[int]User, email string) (User, error) {
	for _, user := range users {
		if user.Email == email {
			return user, nil
		}
	}

	return User{}, errors.New("user does not exist")
}

func findUserById(users map[int]User, id int) (User, error) {
	for _, user := range users {
		if user.ID == id {
			return user, nil
		}
	}

	return User{}, errors.New("user does not exist")
}

func readDB() (Schema, error) {
	_, err := os.Stat(dbFile)
	if err != nil {
		return Schema{
			Chirps:        map[int]Chirp{},
			Users:         map[int]User{},
			RefreshTokens: map[string]bool{},
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
