package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"strconv"
	"time"

	"math/rand"

	"github.com/ReCore-sys/bottombot2/libs/config"
	raven "github.com/ReCore-sys/bottombot2/libs/database"
	"github.com/ReCore-sys/bottombot2/libs/image"
	img "github.com/ReCore-sys/bottombot2/libs/image"
	"github.com/ReCore-sys/bottombot2/libs/utils"
	"github.com/bwmarrin/discordgo"
	"github.com/lus/dgc"
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

	// Read static/ranks.json and parse it into a rank struct
	ranksFile, err := ioutil.ReadFile("static/ranks.json") // Read the ranks file
	if err != nil {
		log.Fatal(err)
	}
	err = json.Unmarshal(ranksFile, &Ranks) // Parse the ranks file into the rank struct
	if err != nil {
		log.Println(err)
	}
	CFG := config.Config()

	router.RegisterCmd(&dgc.Command{
		Name:        "account",
		Description: "Shows usage information about the user",
		Usage:       "account",
		Aliases:     []string{"acc", "bal", "balance", "me"},
		Handler: func(ctx *dgc.Ctx) {
			db, err := raven.OpenSession(CFG.Ravenhost, CFG.Ravenport, "users") // Create a RavenDB session
			if err != nil {
				log.Println(err)
			}
			args := ParseArgs(ctx)
			var target string
			if len(args) == 0 {
				target = ctx.Event.Message.Author.ID
			} else {
				target = ParsePing(args[0])
			}
			if db.DoesExist(target) { // Check if the user already has an account
				acc := img.Account(target)
				file, err := os.Open(acc)
				if err != nil {
					log.Println(err)
				}
				ms := &discordgo.MessageSend{
					Files: []*discordgo.File{
						{
							Name:   target + ".jpeg",
							Reader: file,
						},
					},
				}
				s.ChannelMessageSendComplex(ctx.Event.ChannelID, ms)
				go func() {
					err = file.Close()
					if err != nil {
						println(1)
						log.Println(err)
					}
				}()
				go func() {
					err = os.Remove(acc)
					if err != nil {
						println(2)
						log.Println(err)
					}
				}()
				currentuser, err := db.Get(target)
				if err != nil {
					log.Println(err)
				}
				if ("https://cdn.discordapp.com/avatars/" + ctx.Event.Author.ID + "/" + ctx.Event.Message.Author.Avatar + ".png") != currentuser.PFP {
					currentuser.PFP = ("https://cdn.discordapp.com/avatars/" + ctx.Event.Author.ID + "/" + ctx.Event.Message.Author.Avatar + ".png")
					err = db.Update(currentuser)
					if err != nil {
						log.Println(err)
					}
					image.DownloadFile(currentuser.PFP, acc)
				}
				file = nil
			} else { // Create a new account
				if target == ctx.Event.Author.ID {

					usr := raven.User{
						UID:      ctx.Event.Author.ID,       // Set the user's ID
						Username: ctx.Event.Author.Username, // Set the user's username
						Stocks:   0,
						Bal:      100,
						Rank:     0,
						PFP:      "https://cdn.discordapp.com/avatars/" + ctx.Event.Author.ID + "/" + ctx.Event.Message.Author.Avatar + ".png",
					}
					err = db.Set(usr)
					if err != nil {
						log.Println(err)
					}
					ctx.RespondText("Account created!")
				} else {
					ctx.RespondText("This user either doesn't exist or they don't have an account!")
				}
			}
			if target == ctx.Event.Author.ID {
				go func() {
					user, err := db.Get(target)
					if err != nil {
						log.Println(err)
					}
					user.PFP = "https://cdn.discordapp.com/avatars/" + ctx.Event.Author.ID + "/" + ctx.Event.Message.Author.Avatar + ".png"
					db.Update(user)
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
			ctx.RespondEmbed(&discordgo.MessageEmbed{
				Title:       "Ranks",
				Description: "Here are the ranks!",
				Fields:      fields,
			})
		}})
	println("Registered command: ranks")

	router.RegisterCmd(&dgc.Command{
		Name:        "gamble",
		Description: "1/5 chance of winning 3 times your bet. You Game?",
		Usage:       "gamble <amount>",
		Aliases:     []string{"bet"},
		Handler: func(ctx *dgc.Ctx) {
			if ratecheck(ctx) {
				db, err := raven.OpenSession(CFG.Ravenhost, CFG.Ravenport, "users") // Create a RavenDB session
				GambleResponsesLose := []string{"Damn bro, chill", "LMAO your bank account", "Uhhh you ok there buddy?", "I think you might have a problem", "Maybe you shouldn't gamble that much..."}
				GambleResponsesWin := []string{"Nice. But you won't win every time...", "Make sure it doesn't get out of hand", "Nice work"}
				args := ParseArgs(ctx)
				if len(args) == 0 || args[0] == "" {
					ctx.RespondText("Please specify an amount to gamble.")
					return
				}
				if (regexp.MustCompile(`\d+$`).MatchString(args[0])) == false {
					ctx.RespondText("Please specify a valid amount.")
					return
				}
				amntint, err := strconv.Atoi(args[0])
				amnt := float64(amntint)
				if err != nil {
					log.Println(err)
					ctx.RespondText("Please specify a valid amount.")
					return
				}
				if amnt < 10 {
					ctx.RespondText("You must gamble at least $10.")
					return
				}
				if err != nil {
					log.Println(err)
				}
				if db.DoesExist(ctx.Event.Author.ID) == false { // Check if the user already has an account
					ctx.RespondText("You don't have an account!")
					return
				}
				user, err := db.Get(ctx.Event.Author.ID)
				if err != nil {
					log.Println(err)
				}
				if user.Bal < amnt {
					ctx.RespondText("You don't have enough money!")
					return
				}
				rand.Seed(time.Now().UnixNano())
				if rand.Intn(5) == 0 {
					user.Bal += amnt * 3
					ctx.RespondText("You won! You now have $" + utils.FormatPrice(user.Bal) + " (You won $" + fmt.Sprint(amnt*3) + ")\n " + utils.RandomChoiceStrings(GambleResponsesWin))
				} else {
					user.Bal -= amnt
					ctx.RespondText("You lost! You now have $" + utils.FormatPrice(user.Bal) + " (You lost $" + fmt.Sprint(amnt) + ")\n" + utils.RandomChoiceStrings(GambleResponsesLose))

				}
				err = db.Update(user)
				if err != nil {
					log.Println(err)
				}
				db.Close()
			} else {
				ctx.RespondText("You're being rate limited!")
			}
		}})
	println("Registered command: gamble")

	router.RegisterCmd(&dgc.Command{
		Name:        "stocks",
		Description: "Buy and sell stocks",
		Usage:       "stocks <buy/sell> <amount>",
		Aliases:     []string{"stock", "stonks"},
		Handler: func(ctx *dgc.Ctx) {
			db, err := raven.OpenSession(CFG.Ravenhost, CFG.Ravenport, "users") // Create a RavenDB session
			if err != nil {
				log.Println(err)
			}
			args := ParseArgs(ctx)
			users := db.GetAll()
			if len(args) == 0 {
				price := utils.Price
				ctx.RespondText(fmt.Sprintf("The current price of a stock is $%.2f", price))
			} else if args[0] == "buy" {
				/**=======================
				 *     If the user buys
				 *========================**/
				if regexp.MustCompile(`^\d+$`).MatchString(args[1]) == false {
					ctx.RespondText("Please specify a valid amount.")
					return
				}
				if len(args) == 1 {
					ctx.RespondText("Please specify an amount to buy.")
					return
				}
				amnt, err := strconv.Atoi(args[1])
				if err != nil {
					log.Println(err)
					ctx.RespondText("Please specify a valid amount.")
					return
				}
				if amnt < 1 {
					ctx.RespondText("You must buy at least 1 stock.")
					return
				}
				if db.DoesExist(ctx.Event.Author.ID) == false { // Check if the user already has an account
					ctx.RespondText("You don't have an account!")
					return
				}
				user, err := db.Get(ctx.Event.Author.ID)
				if err != nil {
					log.Println(err)
				}

				var newprice float64
				var userindex int
				userindex = utils.IndexOfUsers(user, users)
				newprice = utils.Price
				/**---------------------------*
				 *    ITERATION LETS GOOOOOOOO
				 *----------------------------**/
				for i := 0; i < amnt; i++ {
					if user.Bal < newprice {
						ctx.RespondText("You don't have enough money to buy " + fmt.Sprint(amnt) + " stocks!")
						return
					}
					user.Bal -= utils.Price
					user.Stocks++
					users[userindex] = user
					newprice = utils.Pricecalc(users)
				}
				ctx.RespondText("You now have " + fmt.Sprint(user.Stocks) + " stocks and $" + utils.FormatPrice(user.Bal) + " left.")
				err = db.Update(user)
				if err != nil {
					log.Println(err)
				}
				utils.UpdatePrice(newprice)
			} else if args[0] == "sell" {
				if len(args) == 1 {
					ctx.RespondText("Please specify an amount to sell.")
					return
				}
				if regexp.MustCompile(`^\d+$`).MatchString(args[1]) == false {
					ctx.RespondText("Please specify a valid amount.")
					return
				}
				amnt, err := strconv.Atoi(args[1])
				if err != nil {
					log.Println(err)
					ctx.RespondText("Please specify a valid amount.")
					return
				}
				if amnt < 1 {
					ctx.RespondText("You must sell at least 1 stock.")
					return
				}
				if db.DoesExist(ctx.Event.Author.ID) == false { // Check if the user already has an account
					ctx.RespondText("You don't have an account!")
					return
				}
				user, err := db.Get(ctx.Event.Author.ID)
				if err != nil {
					log.Println(err)
				}
				if user.Stocks < amnt {
					ctx.RespondText("You don't have enough stocks to sell.")
					return
				}
				var newprice float64
				var userindex int
				userindex = utils.IndexOfUsers(user, users)
				newprice = utils.Price
				/**---------------------------*
				 *    ITERATION LETS GOOOOOOOO
				 *----------------------------**/
				for i := 0; i < amnt; i++ {
					user.Bal += newprice
					user.Stocks--
					users[userindex] = user
					newprice = utils.Pricecalc(users)
				}
				db.Update(user)
				utils.UpdatePrice(newprice)
				ctx.RespondText("You now have " + fmt.Sprint(user.Stocks) + " stocks and $" + utils.FormatPrice(user.Bal) + " left.")
			} else {
				ctx.RespondText("Please specify a valid action.")
			}
			db.Close()
		},
	})
	println("Registered command: stocks")
	return router
}
