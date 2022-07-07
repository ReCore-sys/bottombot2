// Stocks is such a huge part, it needs it's own file smh
package stocks

import (
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/ReCore-sys/bottombot2/libs/config"
	raven "github.com/ReCore-sys/bottombot2/libs/database"
	"github.com/ReCore-sys/bottombot2/libs/logging"
	"github.com/ReCore-sys/bottombot2/libs/utils"
	"github.com/bwmarrin/discordgo"
	"github.com/hako/durafmt"
	"github.com/lus/dgc"
	"gopkg.in/yaml.v2"
)

var UntilChange time.Time
var volatility = make(map[string]float64)

// It registers a command that allows users to buy and sell stocks
//
// router: The router that the command is being registered to
func RegisterStocks(router *dgc.Router) *dgc.Router {

	file, err := os.ReadFile("volatility.yaml")
	if err != nil {
		logging.Log(err)
	}
	err = yaml.Unmarshal(file, &volatility)
	if err != nil {
		logging.Log(err)
	}

	for k := range Prices {
		Prices[k] = 50
	}

	router.RegisterCmd(&dgc.Command{
		Name:        "stocks",
		Description: "Buy and sell stocks",
		Usage:       "stocks <buy/sell> <ticker> <amount>",
		Aliases:     []string{"stock", "stonks"},
		Handler: func(ctx *dgc.Ctx) {
			CFG := config.Config()
			db, err := raven.OpenSession(CFG.Ravenhost, CFG.Ravenport, "users") // Create a RavenDB session
			if err != nil {
				logging.Log(err)
			}

			defer db.Close()
			var ticker string
			args := utils.ParseArgs(ctx)
			//users := db.GetAll()
			if len(args) == 0 {
				/**============================================
				 *               Stock price
				 *=============================================**/
				response := "**Current stock prices:**\n\n"
				for k := range Prices {
					response += fmt.Sprintf("%s: $%.2f\n", k, Prices[k])
				}
				response += fmt.Sprintf("\nThe price will change in %v.", durafmt.Parse(time.Until(UntilChange)).LimitFirstN(2))
				err = ctx.RespondText(response)
				if err != nil {
					logging.Log(err)
				}
			} else if args[0] == "buy" || args[0] == "sell" {
				if len(args) < 3 {
					err = ctx.RespondText("Usage: stocks <buy/sell> <ticker> <amount>")
					if err != nil {
						logging.Log(err)
					}
					return
				}
				ticker = strings.ToUpper(args[1])
				if !db.DoesExist(ctx.Event.Author.ID) {
					err = ctx.RespondText("You don't have an account!")
					if err != nil {
						logging.Log(err)
					}
					return
				}
				/**============================================
				 *               Stock buy/sell
				 *=============================================**/
				if len(args) == 1 {
					err = ctx.RespondText("Please specify an amount.")
					if err != nil {
						logging.Log(err)
					}
					return
				}
				if len(args) == 1 {
					err = ctx.RespondText("Please specify a ticker.")
					if err != nil {
						logging.Log(err)
					}
					return
				} else if !utils.IsIn(ticker, raven.Tickers) {
					err = ctx.RespondText("That ticker is not valid.")
					if err != nil {
						logging.Log(err)
					}
					return
				}
				var amntint int
				switch args[0] {
				case "buy":
					user, err := db.Get(ctx.Event.Author.ID)
					if err != nil {
						logging.Log(err)
					}
					if !(regexp.MustCompile(`\d+$`).MatchString(args[2])) {
						if args[2] == "all" {
							amntint = int(math.Floor(user.Bal / Prices[ticker]))
						} else {
							err = ctx.RespondText("Please specify a valid amount.")
							if err != nil {
								logging.Log(err)
							}
							return
						}
					} else {
						amntint, err = strconv.Atoi(args[2])
					}
					if err != nil {
						logging.Log(err)
					}
					if amntint < 1 {
						err = ctx.RespondText("You must buy at least 1 stock.")
						if err != nil {
							logging.Log(err)
						}
						return
					}
					if user.Bal < float64(amntint)*Prices[ticker] {
						err = ctx.RespondText("You don't have enough money!")
						if err != nil {
							logging.Log(err)
						}
						return
					}

					user.Bal -= float64(amntint) * Prices[ticker]
					user.Stocks[ticker] += amntint
					err = ctx.RespondText("You now have $" + utils.FormatPrice(user.Bal) + " and " + fmt.Sprint(user.Stocks[ticker]) + " (" + ticker + ")" + " stocks.")
					if err != nil {
						logging.Log(err)
					}
					err = db.Update(user)
					if err != nil {
						logging.Log(err)
					}
					db.Close()
				case "sell":
					user, err := db.Get(ctx.Event.Author.ID)
					if err != nil {
						logging.Log(err)
					}
					if !(regexp.MustCompile(`\d+$`).MatchString(fmt.Sprint(args[2]))) {
						if args[2] == "all" {
							amntint = user.Stocks[ticker]
						} else {
							err = ctx.RespondText("Please specify a valid amount.")
							if err != nil {
								logging.Log(err)
							}
							return
						}
					} else {
						amntint, err = strconv.Atoi(args[2])
					}
					if err != nil {
						logging.Log(err)
					}
					if amntint < 1 {
						err = ctx.RespondText("You must sell at least 1 stock.")
						if err != nil {
							logging.Log(err)
						}
						return
					}
					if user.Stocks[ticker] < amntint {
						err = ctx.RespondText("You don't have that many stocks!")
						if err != nil {
							logging.Log(err)
						}
						return
					}
					user.Bal += float64(amntint) * Prices[ticker]
					user.Stocks[ticker] -= amntint
					err = ctx.RespondText("You now have $" + utils.FormatPrice(user.Bal) + " and " + fmt.Sprint(user.Stocks[ticker]) + " stocks from " + ticker + ".")
					if err != nil {
						logging.Log(err)
					}
					err = db.Update(user)
					if err != nil {
						logging.Log(err)
					}

				}

			} else if args[0] == "amount" {
				user, err := db.Get(ctx.Event.Author.ID)
				if err != nil {
					logging.Log(err)
				}
				var stocks string
				for ticker, amnt := range user.Stocks {
					stocks += ticker + ": " + fmt.Sprint(amnt) + "\n"
				}
				err = ctx.RespondText("You have:\n" + stocks)
				if err != nil {
					logging.Log(err)
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

	UntilChange = time.Now().Add(20 * time.Minute)
	for {
		if utils.IntervalCheck(20 * 1000 * 60) {

			UntilChange = time.Now().Add(20 * time.Minute)
			for _, ticker := range raven.Tickers {
				Prices[ticker] = math.Round(GeneratePrice(ticker)*100) / 100
			}

		}

	}
}

// Prices The global price struct. Start with a price of 50.
var Prices = make(map[string]float64)

// UpdatePrice updates the global price variable.
func UpdatePrice(ticker string, price float64) {
	price = math.Round(price*100) / 100

	println(fmt.Sprintf("Price updated to: $%.2f", price))
	Prices[ticker] = price
}

func GeneratePrice(ticker string) float64 {
	rand.Seed(time.Now().UnixNano())
	if Prices[ticker] == 0 {
		Prices[ticker] = 50
	}
	volpercent := volatility[ticker]
	volatility := float64(volpercent / 100.0)
	rnd := rand.Float64()
	changePercent := 2 * volatility * rnd
	if changePercent > volatility {
		changePercent -= (2 * volatility)
	}
	changeAmount := float64(Prices[ticker]) * changePercent
	newPrice := float64(Prices[ticker]) + changeAmount
	return newPrice
}

func UpdatePricesFile(prices map[string]float64) {
	f, err := os.OpenFile("static/prices.json", os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		logging.Log(err)
	}
	var arrayofprices []map[string]float64
	data, err := os.ReadFile("static/prices.json")
	if err != nil {
		logging.Log(err)
	}
	err = json.Unmarshal(data, &arrayofprices)
	if err != nil {
		logging.Log(err)
	}
	arrayofprices = append(arrayofprices, prices)
	if len(arrayofprices) > 360 {
		arrayofprices = arrayofprices[len(arrayofprices)-360:]
	}
	err = json.NewEncoder(f).Encode(arrayofprices)
	if err != nil {
		logging.Log(err)
	}
	err = f.Close()
	if err != nil {
		logging.Log(err)
	}
}
