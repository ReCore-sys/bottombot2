package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"

	cmd "github.com/ReCore-sys/bottombot2/commands"
	"github.com/ReCore-sys/bottombot2/libs/config"
	raven "github.com/ReCore-sys/bottombot2/libs/database"
	"github.com/ReCore-sys/bottombot2/libs/image"
	"github.com/ReCore-sys/bottombot2/libs/utils"
	"github.com/bwmarrin/discordgo"
	"github.com/lus/dgc"
)

func main() {
	CFG := config.Config()
	println("Initializing Image Library")
	image.Initialize()
	db, err := raven.OpenSession(CFG.Ravenhost, CFG.Ravenport, "users")
	if err != nil {
		log.Println(err)
	}
	users := db.GetAll()
	utils.UpdatePrice(utils.Pricecalc(users))
	discord, err := discordgo.New("Bot " + string(CFG.Token)) // Create a new Discord session using the provided bot token.
	cmd.SesSession(discord)                                   // Set the session's session handler.
	Router := dgc.Create(&dgc.Router{
		Prefixes: []string{CFG.Prefix},
	}) // Create a new router with the prefixes defined in the config.
	Router = cmd.Registercommands(Router) // Register all commands in the commands.go file.

	Router.Initialize(discord) // Initialize the router with the discord session.
	_, err = http.Get(CFG.Ravenhost + ":" + strconv.Itoa(CFG.Ravenport))
	if err != nil {
		print("\n")
		log.Fatal("\nCan't connect to DB.\nDid you actually start it?")
	}
	err = discord.Open() // Open the connection to Discord.
	if err != nil {
		log.Println(err)
	}
	fmt.Println("Bot started!") // Print a message to the console to let the user know the bot is online.
	// Wait here until CTRL-C or other term signal is received.
	defer discord.Close() // Close the connection to Discord.

	stop := make(chan os.Signal, 1)   // Create a channel to receive the signal.
	signal.Notify(stop, os.Interrupt) // Notify the channel when a signal is received.
	<-stop                            // Wait for the signal.
	log.Println("Graceful shutdown")  // Print a message to the console to let the user know the bot is shutting down.

}
