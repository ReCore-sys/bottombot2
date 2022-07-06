package utils

import (
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"os"
	"regexp"
	"strings"
	"time"

	raven "github.com/ReCore-sys/bottombot2/libs/database"
	"github.com/ReCore-sys/bottombot2/libs/logging"
	"github.com/bwmarrin/discordgo"
	"github.com/lus/dgc"
)

// IndexOfUsers returns the index of a user in a slice of users.
func IndexOfUsers(element raven.User, data []raven.User) int {
	for k, v := range data {
		if element.UID == v.UID {
			return k
		}
	}
	return -1 //not found.
}

// RandomChoiceStrings returns a random string from a slice of strings.
func RandomChoiceStrings(inp []string) string {
	rand.Seed(time.Now().UnixNano())
	return inp[rand.Intn(len(inp))]
}

// FormatPrice formats a float64 to a string with commas.
func FormatPrice(value float64) string {
	// Formate to 2 decimal places if it is not a whole number.
	// If it is a whole number, include no decimal places.
	// Also include comma seperators in large numbers if needed
	var formatted string
	if math.Mod(value, 1) != 0 {
		formatted = fmt.Sprintf("%.2f", value)
	} else {
		formatted = fmt.Sprintf("%.0f", value)
	}
	if value >= 1000 {
		formatted = comma(value)
	}
	return formatted

}

func comma(value float64) string {
	// Add comma seperators to large numbers every 3 digits.
	// For example, 1234 -> 1,234
	// 10000000 -> 10,000,000
	var formatted string
	formatted = fmt.Sprintf("%.0f", value)
	for i := len(formatted) - 3; i > 0; i -= 3 {
		formatted = formatted[:i] + "," + formatted[i:]
	}
	return formatted
}

// SendFile sends a file to a channel.
func SendFile(file *os.File, ctx *dgc.Ctx, s *discordgo.Session) {
	ms := &discordgo.MessageSend{
		Files: []*discordgo.File{
			{
				Name:   file.Name(),
				Reader: file,
			},
		},
	}
	_, err := s.ChannelMessageSendComplex(ctx.Event.ChannelID, ms)
	if err != nil {
		logging.Log(err)
	}
}

var DoneIntervals int64

func IntervalCheck(intervalms int) bool {
	// Check if the current time is within the interval.
	// If it is, return true.
	// If it is not, return false.
	now := time.Now().UnixMilli()
	if now%int64(intervalms) == 0 {
		if now != DoneIntervals {
			DoneIntervals = now
			return true
		}
	}
	return false
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

func IsIn(item string, itemlist []string) bool {
	for _, v := range itemlist {
		if item == v {
			return true
		}
	}
	return false
}

func LoopStatus(discord *discordgo.Session) {
	for {
		f, err := os.ReadFile("static/status.json")
		if err != nil {
			logging.Log(err)
			continue
		}
		var status [][2]string
		err = json.Unmarshal(f, &status)
		if err != nil {
			logging.Log(err)
			continue
		}
		var choice [2]string
		rand.Seed(time.Now().UnixNano())
		choice = status[rand.Intn(len(status))]
		switch choice[0] {
		case "l":
			err := discord.UpdateListeningStatus(choice[1])
			if err != nil {
				logging.Log(err)
			}
			println("Status updated: Listening to " + choice[1])
		case "p":
			err := discord.UpdateGameStatus(0, choice[1])
			if err != nil {
				logging.Log(err)
			}
			println("Status updated: Playing " + choice[1])
		case "s":
			err := discord.UpdateStreamingStatus(0, choice[1], "https://www.twitch.tv/lolgetcucked")
			if err != nil {
				logging.Log(err)
			}
			println("Status updated: Streaming " + choice[1])
		}

		time.Sleep(time.Minute * 30)

	}
}

func DoesFileExist(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	return false
}
