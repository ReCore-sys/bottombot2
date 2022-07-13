package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strconv"
	"time"

	"math/rand"

	"github.com/ReCore-sys/bottombot2/libs/config"
	mongo "github.com/ReCore-sys/bottombot2/libs/database"
	"github.com/ReCore-sys/bottombot2/libs/image"
	img "github.com/ReCore-sys/bottombot2/libs/image"
	"github.com/ReCore-sys/bottombot2/libs/logging"
	"github.com/ReCore-sys/bottombot2/libs/stocks"
	"github.com/ReCore-sys/bottombot2/libs/utils"
	"github.com/bwmarrin/discordgo"
	"github.com/hako/durafmt"
	"github.com/lus/dgc"
	"gopkg.in/yaml.v2"
)

type rank struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Price int    `json:"price"`
	Color []int  `json:"color"`
}

// Ranks is the global rank struct.
var Ranks []rank
var rateLimits = make(map[string]map[string]time.Time)
var dailys = make(map[string]time.Time)

func ratecheck(ctx *dgc.Ctx) bool {
	// When the command is run, log the user's ID and the command name in the map wth the time they sent it.
	// If less than 5 seconds have passed, return false.
	// If more than 5 seconds have passed, delete the entry from the map and return true.
	if rateLimits[ctx.Command.Name] == nil {
		rateLimits[ctx.Command.Name] = map[string]time.Time{}
	}
	if time.Since(rateLimits[ctx.Command.Name][ctx.Event.Author.ID]) < 5*time.Second {
		return false
	}
	if time.Since(rateLimits[ctx.Command.Name][ctx.Event.Author.ID]) > 5*time.Second {
		delete(rateLimits[ctx.Command.Name], ctx.Event.Author.ID)
	}

	rateLimits[ctx.Command.Name][ctx.Event.Author.ID] = time.Now()

	return true
}

