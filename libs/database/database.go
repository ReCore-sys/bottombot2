package mongo

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/ReCore-sys/bottombot2/libs/config"
	"github.com/ReCore-sys/bottombot2/libs/logging"
	"github.com/lus/dgc"
	"go.mongodb.org/mongo-driver/bson"
	mongodb "go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Database is a struct that contains the database information
type Database struct {
	Host       string          // The IP address of the database
	Port       int             // The port of the database
	Collection string          // The name of the database
	Client     *mongodb.Client // The client of the database
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

var CFG = config.Config()
var Tickers = []string{"ANR", "GST", "ANL", "BKDR"}

func IsUp() bool {
	_, conn := OpenSession(CFG.Server, CFG.Port, CFG.Database)
	return conn != nil
}

// OpenSession opens a session
func OpenSession(url string, port int, collectionName string) (Database, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	options := options.ClientOptions{}
	options.ApplyURI(fmt.Sprintf("mongodb://%s:%d", url, port))
	client, err := mongodb.Connect(ctx, &options)
	if err != nil {
		return Database{}, err
	}
	return Database{
		Host:       url,
		Port:       port,
		Collection: collectionName,
		Client:     client,
	}, client.Ping(ctx, nil)
}

// Get gets a user
func (db *Database) Get(uid string) (User, error) {
	collection := db.Client.Database(CFG.Database).Collection(db.Collection)
	var user User
	err := collection.FindOne(context.TODO(), bson.M{"UID": uid}).Decode(&user)
	if err != nil {
		return User{}, err
	}
	return user, nil
}

// Set sets a user
func (db *Database) Set(user User) error {
	if db.DoesExist(user.UID) {
		return errors.New("User already exists")
	}
	collection := db.Client.Database(CFG.Database).Collection(db.Collection)
	_, err := collection.InsertOne(context.TODO(), user)
	if err != nil {
		return err
	}
	return nil

}

// Update updates a user
func (db *Database) Update(user User) error {
	collection := db.Client.Database(CFG.Database).Collection(db.Collection)
	_, err := collection.UpdateOne(context.TODO(), bson.M{"UID": user.UID}, user)
	if err != nil {
		return err
	}
	return nil
}

// DoesExist checks if a user exists
func (db *Database) DoesExist(uid string) bool {
	collection := db.Client.Database(CFG.Database).Collection(db.Collection)
	var user User
	err := collection.FindOne(context.TODO(), bson.M{"UID": uid}).Decode(&user)
	return err == nil
}

// GetAll does what it says on the tin
func (db *Database) GetAll() []User {
	collection := db.Client.Database(CFG.Database).Collection(db.Collection)
	var users []User
	cur, err := collection.Find(context.TODO(), bson.M{})
	if err != nil {
		logging.Log(err)
		return nil
	}
	for cur.Next(context.TODO()) {
		var user User
		err := cur.Decode(&user)
		if err != nil {
			logging.Log(err)
			return nil
		}
		users = append(users, user)
	}
	return users
}

// Close closes the session
func (db *Database) Close() {
	err := db.Client.Disconnect(context.TODO())
	if err != nil {
		logging.Log(err)
	}
}
