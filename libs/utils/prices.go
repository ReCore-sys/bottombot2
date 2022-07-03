package utils

import (
	"fmt"
	"math/rand"

	raven "github.com/ReCore-sys/bottombot2/libs/database"
)

// RandomChoiceUsers returns a random user from a slice of users.
func RandomChoiceUsers(inp []raven.User) raven.User {
	return inp[rand.Intn(len(inp))]
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