// EcoRoute is the router for the economy/account commands.
func EcoRoute(router *dgc.Router) *dgc.Router {
	if utils.DoesFileExist("./static/dailys.json") {
		f, err := os.ReadFile("./static/dailys.json")
		if err != nil {
			logging.Log(err)
		}
		err = json.Unmarshal(f, &dailys)
		if err != nil {
			logging.Log(err)
		}
	}

	// Read static/ranks.yaml and parse it into a rank struct
	ranksFile, err := ioutil.ReadFile("static/ranks.yaml") // Read the ranks file
	if err != nil {

		logging.Log(err)
	}
	err = yaml.Unmarshal(ranksFile, &Ranks) // Parse the ranks file into the rank struct
	if err != nil {
		println(1)
		logging.Log(err)
	}
	CFG := config.Config()

	router.RegisterCmd(&dgc.Command{
		Name:        "account",
		Description: "Shows usage information about the user",
		Usage:       "account",
		Aliases:     []string{"acc", "bal", "balance", "me"},
		Handler: func(ctx *dgc.Ctx) {
			db, err := mongo.OpenSession(CFG.Server, CFG.Port, CFG.Collection) // Create a RavenDB session
			if err != nil {
				logging.Log(err)
			}
			args := utils.ParseArgs(ctx)
			var target string
			if len(args) == 0 {
				target = ctx.Event.Message.Author.ID
			} else {
				target = utils.ParsePing(args[0])
			}
			if db.DoesExist(target) { // Check if the user already has an account
				acc := img.Account(target)
				file, err := os.Open(acc)
				if err != nil {
					logging.Log(err)
				}
				ms := &discordgo.MessageSend{
					Files: []*discordgo.File{
						{
							Name:   target + ".jpeg",
							Reader: file,
						},
					},
				}
				_, err = s.ChannelMessageSendComplex(ctx.Event.ChannelID, ms)
				if err != nil {
					logging.Log(err)
				}
				go func() {
					err = file.Close()
					if err != nil {
						logging.Log(err)
					}
				}()
				go func() {
					err = os.Remove(acc)
					if err != nil {
						logging.Log(err)
					}
				}()
				currentuser, err := db.Get(target)
				if err != nil {
					logging.Log(err)
				}
				if ("https://cdn.discordapp.com/avatars/" + ctx.Event.Author.ID + "/" + ctx.Event.Message.Author.Avatar + ".png") != currentuser.PFP {
					currentuser.PFP = ("https://cdn.discordapp.com/avatars/" + ctx.Event.Author.ID + "/" + ctx.Event.Message.Author.Avatar + ".png")
					err = db.Update(currentuser)
					if err != nil {
						logging.Log(err)
					}
					err = image.DownloadFile(currentuser.PFP, acc)
					if err != nil {
						logging.Log(err)
					}
				}
				file = nil
			} else { // Create a new account
				if target == ctx.Event.Author.ID {

					usr := mongo.User{
						UID:      ctx.Event.Author.ID,       // Set the user's ID
						Username: ctx.Event.Author.Username, // Set the user's username
						Bal:      100,
						Rank:     0,
						PFP:      "https://cdn.discordapp.com/avatars/" + ctx.Event.Author.ID + "/" + ctx.Event.Message.Author.Avatar + ".png",
					}
					usr.Stocks = make(map[string]int)
					for _, ticker := range mongo.Tickers {
						usr.Stocks[ticker] = 0
					}
					err = db.Set(usr)
					if err != nil {
						logging.Log(err)
					}
					err = ctx.RespondText("Account created!")
					if err != nil {
						logging.Log(err)
					}
				} else {
					err = ctx.RespondText("This user either doesn't exist or they don't have an account!")
					if err != nil {
						logging.Log(err)
					}
				}
			}
			if target == ctx.Event.Author.ID {
				go func() {
					user, err := db.Get(target)
					if err != nil {
						logging.Log(err)
					}
					user.PFP = "https://cdn.discordapp.com/avatars/" + ctx.Event.Author.ID + "/" + ctx.Event.Message.Author.Avatar + ".png"
					err = db.Update(user)
					if err != nil {
						logging.Log(err)
					}
				}()
			}
			db.Close()
		},
	})
	println("Registered command: me")

	router.RegisterCmd(&dgc.Command{
		Name:        "ranks",
		Description: "Shows the ranks",
		Usage:       "ranks",
		Aliases:     []string{"rank"},
		Handler: func(ctx *dgc.Ctx) {
			var fields []*discordgo.MessageEmbedField
			for _, rank := range Ranks {
				if rank.ID > 0 {
					fields = append(fields, &discordgo.MessageEmbedField{
						Name:   rank.Name,
						Value:  "$" + fmt.Sprint(rank.Price) + "\n(ID: " + fmt.Sprint(rank.ID) + ")",
						Inline: true},
					)
				}
			}
			err = ctx.RespondEmbed(&discordgo.MessageEmbed{
				Title:       "Ranks",
				Description: "Here are the ranks!",
				Fields:      fields,
			})
			if err != nil {
				logging.Log(err)
			}
		}})
	println("Registered command: ranks")

	router.RegisterCmd(&dgc.Command{
		Name:        "gamble",
		Description: "1/5 chance of winning 3 times your bet. You Game?",
		Usage:       "gamble <amount>",
		Aliases:     []string{"bet"},
		Handler: func(ctx *dgc.Ctx) {
			if ratecheck(ctx) {
				db, err := mongo.OpenSession(CFG.Server, CFG.Port, CFG.Collection) // Create a RavenDB session
				if err != nil {
					logging.Log(err)
				}
				GambleResponsesLose := []string{"Damn bro, chill", "LMAO your bank account", "Uhhh you ok there buddy?", "I think you might have a problem", "Maybe you shouldn't gamble that much..."}
				GambleResponsesWin := []string{"Nice. But you won't win every time...", "Make sure it doesn't get out of hand", "Nice work"}
				args := utils.ParseArgs(ctx)
				if len(args) == 0 || args[0] == "" {
					err = ctx.RespondText("Please specify an amount to gamble.")
					if err != nil {
						logging.Log(err)
					}
					return
				}
				if !(regexp.MustCompile(`\d+$`).MatchString(args[0])) {
					err = ctx.RespondText("Please specify a valid amount.")
					if err != nil {
						logging.Log(err)
					}
					return
				}
				amntint, err := strconv.Atoi(args[0])
				amnt := float64(amntint)
				if err != nil {
					logging.Log(err)
					err = ctx.RespondText("Please specify a valid amount.")
					if err != nil {
						logging.Log(err)
					}
					return
				}
				if amnt < 10 {
					err = ctx.RespondText("You must gamble at least $10.")
					if err != nil {
						logging.Log(err)
					}
					return
				}
				if err != nil {
					logging.Log(err)
				}
				if !db.DoesExist(ctx.Event.Author.ID) { // Check if the user already has an account
					err = ctx.RespondText("You don't have an account!")
					if err != nil {
						logging.Log(err)
					}
					return
				}
				user, err := db.Get(ctx.Event.Author.ID)
				if err != nil {
					logging.Log(err)
				}
				if user.Bal < amnt {
					err = ctx.RespondText("You don't have enough money!")
					if err != nil {
						logging.Log(err)
					}
					return
				}
				rand.Seed(time.Now().UnixNano())
				if rand.Intn(5) == 0 {
					user.Bal += amnt * 3
					err = ctx.RespondText("You won! You now have $" + utils.FormatPrice(user.Bal) + " (You won $" + fmt.Sprint(amnt*3) + ")\n " + utils.RandomChoiceStrings(GambleResponsesWin))
					if err != nil {
						logging.Log(err)
					}
				} else {
					user.Bal -= amnt
					err = ctx.RespondText("You lost! You now have $" + utils.FormatPrice(user.Bal) + " (You lost $" + fmt.Sprint(amnt) + ")\n" + utils.RandomChoiceStrings(GambleResponsesLose))

					if err != nil {
						logging.Log(err)
					}
				}
				err = db.Update(user)
				if err != nil {
					logging.Log(err)
				}
				db.Close()
			} else {
				err = ctx.RespondText("You're being rate limited!")
				if err != nil {
					logging.Log(err)
				}
			}
		}})
	println("Registered command: gamble")

	router = stocks.RegisterStocks(router)

	router.RegisterCmd(&dgc.Command{
		Name:        "daily",
		Aliases:     []string{},
		Description: "grants a daily bonus",
		Usage:       "daily",
		Handler: func(ctx *dgc.Ctx) {
			db, err := mongo.OpenSession(CFG.Server, CFG.Port, CFG.Collection) // Create a RavenDB session
			if err != nil {
				logging.Log(err)
			}
			if !db.DoesExist(ctx.Event.Author.ID) { // Check if the user already has an account
				err = ctx.RespondText("You don't have an account!")
				if err != nil {
					logging.Log(err)
				}
				return
			}
			user, err := db.Get(ctx.Event.Author.ID)
			if err != nil {
				logging.Log(err)
			}
			if _, ok := dailys[ctx.Event.Author.ID]; ok {
				if dailys[ctx.Event.Author.ID].Unix()+(24*60*60) < time.Now().Unix() {

					dailys[ctx.Event.Author.ID] = time.Now()
					user.Bal += 100
					err = ctx.RespondText("You received $" + utils.FormatPrice(100) + " for your daily!")
					if err != nil {
						logging.Log(err)
					}
					err = db.Update(user)
					if err != nil {
						logging.Log(err)
					}
					dailys[ctx.Event.Author.ID] = time.Now()
					db.Close()
				} else {
					err = ctx.RespondText(fmt.Sprintf("You have already claimed your daily! Try again in %s", durafmt.Parse(time.Until(dailys[ctx.Event.Author.ID].Add(time.Hour*24))).LimitFirstN(2)))
					if err != nil {
						logging.Log(err)
					}
					return
				}
			} else {
				dailys[ctx.Event.Author.ID] = time.Now()
				user.Bal += 100
				err = ctx.RespondText("You received $" + utils.FormatPrice(100) + " for your daily!")
				if err != nil {
					logging.Log(err)
				}
				err = db.Update(user)
				if err != nil {
					logging.Log(err)
				}
				db.Close()

				dailys[ctx.Event.Author.ID] = time.Now()
			}
			f, err := os.OpenFile("static/dailys.json", os.O_RDWR|os.O_CREATE, 0644)
			if err != nil {
				logging.Log(err)
			}
			defer f.Close()
			err = json.NewEncoder(f).Encode(dailys)
			if err != nil {
				logging.Log(err)
			}

		}})

	router.RegisterCmd(&dgc.Command{
		Name:        "prices",
		Aliases:     []string{},
		Description: "Shows the previous prices of the stocks",
		Usage:       "prices",
		Handler: func(ctx *dgc.Ctx) {
			data, err := os.ReadFile("static/prices.json")
			if err != nil {
				logging.Log(err)
			}
			var prices []map[string]float64
			err = json.Unmarshal(data, &prices)
			if err != nil {
				logging.Log(err)
			}
			stocks.CreateGraph(prices)
			file, err := os.Open("graph.png")
			if err != nil {
				logging.Log(err)
			}
			ms := &discordgo.MessageSend{
				Files: []*discordgo.File{
					{
						Name:   "price.jpeg",
						Reader: file,
					},
				},
			}
			_, err = s.ChannelMessageSendComplex(ctx.Event.ChannelID, ms)
			if err != nil {
				logging.Log(err)
			}
		},
	})

	return router
}
