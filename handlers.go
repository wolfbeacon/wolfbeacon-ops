package main

import (
	slackbot "github.com/wolfbeacon/go-slackbot"
	"github.com/nlopes/slack"
	"golang.org/x/net/context"
)

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

func ListS3BucketsHandler(ctx context.Context, bot *slackbot.Bot, evt *slack.MessageEvent){
	task := ListS3BucketsTask{}
	result := task.Run(make([]string, 0), make([]string, 0))
	if result[1] == "error" {
		bot.Reply(evt, "Failed to list S3 buckets. " + result[2], slackbot.WithTyping)
	} else {
		var reply string
		buckets := result[2:]
		
		for _, name := range buckets {
			reply += name + "\n"
		}
	
		attachment := slack.Attachment{
			Color: "#1565C0",
			Fallback: reply,
			Text: reply,
			Footer: "Data from AWS S3",
		}
	
		attachments := make([]slack.Attachment, 0)
	
		attachments = append(attachments, attachment)
	
		bot.ReplyWithAttachments(evt, attachments, slackbot.WithTyping)
	}
}