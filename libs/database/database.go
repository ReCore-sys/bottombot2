package db

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/ReCore-sys/bottombot2/libs/config"
	"github.com/ReCore-sys/bottombot2/libs/logging"
	"github.com/lus/dgc"
)

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
var ChangeTime time.Time

func Client() http.Client {
	return http.Client{
		Timeout: time.Second * 1,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}
}

func IsUp() bool {
	CFG := config.Config()
	client := Client()
	_, err := client.Get(fmt.Sprintf("https://%s:%d/api/v1/ping", CFG.Server, CFG.Port))
	if err != nil {
		println(err.Error())
	}
	return err == nil
}

// Get gets a user
func Get(uid string) (User, error) {
	client := Client()
	resp, err := client.Get(fmt.Sprintf("https://%s:%d/api/v1/user/%s", CFG.Server, CFG.Port, uid))
	if err != nil {
		return User{}, err
	}
	defer resp.Body.Close()
	var user User
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return User{}, err
	}
	err = json.Unmarshal(data, &user)
	if err != nil {
		return User{}, err
	}
	return user, nil
}

// Set sets a user
func Set(user User) error {
	client := Client()
	json, err := json.Marshal(user)
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("https://%s:%d/api/v1/user", CFG.Server, CFG.Port), bytes.NewBuffer(json))
	if err != nil {
		return err
	}
	req.Header.Set("Auth-Key", CFG.Apipass)
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil

}

// Update updates a user
func Update(user User) error {
	client := Client()
	json, err := json.Marshal(user)
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPatch, fmt.Sprintf("https://%s:%d/api/v1/user", CFG.Server, CFG.Port), bytes.NewBuffer(json))
	if err != nil {
		return err
	}
	req.Header.Set("Auth-Key", CFG.Apipass)
	req.Header.Set("Content-type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	code := resp.StatusCode
	_ = code
	return err
}

// DoesExist checks if a user exists
func DoesExist(uid string) bool {
	client := Client()
	resp, err := client.Get(fmt.Sprintf("https://%s:%d/api/v1/exist/%s", CFG.Server, CFG.Port, uid))
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return false
	}
	return string(data) == "true"
}

// GetAll does what it says on the tin
func GetAll() []User {
	var users []User
	client := Client()
	resp, err := client.Get(fmt.Sprintf("https://%s:%d/api/v1/users", CFG.Server, CFG.Port))
	if err != nil {
		logging.Log(err)
		return users
	}
	defer resp.Body.Close()
	err = json.NewDecoder(resp.Body).Decode(&users)
	if err != nil {
		logging.Log(err)
		return users
	}
	return users
}

func SendStocks(stocks map[string]float64) {
	client := Client()
	stocks["change"] = float64(ChangeTime.Unix())
	fmt.Printf("%+v\n", stocks)
	json, err := json.Marshal(stocks)
	if err != nil {
		logging.Log(err)
		return
	}
	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("https://%s:%d/api/v1/stocks", CFG.Server, CFG.Port), bytes.NewBuffer(json))
	if err != nil {
		logging.Log(err)
		return
	}
	req.Header.Set("Auth-Key", CFG.Apipass)
	req.Header.Set("Content-type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		logging.Log(err)
		return
	}
	defer resp.Body.Close()
}
