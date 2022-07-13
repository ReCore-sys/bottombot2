package mongo

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/ReCore-sys/bottombot2/libs/logging"
	"github.com/lus/dgc"
)

// I can't be bothered creating a weird JSON format or using dependency injection so Imma just hardcode it

var AllItems = []Item{
	{
		ID:          "aceofspades",
		Name:        "Ace of Spades",
		Description: []string{`"Eyes up Guardian"`, "When successfully robbed, there is a 1% chance for the defender to steal 90% of the attacker's assets"},
		Price:       800,
		Defend: func(attacker User, defender User, chance float64, ctx *dgc.Ctx) float64 {
			if chance == -1 {
				return -1 // If it's negative 1, return negative 1. This indicates that nothing should happen
			}
			if chance > 0.5 {
				chance *= 0.9
			}
			rand.Seed(time.Now().UnixNano())
			if rand.Float64() < chance {
				stolenmoney := attacker.Bal * 0.9
				var totalstocks int
				for ticker, amount := range attacker.Stocks {
					defender.Stocks[ticker] += amount
					attacker.Stocks[ticker] = 0
					totalstocks += amount
				}
				defender.Bal += stolenmoney
				attacker.Bal -= stolenmoney
				err := ctx.RespondText(fmt.Sprintf("%s fucked up. %s stole $%v and %v stocks from %s", attacker.Username, defender.Username, stolenmoney, totalstocks, attacker.Username))
				if err != nil {
					logging.Log(err)
				}
			}
			db, err := OpenSession("localhost", 8080, "users")
			if err != nil {
				logging.Log(err)
			}
			err = db.Set(attacker)
			if err != nil {
				logging.Log(err)
			}
			err = db.Set(defender)
			if err != nil {
				logging.Log(err)
			}

			return -1
		},
		Attack: func(attacker User, defender User, chance float64, ctx *dgc.Ctx) float64 {
			return chance
		},
		Image: "aceofspades.png",
	},
	{
		ID:          "acrustysock",
		Name:        "A crusty sock",
		Description: []string{`"Why is it moving..."`, "Reduces the chance of being robbed by 25%"},
		Price:       400,
		Defend: func(attacker User, defender User, chance float64, ctx *dgc.Ctx) float64 {
			if chance == -1 {
				return -1 // If it's negative 1, return negative 1. This indicates that nothing should happen
			}
			chance *= 0.75
			return chance
		},
		Attack: func(attacker User, defender User, chance float64, ctx *dgc.Ctx) float64 {
			return chance
		},
		Image: "crustysock.png",
	},
	{
		ID:          "taxpayersrage",
		Name:        "Taxpayer's Rage",
		Description: []string{`"Where the hell did this extra money come from???`, "Has a chance to grant you the extra money depending on your chances of successfully robbing someone. Chances of getting extra pay is inverse to your robbing chance"},
		Price:       750,
		Defend: func(attacker User, defender User, chance float64, ctx *dgc.Ctx) float64 {
			return chance
		},
		Attack: func(attacker User, defender User, chance float64, ctx *dgc.Ctx) float64 {
			if chance == -1 {
				return -1 // If it's negative 1, return negative 1. This indicates that nothing should happen
			}
			rand.Seed(time.Now().UnixNano())
			if rand.Float64() > chance {
				attacker.Bal += (750 * chance)
				err := ctx.RespondText(fmt.Sprintf("%s got a bonus of $%v from Taxpayer's Rage", attacker.Username, (750 * chance)))
				if err != nil {
					logging.Log(err)
				}
			}
			return chance
		},
		Image: "",
	},
}
