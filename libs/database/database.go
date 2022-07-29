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

type Stats struct {
	User   string  `json:"User"`   // User is the user's ID
	Type   string  `json:"Type"`   // Type is the type of stat. Example: "bal" or "stk"
	Amount float64 `json:"Amount"` // Amount is the amount of money/stocks/other data
	Method string  `json:"Method"` // Method is the method of how the stat was gotten, example: "stocks", "daily", "gamble"
	Time   int64   `json:"Time"`   // Time is the time the stat was created
	Data2  any     `json:"Data2"`  // Data2 is arbitrary data that can be used for whatever is needed
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

// SendStocks() takes a map of strings and floats, converts it to JSON, and sends it to the server
//
// @param stocks map[string]float64
func SendStocks(stocks map[string]float64) {
	client := Client()
	stocks["change"] = float64(ChangeTime.Unix())
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

func GetStocks() map[string]float64 {
	client := Client()
	resp, err := client.Get(fmt.Sprintf("https://%s:%d/api/v1/stocks", CFG.Server, CFG.Port))
	if err != nil {
		logging.Log(err)
		return nil
	}
	defer resp.Body.Close()
	var stocks = make(map[string]float64)
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logging.Log(err)
		return nil
	}
	err = json.Unmarshal(data, &stocks)
	if err != nil {
		logging.Log(err)
		return nil
	}
	return stocks
}

// SendStat() takes a Stats struct, converts it to JSON, sends it to the server, and returns an error
// if there is one
//
// @param Stats stat The stats object that you want to send.
//
// @return error The response from the server.
func SendStat(stat Stats) error {
	client := Client()
	json, err := json.Marshal(stat)
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("https://%s:%d/api/v1/stats", CFG.Server, CFG.Port), bytes.NewBuffer(json))
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
	return nil
}

// Just a simple wrapper so I can send stats when chaining onto the struct
func (stat Stats) Send() error {
	return SendStat(stat)
}
