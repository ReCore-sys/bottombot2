// Stocks is such a huge part, it needs it's own file smh
package stocks

import (
	"fmt"
	"log"
	"math"
	"math/rand"
	"regexp"
	"strconv"
	"time"

	"github.com/ReCore-sys/bottombot2/libs/config"
	raven "github.com/ReCore-sys/bottombot2/libs/database"
	"github.com/ReCore-sys/bottombot2/libs/utils"
	"github.com/bwmarrin/discordgo"
	"github.com/lus/dgc"
)

var UntilChange time.Duration

// It registers a command that allows users to buy and sell stocks
//
// router: The router that the command is being registered to
func RegisterStocks(router *dgc.Router) *dgc.Router {
	router.RegisterCmd(&dgc.Command{
		Name:        "stocks",
		Description: "Buy and sell stocks",
		Usage:       "stocks <buy/sell> <amount>",
		Aliases:     []string{"stock", "stonks"},
		Handler: func(ctx *dgc.Ctx) {
			CFG := config.Config()
			db, err := raven.OpenSession(CFG.Ravenhost, CFG.Ravenport, "users") // Create a RavenDB session
			if err != nil {
				log.Println(err)
			}

			defer db.Close()
			args := utils.ParseArgs(ctx)
			//users := db.GetAll()
			price := Price
			if len(args) == 0 {
				/**============================================
				 *               Stock price
				 *=============================================**/
				response := fmt.Sprintf("The current stock price is $%v\n", price)
				response += fmt.Sprintf("The price will change in %v.\n", UntilChange)
				err = ctx.RespondText(response)
				if err != nil {
					log.Println(err)
				}
			} else if args[0] == "buy" || args[0] == "sell" {
				if !db.DoesExist(ctx.Event.Author.ID) {
					err = ctx.RespondText("You don't have an account!")
					if err != nil {
						log.Println(err)
					}
					return
				}
				/**============================================
				 *               Stock buy/sell
				 *=============================================**/
				if len(args) == 1 {
					err = ctx.RespondText("Please specify an amount.")
					if err != nil {
						log.Println(err)
					}
					return
				}

				switch args[0] {
				case "buy":
					user, err := db.Get(ctx.Event.Author.ID)
					if err != nil {
						log.Println(err)
					}
					if !(regexp.MustCompile(`\d+$`).MatchString(args[1])) {
						if args[1] == "all" {
							args[1] = fmt.Sprintf("%f", math.Floor(user.Bal/price))
						} else {
							err = ctx.RespondText("Please specify a valid amount.")
							if err != nil {
								log.Println(err)
							}
							return
						}
					}
					amntint, err := strconv.Atoi(args[1])
					if err != nil {
						log.Println(err)
					}
					if amntint < 1 {
						err = ctx.RespondText("You must buy at least 1 stock.")
						if err != nil {
							log.Println(err)
						}
						return
					}
					if user.Bal < float64(amntint)*price {
						err = ctx.RespondText("You don't have enough money!")
						if err != nil {
							log.Println(err)
						}
						return
					}

					user.Bal -= float64(amntint) * price
					user.Stocks += amntint
					err = ctx.RespondText("You now have $" + utils.FormatPrice(user.Bal) + " and " + fmt.Sprint(user.Stocks) + " stocks.")
					if err != nil {
						log.Println(err)
					}
					err = db.Update(user)
					if err != nil {
						log.Println(err)
					}
					db.Close()
				case "sell":
					user, err := db.Get(ctx.Event.Author.ID)
					if err != nil {
						log.Println(err)
					}
					if !(regexp.MustCompile(`\d+$`).MatchString(args[1])) {
						if args[1] == "all" {
							args[1] = fmt.Sprintf("%d", user.Stocks)
						} else {
							err = ctx.RespondText("Please specify a valid amount.")
							if err != nil {
								log.Println(err)
							}
							return
						}
					}
					amntint, err := strconv.Atoi(args[1])
					if err != nil {
						log.Println(err)
					}
					if amntint < 1 {
						err = ctx.RespondText("You must sell at least 1 stock.")
						if err != nil {
							log.Println(err)
						}
						return
					}
					if user.Stocks < amntint {
						err = ctx.RespondText("You don't have that many stocks!")
						if err != nil {
							log.Println(err)
						}
						return
					}
					user.Bal += float64(amntint) * price
					user.Stocks -= amntint
					err = ctx.RespondText("You now have $" + utils.FormatPrice(user.Bal) + " and " + fmt.Sprint(user.Stocks) + " stocks.")
					if err != nil {
						log.Println(err)
					}
					err = db.Update(user)
					if err != nil {
						log.Println(err)
					}

				}
			}
		},
	})
	println("Registered command: stocks")
	return router
}

// It updates the bot's status every 20 minutes with the current stock price
//
// @param discord *discordgo.Session
func PriceLoop(discord *discordgo.Session) {
	for {
		if utils.IntervalCheck(20 * 1000 * 60) {
			UntilChange = time.Until(time.Now().Add(20 * time.Minute))
			Price = math.Round(GeneratePrice()*100) / 100
			err := discord.UpdateGameStatus(0, fmt.Sprintf("with $%v in stocks", Price))
			if err != nil {
				log.Println(err)
			}
		}
	}
}

// Price The global price struct. Start with a price of 50.
var Price = 50.0

// UpdatePrice updates the global price variable.
func UpdatePrice(price float64) {
	println(fmt.Sprintf("Price updated to: $%.2f", price))
	price = math.Round(price*100) / 100
	Price = price
}

func GeneratePrice() float64 {
	volpercent := 4.0
	volatility := float64(volpercent / 100.0)
	rnd := rand.Float64()
	changePercent := 2 * volatility * rnd
	if changePercent > volatility {
		changePercent -= (2 * volatility)
	}
	changeAmount := Price * changePercent
	newPrice := Price + changeAmount
	return newPrice
}
