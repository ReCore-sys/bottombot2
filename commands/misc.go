package cmd

import (
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/lus/dgc"
)

// MiscRoute is a route that handles misc commands
func MiscRoute(router *dgc.Router) *dgc.Router {
	router.RegisterCmd(&dgc.Command{
		Name:        "ping",
		Description: "Responds with 'pong!'",
		Usage:       "ping",
		Example:     "ping",
		IgnoreCase:  true,
		Handler: func(ctx *dgc.Ctx) {
			time := time.Now()
			msgtime := ctx.Event.Message.Timestamp
			timediff := time.Sub(msgtime)
			print(timediff)
			ctx.RespondEmbed(&discordgo.MessageEmbed{
				Title:       "Ping!",
				Description: "This message took " + fmt.Sprint(timediff) + " to send.",
			})
		},
	})
	println("Registered command: ping")

	router.RegisterCmd(&dgc.Command{
		Name:        "info",
		Description: "sends an embed with information about the bot",
		Usage:       "info",
		Example:     "info",
		Handler: func(ctx *dgc.Ctx) {
			ctx.RespondEmbed(&discordgo.MessageEmbed{
				Title:       "Info",
				Description: "This is bottombot, the single most fucked up bot in the world.",
			})
		},
	})
	println("Registered command: info")

	router.RegisterCmd(&dgc.Command{
		Name:        "duck",
		Description: "Sends an image of a duck",
		Usage:       "duck",
		Example:     "duck",
		Handler: func(ctx *dgc.Ctx) {
			ctx.RespondText("https://www.chromethemer.com/download/hd-wallpapers/another-duck-in-the-snow-3840x2160.jpg")
		},
	})
	println("Registered command: duck")
	return router
}
