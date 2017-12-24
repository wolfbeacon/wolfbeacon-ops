package main

import (
	"encoding/json"
	"fmt"
	slackbot "github.com/wolfbeacon/go-slackbot"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/codebuild"
	"github.com/aws/aws-sdk-go/service/elasticbeanstalk"
	"github.com/nlopes/slack"
	"golang.org/x/net/context"
	"github.com/robfig/cron"
	"os"
	"github.com/satori/go.uuid"
	"strings"
)

// Will return User with no permimssion if not found
func FindUser(email string) User {
	for _, u := range users {
		if u.Email == email {
			return u
		}
	}
	return User{email, make([]string, 0)}
}

var api *slack.Client
var users []User
var config Configuration
var buildStatuses map[string]string
var envStatuses map[string]string
var bot *slackbot.Bot
var codeBuildBuilds []*codebuild.Build
var elasticBeanstalkEnviroments []*elasticbeanstalk.EnvironmentDescription

var firstBuildStatusLoad bool
var firstEnviromentStatusLoad bool

func main() {

	codeBuildBuilds = make([]*codebuild.Build, 0)
	elasticBeanstalkEnviroments = make([]*elasticbeanstalk.EnvironmentDescription, 0)

	firstBuildStatusLoad = true
	firstEnviromentStatusLoad = true

	// Read configurations
	file, _ := os.Open("./config/settings.json")
	decoder := json.NewDecoder(file)
	config = Configuration{}
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
	bot = slackbot.New(config.SlackKey)
	toMe := bot.Messages(slackbot.Mention, slackbot.DirectMention, slackbot.DirectMessage).Subrouter()
	buildStatuses = make(map[string]string)
	envStatuses = make(map[string]string)

	// Configure cron jobs
	c := cron.New()
	c.AddFunc("*/10 * * * * *", CheckBuildStatusCronJob)
	c.AddFunc("*/10 * * * * *", CheckEnviromentsCronJob)
	c.Start()


	ReloadCodeBuildBuilds()
	ReloadElasticBeanstalkEnviroments()

	toMe.Hear("(?i)(hi|hello).*").MessageHandler(HelloHandler)
	toMe.Hear("^help$").MessageHandler(HelpHandler)
	toMe.Hear("list builds").MessageHandler(ListCodeBuildBuildsHandler)
	toMe.Hear("run build .*").MessageHandler(StartCodeBuildBuildHandler)
	toMe.Hear("list envs").MessageHandler(ListElasticBeanstalkEnviromentsHandler)
	toMe.Hear("rebuild env .*").MessageHandler(RebuildElasticBeanstalkEnviromentHandler)
	toMe.Hear("list projects").MessageHandler(ListCodeBuildProjectsHandler)
	toMe.Hear("list buckets").MessageHandler(ListS3BucketsHandler)

	bot.Run()

}

// list projects handler
func ListCodeBuildProjectsHandler(ctx context.Context, bot *slackbot.Bot, evt *slack.MessageEvent){
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(config.AWSRegion)},
	)
	svc := codebuild.New(sess)
	projects, err := svc.ListProjects(&codebuild.ListProjectsInput{SortOrder: aws.String("ASCENDING")})
	if err != nil {
		bot.Reply(evt, "Failed to list build projects.", slackbot.WithTyping)
	}

	var result string

	for _, project := range projects.Projects {
		result += *project + "\n"
	}

	attachment := slack.Attachment{
		Color: "#1565C0",
		Fallback: result,
		Text: result,
		Footer: "Data from AWS CodeBuild",
	}

	attachments := make([]slack.Attachment, 0)

	attachments = append(attachments, attachment)

	bot.ReplyWithAttachments(evt, attachments, slackbot.WithTyping)
}

// list builds handler, will only list last 5 builds
func ListCodeBuildBuildsHandler(ctx context.Context, bot *slackbot.Bot, evt *slack.MessageEvent) {
	attachments := make([]slack.Attachment, 0)

	for _, build := range codeBuildBuilds[(len(codeBuildBuilds)-5):] {
		attachments = append(attachments, ConstructCodeBuildBuildAttachment(build))
	}
	
	bot.ReplyWithAttachments(evt, attachments, slackbot.WithTyping)
}

