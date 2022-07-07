package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	crand "crypto/rand"
	"encoding/hex"
	"encoding/json"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"reflect"
	"regexp"
	"runtime"
	"strconv"
	"testing"
	"time"

	"github.com/ReCore-sys/bottombot2/libs/config"
	raven "github.com/ReCore-sys/bottombot2/libs/database"
	"github.com/ReCore-sys/bottombot2/libs/stocks"
	"github.com/ReCore-sys/bottombot2/libs/utils"
)

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
var filename = "./testfile.txt"

func randSeq(n int) string { // Generate random string
	rand.Seed(time.Now().UnixNano())
	b := make([]rune, n) // Allocate memory
	for i := range b {   // Fill with random letters
		b[i] = letters[rand.Intn(len(letters))] // Generate random letter
	}
	return string(b) // Return string
}
func createHash(key string) string {
	hasher := md5.New()
	hasher.Write([]byte(key))
	return hex.EncodeToString(hasher.Sum(nil))
}

func encrypt(data []byte, passphrase string) []byte {
	block, _ := aes.NewCipher([]byte(createHash(passphrase)))
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		panic(err.Error())
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(crand.Reader, nonce); err != nil {
		panic(err.Error())
	}
	ciphertext := gcm.Seal(nonce, nonce, data, nil)
	return ciphertext
}

func decrypt(data []byte, passphrase string) []byte {
	key := []byte(createHash(passphrase))
	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err.Error())
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		panic(err.Error())
	}
	nonceSize := gcm.NonceSize()
	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		panic(err.Error())
	}
	return plaintext
}

func TestRaven(t *testing.T) {
	CFG := config.Config()
	_, err := raven.OpenSession(CFG.Ravenhost, CFG.Ravenport, "users") // Open database session
	if err != nil {                                                    // Check if session is opened
		t.Errorf("Unable to open database session: %s", err)
		return

	}
}

func TestConnection(t *testing.T) {
	// Check connection to Discord bot gateway, google and my RavenDB server
	CFG := config.Config()
	ips := []string{
		"https://www.google.com/",
		"https://gateway.discord.gg/",
		CFG.Ravenhost + ":" + strconv.Itoa(CFG.Ravenport),
	}
	for _, ip := range ips {
		_, err := http.Get(ip) // Get the website
		if err != nil {
			t.Errorf("Wifi test for %s failed: %s", ip, err)
		}
	}
}
func TestWrite(t *testing.T) {
	// Write string to file
	err := ioutil.WriteFile(filename, []byte(randSeq(64)), 0644)
	if err != nil {
		t.Errorf("File write failed: %s", err)
	}
}

func TestRead(t *testing.T) {
	// check we can read the file
	_, err := ioutil.ReadFile(filename)
	if err != nil {
		t.Errorf("File read failed: %s", err)
	}
}

func TestLargeWrite(t *testing.T) {
	// Write 10 MB of data to file
	err := ioutil.WriteFile(filename, []byte(randSeq(1e+7)), 0644)
	if err != nil {
		t.Errorf("File write failed: %s", err)
	}
}
func TestDel(t *testing.T) {
	// Delete file
	err := os.Remove(filename)
	if err != nil {
		t.Errorf("File delete failed: %s", err)
	}
}

func TestValidUsers(t *testing.T) {
	CFG := config.Config()
	// Check all users have a valid account
	db, err := raven.OpenSession(CFG.Ravenhost, CFG.Ravenport, "users") // Open database session
	if err != nil {
		t.Errorf("Unable to open database session: %s", err)
	}
	allusrs := db.GetAll() // Get an array of all users
	for _, usr := range allusrs {
		if usr.Username == "" { // Check if username is empty
			t.Errorf("User %s has no username", usr.UID)
		}
		if reflect.TypeOf(usr.Username) != reflect.TypeOf("") { // Check if username is a string
			t.Errorf("User %s has invalid username", usr.UID)
		}
		if usr.Bal < 0 { // Check if balance is negative
			t.Errorf("User %s has negative balance", usr.UID)

			for _, ticker := range raven.Tickers {
				if usr.Stocks[ticker] < 0 { // Check if stocks is negative
					t.Errorf("User %s has negative stock count", usr.UID)
				}
			}
		}
		r, err := regexp.Compile(`\d+`) // Check if username contains numbers
		if err != nil {
			t.Errorf("Regex compile failed: %s", err)
		}
		mat := r.MatchString(usr.UID)
		if mat == false { // If it doesn't fail the test
			t.Errorf("User %s has invalid ID", usr.UID)
		}
	}
}

func TestVersion(t *testing.T) {
	v := runtime.Version()
	// Get all characters in the version string that are either a digit or a dot
	r, err := regexp.Compile(`[0-9]+\.[0-9]+`)
	if err != nil {
		t.Errorf("Regex compile failed: %s", err)
	}
	foundstring := r.FindString(v)
	mat, err := strconv.ParseFloat(foundstring, 64)
	if err != nil {
		t.Errorf("Version string is not a float: %s", err)
	}
	if mat < 1.16 {
		t.Errorf("Bottombot2 requires Go 1.16 or higher. You are running %s", v)
	}
}

func TestRand(t *testing.T) {
	rand.Seed(time.Now().UnixNano())
	first := rand.Intn(100000)
	second := rand.Intn(100000)
	if first == second {
		t.Errorf("Random numbers are the same: %d", first)
	}
}

func TestCrypto(t *testing.T) {
	// Test encryption and decryption
	key := []byte("bottombot2")
	data := "bottombot2"
	ciphertext := encrypt(key, data)
	plaintext := decrypt([]byte(ciphertext), string(key))
	if string(plaintext) != data {
		t.Errorf("Encryption failed: %s", string(plaintext))
	}
}

func TestNumformatting(t *testing.T) {
	num1 := 12.6
	num2 := 12.00
	num3 := 120000.00
	if utils.FormatPrice(num1) != "12.60" {
		t.Errorf("Number formatting failed: %s", utils.FormatPrice(num1))
	}
	if utils.FormatPrice(num2) != "12" {
		t.Errorf("Number formatting failed: %s", utils.FormatPrice(num2))
	}
	if utils.FormatPrice(num3) != "120,000" {
		t.Errorf("Number formatting failed: %s", utils.FormatPrice(num3))
	}
}

func TestGraph(t *testing.T) {
	f, err := os.ReadFile("static/prices.json")
	if err != nil {
		t.Errorf("File read failed: %s", err)
	}
	var prices []map[string]float64
	err = json.Unmarshal(f, &prices)
	if err != nil {
		t.Errorf("JSON unmarshal failed: %s", err)
	}
	stocks.CreateGraph(prices)
}
