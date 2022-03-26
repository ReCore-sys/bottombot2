package utils

import (
	"fmt"
	"math"
	"math/rand"
	"time"

	raven "github.com/ReCore-sys/bottombot2/libs/database"
)

// IndexOfUsers returns the index of a user in a slice of users.
func IndexOfUsers(element raven.User, data []raven.User) int {
	for k, v := range data {
		if element.UID == v.UID {
			return k
		}
	}
	return -1 //not found.
}

// RandomChoiceStrings returns a random string from a slice of strings.
func RandomChoiceStrings(inp []string) string {
	rand.Seed(time.Now().UnixNano())
	return inp[rand.Intn(len(inp))]
}

// FormatPrice formats a float64 to a string with commas.
func FormatPrice(value float64) string {
	// Formate to 2 decimal places if it is not a whole number.
	// If it is a whole number, include no decimal places.
	// Also include comma seperators in large numbers if needed
	var formatted string
	if math.Mod(value, 1) != 0 {
		formatted = fmt.Sprintf("%.2f", value)
	} else {
		formatted = fmt.Sprintf("%.0f", value)
	}
	if value >= 1000 {
		formatted = comma(value)
	}
	return formatted

}

func comma(value float64) string {
	// Add comma seperators to large numbers every 3 digits.
	// For example, 1234 -> 1,234
	// 10000000 -> 10,000,000
	var formatted string
	formatted = fmt.Sprintf("%.0f", value)
	for i := len(formatted) - 3; i > 0; i -= 3 {
		formatted = formatted[:i] + "," + formatted[i:]
	}
	return formatted
}
