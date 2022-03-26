package cmd

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/ReCore-sys/bottombot2/libs/config"
	"github.com/bwmarrin/discordgo"
	"github.com/lus/dgc"
)

var s *discordgo.Session

// SesSession allows you to set the session's session handler.
func SesSession(session *discordgo.Session) {
	s = session
}

// ParseArgs returns a slice of strings that are the arguments of the command.
func ParseArgs(ctx *dgc.Ctx) []string {
	args := ctx.Arguments.Raw()
	res := strings.Split(args, " ")
	if len(res) == 0 {
		return []string{}
	}
	if res[0] == "" {
		return []string{}
	}
	return res
}

// ParsePing gets an ID from a ping
func ParsePing(arg string) string {
	reg := regexp.MustCompile(`^<@!?(\d+)>$`)
	result := reg.FindStringSubmatch(arg)
	if len(result) == 0 {
		return ""
	}
	return result[1]
}

// Registercommands is a general command that recieves a router object then passes it around all the command routers
func Registercommands(router *dgc.Router) *dgc.Router {
	// I am probably having my addressed traced by go devs right now because of
	// how I did this but the IRS still hasn't caught up with me so I'll be fine.
	cfg := config.Config()
	router = MiscRoute(router)
	router = EcoRoute(router)

	router.RegisterCmd(&dgc.Command{ // Add the help command
		Name:        "help",
		Description: "Shows help about the commands",
		Usage:       "help <command>",
		Example:     "help ping",
		Handler: func(ctx *dgc.Ctx) {
			var commands []string            // Create a slice of strings to store the commands in.
			if ctx.Arguments.Amount() == 0 { // If there are no arguments, show all commands.
				for _, cmd := range ctx.Router.Commands {
					formatted := fmt.Sprintf("%s - %s \n*%s%s*\n", cmd.Name, cmd.Description, cfg.Prefix, cmd.Usage) // Format the command for the help command.
					commands = append(commands, formatted)                                                           // Append the formatted command to the slice.
				}
				msg := "Available commands: \n" + strings.Join(commands, "\n")
				ctx.RespondText(msg)
			} else { // If there are arguments, show the help for the specified command.
				for _, cmd := range ctx.Router.Commands { // Loop through all commands.
					if cmd.Name == ctx.Arguments.Get(0).Raw() { // If the command name matches the argument,
						embed := &discordgo.MessageEmbed{ // Create an embed.
							Title:       cmd.Name,
							Description: cmd.Description,
							Fields: []*discordgo.MessageEmbedField{
								{
									Name:   "Usage",
									Value:  cfg.Prefix + cmd.Usage,
									Inline: false,
								},
								{
									Name:   "Example",
									Value:  cfg.Prefix + cmd.Example,
									Inline: false,
								},
							}}
						ctx.RespondEmbed(embed) // Respond with the embed.
						return
					}
				}
				ctx.RespondText("No command found with that name") // If no command was found, respond with this message.
			}
		}})
	print("Registered command: help")

	return router
}
