package cmd

import (
	"fmt"
	"time"

	"github.com/ReCore-sys/bottombot2/libs/logging"
	"github.com/bwmarrin/discordgo"
	"github.com/inhies/go-bytesize"
	"github.com/jaypipes/ghw"
	"github.com/lus/dgc"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/mem"
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
			msgtime := ctx.Event.Message.Timestamp
			timediff := time.Since(msgtime)
			print(timediff)
			err := ctx.RespondEmbed(&discordgo.MessageEmbed{
				Title:       "Ping!",
				Description: "This message took " + fmt.Sprint(timediff) + " to send.",
			})
			if err != nil {
				logging.Log(err)
			}
		},
	})
	println("Registered command: ping")

	router.RegisterCmd(&dgc.Command{
		Name:        "info",
		Description: "sends an embed with information about the bot",
		Usage:       "info",
		Example:     "info",
		Handler: func(ctx *dgc.Ctx) {
			err := ctx.RespondEmbed(&discordgo.MessageEmbed{
				Title:       "Info",
				Description: "This is bottombot, the single most fucked up bot in the world.",
			})
			if err != nil {
				logging.Log(err)
			}
		},
	})
	println("Registered command: info")

	router.RegisterCmd(&dgc.Command{
		Name:        "duck",
		Description: "Sends an image of a duck",
		Usage:       "duck",
		Example:     "duck",
		Handler: func(ctx *dgc.Ctx) {
			err := ctx.RespondText("https://www.chromethemer.com/download/hd-wallpapers/another-duck-in-the-snow-3840x2160.jpg")
			if err != nil {
				logging.Log(err)
			}
		},
	})
	println("Registered command: duck")

	router.RegisterCmd(&dgc.Command{
		Name:        "stats",
		Description: "gets system stats",
		Usage:       "stats",
		Handler: func(ctx *dgc.Ctx) {
			mem, err := mem.VirtualMemory()
			if err != nil {
				logging.Log(err)
			}
			aaaaaaaaa, err := cpu.Info()
			cpu := aaaaaaaaa[0]
			if err != nil {
				logging.Log(err)
			}
			host, err := host.Info()
			if err != nil {
				logging.Log(err)
			}
			gpu, err := ghw.GPU()
			if err != nil {
				logging.Log(err)
			}
			fields := []*discordgo.MessageEmbedField{
				{
					Name:   "OS",
					Value:  fmt.Sprintf("%s %s", host.Platform, host.PlatformVersion),
					Inline: false,
				},
				{
					Name:   "CPU",
					Value:  fmt.Sprintf("Name: %s\n@ %.1f Ghz\nCores: %d\n", cpu.ModelName, cpu.Mhz/1000, cpu.Cores),
					Inline: false,
				},
				{
					Name:   "Memory",
					Value:  fmt.Sprintf("Total: %s\nUsed: %s\nFree: %s\n", bytesize.New(float64(mem.Total)), bytesize.New(float64(mem.Used)), bytesize.New(float64(mem.Free))),
					Inline: false,
				},
				{
					Name:  "GPU",
					Value: gpu.GraphicsCards[0].DeviceInfo.Product.Name,
				},
			}
			err = ctx.RespondEmbed(&discordgo.MessageEmbed{
				Title:  "Stats",
				Type:   discordgo.EmbedTypeRich,
				Fields: fields,
			})
			if err != nil {
				logging.Log(err)
			}
		},
	})
	println("Registered command: stats")
	return router
}
