package raven

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"

	"github.com/ReCore-sys/bottombot2/libs/logging"
	"github.com/lus/dgc"
	ravendb "github.com/ravendb/ravendb-go-client"
)

// Database is a struct that contains the database information
type Database struct {
	Host         string                   // The IP address of the database
	Port         int                      // The port of the database
	DatabaseName string                   // The name of the database
	Dstore       *ravendb.DocumentStore   // The document store
	Session      *ravendb.DocumentSession // The session
}

// User is a struct that contains the user information
type User struct {
	UID      string         `json:"UID"`      // Unique ID, created by discord
	Username string         `json:"Username"` // User's username. No identifiers
	Stocks   map[string]int `json:"Stocks"`   // User's stocks
	Bal      float64        `json:"Bal"`      // User's balance
	Rank     int            `json:"Rank"`     // User's rank
	PFP      string         `json:"PFP"`      // URL to user's profile picture
	Inv      []string       `json:"Inv"`      // User's inventory as ID's of items
	Equipped []string       `json:"Equipped"` // User's equipped items as ID's of items
}

// Item is a struct that holds info about an item
type Item struct {
	ID          string                                                                   `json:"ID"`          // ID is the ID of the item
	Name        string                                                                   `json:"Name"`        // Name is the name of the item
	Description []string                                                                 `json:"Description"` // Description is a description of the item
	Price       float64                                                                  `json:"Price"`       // Price is the price of the item
	Defend      func(attacker User, defender User, chance float64, ctx *dgc.Ctx) float64 // The function to execute from a player when they get robbed
	Attack      func(attacker User, defender User, chance float64, ctx *dgc.Ctx) float64 // The function to execute from a player when they attack someone else
	Image       string                                                                   `json:"Image"` // Image is the path to the item's image
}

var Tickers = []string{"ANR", "GST", "ANL", "BKDR"}

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
	if !db.DoesExist(user.UID) {
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
		logging.Log(fmt.Errorf("store.OpenSession() failed with %s", err))
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