func StartCodeBuildBuildHandler(ctx context.Context, bot *slackbot.Bot, evt *slack.MessageEvent) {
	
	// Get user information
	u, _ := api.GetUserInfo(evt.User)
	user := FindUser(u.Profile.Email)

	// If user have permission to start a build
	if user.Can("start-build") {

		evt.Text = slackbot.StripDirectMention(evt.Text)
		args := strings.Split(evt.Text, " ")
		if len(args) > 2 {
			sess, err := session.NewSession(&aws.Config{
				Region: aws.String(config.AWSRegion)},
			)
			svc := codebuild.New(sess)
	
			_, err = svc.StartBuild(&codebuild.StartBuildInput{ProjectName: aws.String(args[2])})
	
			if err != nil {
				bot.Reply(evt, "Failed to start build ..." + err.Error() + "<@" + evt.User + ">", slackbot.WithTyping)
			} else {
				bot.Reply(evt, "Project build started ... <@" + evt.User + ">" , slackbot.WithTyping)
			}
		} else {
			bot.Reply(evt, "You must specify a project. <@" + evt.User + ">", slackbot.WithTyping)
		}

	} else {
		bot.Reply(evt, "You do not have permission to start a new build. <@" + evt.User + ">", slackbot.WithTyping)
	}
}

func ListElasticBeanstalkEnviromentsHandler(ctx context.Context, bot *slackbot.Bot, evt *slack.MessageEvent) {
	attachments := make([]slack.Attachment, 0)
	
	for _, env := range elasticBeanstalkEnviroments {
		attachments = append(attachments, ConstructElasticBeanstalkEnviromentAttachment(env, false))
	}
	
	bot.ReplyWithAttachments(evt, attachments, slackbot.WithTyping)
}

func RebuildElasticBeanstalkEnviromentHandler(ctx context.Context, bot *slackbot.Bot, evt *slack.MessageEvent) {
	// Get user information
	u, _ := api.GetUserInfo(evt.User)
	user := FindUser(u.Profile.Email)

	// If user have permission to start a build
	if user.Can("rebuild-env") {
		evt.Text = slackbot.StripDirectMention(evt.Text)
		args := strings.Split(evt.Text, " ")
		if len(args) > 3 {
			sess, err := session.NewSession(&aws.Config{
				Region: aws.String(config.AWSRegion)},
			)
			v := uuid.NewV4().String()
			svc := elasticbeanstalk.New(sess)
			input := &elasticbeanstalk.CreateApplicationVersionInput{
				ApplicationName:       aws.String(args[2]),
				AutoCreateApplication: aws.Bool(true),
				Process:               aws.Bool(true),
				SourceBundle: &elasticbeanstalk.S3Location{
					S3Bucket: aws.String("wolfbeacon-dockerrun-aws"),
					S3Key:    aws.String(args[2] + "/Dockerrun.aws.json"),
				},
				VersionLabel: aws.String(v),
			}
			
			_, err = svc.CreateApplicationVersion(input)
			if err != nil {
				bot.Reply(evt, "Failed to initialize a rebuild", slackbot.WithTyping)
			} else {
	
				input := &elasticbeanstalk.UpdateEnvironmentInput{
					EnvironmentName: aws.String(args[3]),
					VersionLabel:    aws.String(v),
				}
				
				_, err = svc.UpdateEnvironment(input)
				if err != nil {
					bot.Reply(evt, "Failed to initialize a rebuild " + err.Error() + "<@" + evt.User + ">", slackbot.WithTyping)
				} else {
					bot.Reply(evt, "Now rebuilding enviroment... this can take a few minutes to complete. <@" + evt.User + ">", slackbot.WithTyping)
				}
			}
		} else {
			bot.Reply(evt, "You must specify a enviroment name. <@" + evt.User + ">", slackbot.WithTyping)
		}
	} else {
		bot.Reply(evt, "You do not have permission to rebuild an enviroment. <@" + evt.User + ">", slackbot.WithTyping)
	}
}

func PostToChannelWithAttachments(text string, attachments []slack.Attachment){
	params := slack.PostMessageParameters{AsUser: true}
	params.Attachments = attachments

	bot.Client.PostMessage(config.AnnounceChannel, text, params)
}

// func ReloadCodeBuildBuilds() bool {
// 	sess, err := session.NewSession(&aws.Config{
// 		Region: aws.String(config.AWSRegion)},
// 	)
// 	svc := codebuild.New(sess)
// 	names, err := svc.ListBuilds(&codebuild.ListBuildsInput{SortOrder: aws.String("ASCENDING")})
// 	if err != nil {
// 		return false
// 	}
// 	builds, err := svc.BatchGetBuilds(&codebuild.BatchGetBuildsInput{Ids: names.Ids})
	
// 	if err != nil {
// 		return false
// 	}

// 	codeBuildBuilds = builds.Builds

// 	return true
// }

func ReloadElasticBeanstalkEnviroments() bool {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(config.AWSRegion)},
	)

	ebSvc := elasticbeanstalk.New(sess)

	environments, err := ebSvc.DescribeEnvironments(&elasticbeanstalk.DescribeEnvironmentsInput{})

	if err != nil {
		return false
	}

	elasticBeanstalkEnviroments = environments.Environments
	return true
}

