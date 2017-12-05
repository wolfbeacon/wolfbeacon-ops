package main

import (
	"encoding/json"
	"fmt"
	slackbot "github.com/BeepBoopHQ/go-slackbot"
	"github.com/nlopes/slack"
	"golang.org/x/net/context"
	"os"
)

type Configuration struct {
	SlackKey string `json:"slack-key"`
}

type User struct {
	Email       string   `json:"email"`
	Permissions []string `json:"permissions"`
}

func (u *User) can(permission string) bool {
	for _, p := range u.Permissions {
		if permission == p {
			return true
		}
	}
	return false
}

// Will return User with no permimssion if not found
func FindUser(users []User, email string) User {
	for _, u := range users {
		if u.Email == email {
			return u
		}
	}
	return User{email, make([]string, 0)}
}

var api *slack.Client
var users []User

func main() {

	// Read configurations
	file, _ := os.Open("./config/settings.json")
	decoder := json.NewDecoder(file)
	config := Configuration{}
	err := decoder.Decode(&config)
	if err != nil {
		fmt.Println("Error reading config file:", err)
	}
	fmt.Println("Slack Key:", config.SlackKey)

	file.Close()

	// Read user configurations

	file, _ = os.Open("./config/users.json")
	decoder = json.NewDecoder(file)
	users = make([]User, 0)
	err = decoder.Decode(&users)
	if err != nil {
		fmt.Println("Error reading users config file:", err)
	}
	fmt.Println(len(users), "user(s) read.")

	file.Close()

	api = slack.New(config.SlackKey)
	bot := slackbot.New(config.SlackKey)
	toMe := bot.Messages(slackbot.DirectMessage, slackbot.DirectMention).Subrouter()

	toMe.Hear("(?i)(hi|hello).*").MessageHandler(HelloHandler)
	toMe.Hear("^help$").MessageHandler(HelpHandler)
	toMe.Hear("^me$").MessageHandler(MeHandler)

	bot.Run()

}

func HelloHandler(ctx context.Context, bot *slackbot.Bot, evt *slack.MessageEvent) {
	bot.Reply(evt, "Hello! If you need help, type `help`.", slackbot.WithTyping)
}

func HelpHandler(ctx context.Context, bot *slackbot.Bot, evt *slack.MessageEvent) {
	msg := `*Here are all the commands you can use:*
	help - Display help message
	list {playbooks | tasks} - List playbook or tasks currently running
	run {playbook} <playbook-name> - Run playbook
	me - Show detailed status about yourself
	`
	bot.Reply(evt, msg, slackbot.WithTyping)
}

func MeHandler(ctx context.Context, bot *slackbot.Bot, evt *slack.MessageEvent) {
	u, err := api.GetUserInfo(evt.User)
	if err != nil {
		fmt.Println("Failed to get user info", evt.User)
	} else {
		msg := "Hi, " + u.Profile.Email + "\n"
		user := FindUser(users, u.Profile.Email)
		if len(user.Permissions) == 0 {
			msg += "Sorry, seems like you don't have any permissions yet..."
		} else {
			msg += "Here are your permissions:\n"
			for _, permission := range user.Permissions {
				msg += "`" + permission + "`\n"
			}
		}
		bot.Reply(evt, msg, slackbot.WithTyping)
	}
}
