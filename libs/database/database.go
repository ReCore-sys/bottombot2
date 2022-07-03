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
	Host         string                   // The IP address of the database
	Port         int                      // The port of the database
	DatabaseName string                   // The name of the database
	Dstore       *ravendb.DocumentStore   // The document store
	Session      *ravendb.DocumentSession // The session
}

// User is a struct that contains the user information
type User struct {
	UID      string  `json:"UID"`      // Unique ID, created by discord
	Username string  `json:"Username"` // User's username. No identifiers
	Stocks   int     `json:"Stocks"`   // User's stocks
	Bal      float64 `json:"Bal"`      // User's balance
	Rank     int     `json:"Rank"`     // User's rank
	PFP      string  `json:"PFP"`      // URL to user's profile picture
}

// Item is a struct that holds info about an item
type Item struct {
	ID          string     `json:"ID"`          // ID is the ID of the item
	Name        string     `json:"Name"`        // Name is the name of the item
	Type        string     `json:"Type"`        // Type is the type of the item (weapon, armour, etc)
	Description string     `json:"Description"` // Description is a description of the item
	Rarity      [2]float64 `json:"Rarity"`      // Rarity is the rarity of the item
	Damage      [2]float64 `json:"Damage"`      // Damage is the damage of the item.
	OnUse       string     `json:"Onuse"`       // When the item is used, this string is looked up in the map of item use functions and run
	IsActive    bool       `json:"IsActive"`    // IsActive is whether the item is equiped or not
}

// Combat is a struct that holds info regarding combat stuff like health, items, level, etc
type Combat struct {
	Health       float64 `json:"Health"`       // Health is the amount of health the user has
	HealthMax    float64 `json:"HealthMax"`    // HealthMax is the max amount of health the user can have
	Level        int     `json:"Level"`        // Level is the level of the user
	XP           int     `json:"XP"`           // XP is the amount of XP the user has
	Inv          []Item  `json:"Inv"`          // Inv is the inventory of the user. Items are stored as a map of item ID to quantity
	ActiveWeapon Item    `json:"ActiveWeapon"` // ActiveWeapon is the ID of the active weapon
	ActiveArmour []Item  `json:"ActiveArmour"` // ActiveArmour is the IDs of the active armour
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