func ConstructCodeBuildBuildAttachment(build *codebuild.Build) slack.Attachment {
	var c string
	var statusText string

	switch aws.StringValue(build.BuildStatus) {
	case "SUCCEEDED":
		c = "#7CD197"
		statusText = "Succeeded ✓"
	case "FAILED":
		c = "#C62828"
		statusText = "Failed ×"
	case "FAULT":
		c = "#C62828"
		statusText = "Fault ×"
	case "TIMED_OUT":
		c = "#C62828"
		statusText = "Timed Out ×"
	default:
		c = "#F9A825"
		statusText = "In Progress"
	}

	result := slack.Attachment{
		Color: c,
		Fallback: aws.StringValue(build.ProjectName),
		Text: aws.StringValue(build.ProjectName) + " | " + statusText,
		Footer: "Started: " + build.StartTime.String() + " | ID: " + aws.StringValue(build.Id),
	}

	return result
}

func ConstructElasticBeanstalkEnviromentAttachment(env *elasticbeanstalk.EnvironmentDescription, showDetail bool) slack.Attachment {
	var c string
	var statusText string

	switch aws.StringValue(env.Health) {
	case "Green":
		c = "#7CD197"
		statusText = "Functional ✓"
	case "Yellow":
		c = "#F9A825"
		statusText = "Warning ×"
	case "Red":
		c = "#C62828"
		statusText = "Fault ×"
	default:
		c = "#9E9E9E"
		statusText = "Unknown"
	}

	if aws.StringValue(env.Status) == "Updating" {
		c = "#0288D1"
		statusText += " Updaing ..."
	}

	if aws.StringValue(env.Status) == "Launching" {
		c = "#0288D1"
		statusText += " Launching ..."
	}
	
	result := slack.Attachment{
		Color: c,
		Fallback: aws.StringValue(env.EnvironmentName),
		Text: aws.StringValue(env.ApplicationName) + " " + aws.StringValue(env.EnvironmentName) + " | " + statusText,
	}

	if showDetail {
		result.Text += "\n" + "Enviroment Status: " + aws.StringValue(env.Status) + "\n"
		result.Text += "Version: " + aws.StringValue(env.VersionLabel)
		result.Footer = "Last update: " + env.DateUpdated.String()
	}

	return result
}

// Cron job to check all CodeBuild jobs.
func CheckBuildStatusCronJob() {

	// Take a copy of old builds
	oldBuilds := codeBuildBuilds

	// Reload jobs
	if !ReloadCodeBuildBuilds() {
		fmt.Println("Failed to reload codebuild build jobs")
	}

	oldBuildFound := false

	for _, newBuild := range codeBuildBuilds {
		for _, oldBuild := range oldBuilds {
			// We are comparing two same builds
			if aws.StringValue(oldBuild.Id) == aws.StringValue(newBuild.Id) {
				oldBuildFound = true
				// If build status has changed
				if aws.StringValue(oldBuild.BuildStatus) != aws.StringValue(newBuild.BuildStatus) {
					attachments := make([]slack.Attachment, 0)
					attachments = append(attachments, ConstructCodeBuildBuildAttachment(newBuild))
					PostToChannelWithAttachments("Build status updated", attachments)
				}
			}
		}
		if !oldBuildFound {
			// We have detected a new build job
			attachments := make([]slack.Attachment, 0)
			attachments = append(attachments, ConstructCodeBuildBuildAttachment(newBuild))
			PostToChannelWithAttachments("New build job started ...", attachments)
		}
		oldBuildFound = false
	}
}

// Cron job to check elasticbeanstalk enviroments.
func CheckEnviromentsCronJob(){

	// Take a copy of old builds
	oldEnviroments := elasticBeanstalkEnviroments

	// We first reload the env
	if !ReloadElasticBeanstalkEnviroments() {
		fmt.Println("Failed to load enviroments")
	}

	oldEnviromentFound := false

	for _, newEnv := range elasticBeanstalkEnviroments {
		for _, oldEnv := range oldEnviroments {
			// We are comparing two same envs
			if aws.StringValue(oldEnv.EnvironmentName) == aws.StringValue(newEnv.EnvironmentName) {
				oldEnviromentFound = true
				// If env status has changed
				if aws.StringValue(oldEnv.Status) != aws.StringValue(newEnv.Status) {
					attachments := make([]slack.Attachment, 0)
					attachments = append(attachments, ConstructElasticBeanstalkEnviromentAttachment(newEnv, true))
					PostToChannelWithAttachments("Enviroment status updated", attachments)
				}
			}
		}
		if !oldEnviromentFound {
			// We have detected a new env
			attachments := make([]slack.Attachment, 0)
			attachments = append(attachments, ConstructElasticBeanstalkEnviromentAttachment(newEnv, true))
			PostToChannelWithAttachments("New enviroment created ...", attachments)
		}
		oldEnviromentFound = false
	}
}