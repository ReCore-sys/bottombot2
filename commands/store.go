package cmd

import (
	"fmt"
	"strings"

	raven "github.com/ReCore-sys/bottombot2/libs/database"
	"github.com/ReCore-sys/bottombot2/libs/logging"
	"github.com/bwmarrin/discordgo"
	"github.com/lus/dgc"
)

func StoreRoute(route *dgc.Router) *dgc.Router {

	route.RegisterCmd(&dgc.Command{
		Name:        "store",
		Aliases:     []string{"shop"},
		Description: "Shows the store",
		Usage:       "store",
		Handler: func(ctx *dgc.Ctx) {
			fields := []*discordgo.MessageEmbedField{}
			for _, item := range raven.AllItems {
				if len(item.Description) > 1 {
					fields = append(fields, &discordgo.MessageEmbedField{
						Name:   item.Name + fmt.Sprintf(" ($%v)", item.Price),
						Value:  strings.Join(item.Description, "\n"),
						Inline: false,
					})
				} else {
					fields = append(fields, &discordgo.MessageEmbedField{
						Name:   item.Name,
						Value:  item.Description[0],
						Inline: true,
					})
				}
			}
			err := ctx.RespondEmbed(&discordgo.MessageEmbed{
				Title:  "Store",
				Fields: fields,
			})
			if err != nil {
				logging.Log(err)
			}
		},
	})

	return route
}
