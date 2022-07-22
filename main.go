package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"

	cmd "github.com/ReCore-sys/bottombot2/commands"
	"github.com/ReCore-sys/bottombot2/libs/config"
	mongo "github.com/ReCore-sys/bottombot2/libs/database"
	"github.com/ReCore-sys/bottombot2/libs/image"
	"github.com/ReCore-sys/bottombot2/libs/logging"
	"github.com/ReCore-sys/bottombot2/libs/stocks"
	"github.com/ReCore-sys/bottombot2/libs/utils"
	"github.com/bwmarrin/discordgo"
	"github.com/lus/dgc"
)

func main() {

	CFG := config.Config()
	println("Initializing Image Library")
	image.Initialize()
	discord, err := discordgo.New("Bot " + string(CFG.Token)) // Create a new Discord session using the provided bot token.
	if err != nil {
		logging.Log(err)
	}

	cmd.SetSession(discord) // Set the session's session handler.
	Router := dgc.Create(&dgc.Router{
		Prefixes: []string{CFG.Prefix},
	}) // Create a new router with the prefixes defined in the config.

	Router = cmd.Registercommands(Router) // Register all commands in the commands.go file.

	Router.Initialize(discord) // Initialize the router with the discord session.
	if !mongo.IsUp() {
		log.Println("\nCan't connect to DB.\nDid you actually start it?")
	} else {

		go stocks.PriceLoop(discord)
		for _, ticker := range mongo.Tickers {
			stocks.UpdatePrice(ticker, stocks.GeneratePrice(ticker))

		}
		stocks.UpdatePricesFile(stocks.Prices)
	}
	err = discord.Open() // Open the connection to Discord.
	if err != nil {
		logging.Log(err)
	}
	go utils.LoopStatus(discord)

	fmt.Println("Bot started!") // Print a message to the console to let the user know the bot is online.
	if !mongo.IsUp() {
		// Print the warning in red
		fmt.Printf("\x1b[31m\n")
		fmt.Println("Database functionality disabled.")

		fmt.Printf("\x1b[0m\n")
	}
	// Wait here until CTRL-C or other term signal is received.
	defer discord.Close() // Close the connection to Discord.

	stop := make(chan os.Signal, 1)        // Create a channel to receive the signal.
	signal.Notify(stop, os.Interrupt)      // Notify the channel when a signal is received.
	<-stop                                 // Wait for the signal.
	logging.LogString("Graceful shutdown") // Print a message to the console to let the user know the bot is shutting down.

}
