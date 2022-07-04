package utils

import (
	"math/rand"

	raven "github.com/ReCore-sys/bottombot2/libs/database"
)

// RandomChoiceUsers returns a random user from a slice of users.
func RandomChoiceUsers(inp []raven.User) raven.User {
	return inp[rand.Intn(len(inp))]
}
