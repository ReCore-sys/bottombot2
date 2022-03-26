package raven

import (
	"errors"
	"fmt"
	"log"
	"reflect"
	"strconv"

	ravendb "github.com/ravendb/ravendb-go-client"
)

// Database is a struct that contains the database information
type Database struct {
	Host         string
	Port         int
	DatabaseName string
	Dstore       *ravendb.DocumentStore
	Session      *ravendb.DocumentSession
}

// User is a struct that contains the user information
type User struct {
	UID      string  `json:"UID"`
	Username string  `json:"Username"`
	Stocks   int     `json:"Stocks"`
	Bal      float64 `json:"Bal"`
	Rank     int     `json:"Rank"`
	PFP      string  `json:"PFP"`
}

func getDocumentStore(url string, port int, databaseName string) (*ravendb.DocumentStore, error) {
	serverNodes := []string{url + ":" + strconv.Itoa(port)}
	store := ravendb.NewDocumentStore(serverNodes, databaseName)
	if err := store.Initialize(); err != nil {
		return nil, err
	}
	return store, nil
}

// OpenSession opens a session
func OpenSession(url string, port int, databaseName string) (Database, error) {
	store, err := getDocumentStore(url, port, databaseName)
	if err != nil {
		return Database{}, fmt.Errorf("getDocumentStore() failed with %s", err)
	}

	session, err := store.OpenSession("Users")
	if err != nil {
		return Database{}, fmt.Errorf("store.OpenSession() failed with %s", err)
	}
	return Database{
		Host:         url,
		Port:         port,
		DatabaseName: databaseName,
		Dstore:       store,
		Session:      session,
	}, nil
}

// Get gets a user
func (db *Database) Get(uid string) (User, error) {
	var user *User
	tp := reflect.TypeOf(&User{})
	q := db.Session.QueryCollectionForType(tp)
	q = q.WhereEquals("UID", uid)
	err := q.First(&user)
	if err != nil {
		return User{}, err
	}
	dupe := *user
	return dupe, nil

}

// Set sets a user
func (db *Database) Set(user User) error {
	if db.DoesExist(user.UID) == false {
		// User doesn't exist, create it
		err := db.Session.Store(&user)
		if err != nil {
			return err
		}
		err = db.Session.SaveChanges()
		return err
	}
	return errors.New("User already exists")

}

// Update updates a user
func (db *Database) Update(user User) error {
	// HACK: This works but its a really bad way to do it
	if db.DoesExist(user.UID) {
		// get the document ID of the usr
		var usr *User
		tp := reflect.TypeOf(&User{})
		q := db.Session.QueryCollectionForType(tp)
		q = q.WhereEquals("UID", user.UID)
		err := q.First(&usr)
		if err != nil {
			return err
		}
		usr.Bal = user.Bal
		usr.Rank = user.Rank
		usr.Stocks = user.Stocks
		usr.Username = user.Username
		usr.PFP = user.PFP
		err = db.Session.Store(usr)
		if err != nil {
			return err
		}
		err = db.Session.SaveChanges()
		if err != nil {
			return err
		}
		return nil

	}

	return errors.New("User doesn't exist")
}

// DoesExist checks if a user exists
func (db *Database) DoesExist(uid string) bool {
	var user *User
	tp := reflect.TypeOf(&User{})
	q := db.Session.QueryCollectionForType(tp)
	q = q.WhereEquals("UID", uid)
	err := q.First(&user)
	if err != nil {
		return false
	}
	if user == nil {
		return false
	}
	return true
}

// GetAll does what it says on the tin
func (db *Database) GetAll() []User {
	var users []*User
	tp := reflect.TypeOf(&User{})
	q := db.Session.QueryCollectionForType(tp)
	err := q.GetResults(&users)
	if err != nil {
		log.Println(fmt.Errorf("store.OpenSession() failed with %s", err))
	}
	var dupe []User
	for _, user := range users {
		dupe = append(dupe, *user)
	}
	return dupe
}

// Close closes the session
func (db *Database) Close() {
	db.Session.Close()
}
