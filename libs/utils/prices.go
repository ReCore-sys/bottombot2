package utils

import (
	"fmt"
	"math"
	"math/rand"

	raven "github.com/ReCore-sys/bottombot2/libs/database"
)

// Price The global price struct. Start with a price of 50.
var Price = 50.0

// UpdatePrice updates the global price variable.
func UpdatePrice(price float64) {
	Price = price
}

// Pricecalc calculates the price of a stock
func Pricecalc(users []raven.User) float64 {
	price := 50.0         // Start with a price of 50
	ratio := ratio(users) // Calculate the ratio of stocks to raw cash
	price = 50 * ratio    // Multiply 50 by the ratio to get the new price
	return price

}

// RandomChoiceUsers returns a random user from a slice of users.
func RandomChoiceUsers(inp []raven.User) raven.User {
	return inp[rand.Intn(len(inp))]
}

// ratio calculates the ratio of stocks to raw cash
func ratio(inp []raven.User) float64 {
	var totalCash, totalStocks float64
	// Iterate through the users and add their cash and stocks to the total
	for _, v := range inp {
		totalCash += float64(v.Bal)
		totalStocks += float64(v.Stocks)
	}
	if totalStocks == 0 {
		totalStocks = 1 // Prevent divide by zero
	}
	stocksval := (totalStocks * Price) // Calculate the value of the stocks
	return totalCash / stocksval       // Return the ratio of the two
}

// Createusers creates a slice of users with random stats.
func Createusers(amnt int) []raven.User {
	var users []raven.User
	for i := 0; i < amnt; i++ {
		users = append(users, raven.User{
			UID:      fmt.Sprintf("%d", i),
			Username: fmt.Sprintf("user%d", i),
			Stocks:   rand.Intn(100),
			Bal:      rand.Float64() * 100,
			Rank:     0,
			PFP:      "https://i.imgur.com/XqQXQ8l.png",
		},
		)
	}
	return users
}
func main() {
	var prices []float64
	for i := 0; i < 100; i++ {
		users := Createusers(100)
		price := Pricecalc(users)
		prices = append(prices, price)
		random := RandomChoiceUsers(users)
		index := IndexOfUsers(random, users)
		if price < 50 {
			// Buy
			random.Bal += price * float64(random.Stocks)
			random.Stocks = 0
		} else {
			// Sell
			random.Stocks += int(math.Floor(random.Bal / price))
			random.Bal = 0
		}
		users[index] = random
		fmt.Printf("%f\n", price)
	}
}
