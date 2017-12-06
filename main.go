package main

import (
	"encoding/json"
	"fmt"
	slackbot "github.com/BeepBoopHQ/go-slackbot"
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

type Configuration struct {
	SlackKey  string `json:"slack-key"`
	AWSRegion string `json:"aws-region"`
	AnnounceChannel string `json:"announce-channel"`
}

type User struct {
	Email       string   `json:"email"`
	Permissions []string `json:"permissions"`
}

type BuildStatuses map[string]string
type EnvStatuses map[string]string

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
var config Configuration
var buildStatuses map[string]string
var envStatuses map[string]string

func main() {

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
	bot := slackbot.New(config.SlackKey)
	toMe := bot.Messages(slackbot.DirectMessage, slackbot.DirectMention).Subrouter()
	buildStatuses = make(map[string]string)
	envStatuses = make(map[string]string)

	c := cron.New()
	c.AddFunc("*/10 * * * * *", func() { 
		sess, err := session.NewSession(&aws.Config{
			Region: aws.String(config.AWSRegion)},
		)
	
		svc := codebuild.New(sess)
	
		names, err := svc.ListBuilds(&codebuild.ListBuildsInput{SortOrder: aws.String("ASCENDING")})
	
		if err != nil {
			fmt.Println("Failed to load builds")
		}
	
		builds, err := svc.BatchGetBuilds(&codebuild.BatchGetBuildsInput{Ids: names.Ids})
	
		if err != nil {
			fmt.Println("Failed to load builds")
		}

		for id, status := range buildStatuses {
			for _, build := range builds.Builds {
				if *build.Id == id {
					if status != aws.StringValue(build.BuildStatus) {
						// Build status has changed
						bot.RTM.SendMessage(bot.RTM.NewOutgoingMessage("Build status has changed for build `" + *build.Id + "`, new status: " + aws.StringValue(build.BuildStatus), config.AnnounceChannel))
						buildStatuses[id] = aws.StringValue(build.BuildStatus)
					}
				}
			}
		}

		

		
	})

	c.AddFunc("*/10 * * * * *", func() { 
		sess, err := session.NewSession(&aws.Config{
			Region: aws.String(config.AWSRegion)},
		)

		ebSvc := elasticbeanstalk.New(sess)

		environments, err := ebSvc.DescribeEnvironments(&elasticbeanstalk.DescribeEnvironmentsInput{})
		
		if err != nil {
			fmt.Println("Failed to list envs")
		}

		for _, env := range environments.Environments {
			if _, ok := envStatuses[aws.StringValue(env.EnvironmentName)]; ok {
				if envStatuses[aws.StringValue(env.EnvironmentName)] != aws.StringValue(env.HealthStatus) {
					bot.RTM.SendMessage(bot.RTM.NewOutgoingMessage("Enviroment health status changed `" + aws.StringValue(env.EnvironmentName) + "`, new status: " + aws.StringValue(env.HealthStatus), config.AnnounceChannel))
					envStatuses[aws.StringValue(env.EnvironmentName)] = aws.StringValue(env.HealthStatus)
				}
			} else {
				envStatuses[aws.StringValue(env.EnvironmentName)] = aws.StringValue(env.HealthStatus)
			}
		}
	})
	c.Start()


	toMe.Hear("(?i)(hi|hello).*").MessageHandler(HelloHandler)
	toMe.Hear("^help$").MessageHandler(HelpHandler)
	toMe.Hear("^me$").MessageHandler(MeHandler)
	toMe.Hear("list projects").MessageHandler(ListCodeBuildProjectsHandler)
	toMe.Hear("list builds").MessageHandler(ListCodeBuildBuildsHandler)
	toMe.Hear("run build .*").MessageHandler(StartCodeBuildBuildHandler)
	toMe.Hear("delete build .*").MessageHandler(DeleteCodeBuildBuildHandler)

	toMe.Hear("list apps").MessageHandler(ListElasticBeanstalkApplicationsHandler)
	toMe.Hear("list envs").MessageHandler(ListElasticBeanstalkEnviromentsHandler)
	toMe.Hear("rebuild env .*").MessageHandler(RebuildElasticBeanstalkEnviromentHandler)

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

func ListCodeBuildProjectsHandler(ctx context.Context, bot *slackbot.Bot, evt *slack.MessageEvent) {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(config.AWSRegion)},
	)

	svc := codebuild.New(sess)

	result, err := svc.ListProjects(
		&codebuild.ListProjectsInput{
			SortBy:    aws.String("NAME"),
			SortOrder: aws.String("ASCENDING")})

	if err != nil {
		bot.Reply(evt, "Failed to list projects", slackbot.WithTyping)
	}


	attachments := make([]slack.Attachment, 0)

	txt := ""

	for _, p := range result.Projects {
		txt += " - " + *p + "\n"
	}

	attachment := slack.Attachment{
		Fallback:   txt,
		Text: txt,
		Color:     "#7CD197",
		Title: "Projects",
		Footer: "Data from AWS CodeBuild",
	}

	attachments = append(attachments, attachment)

	bot.ReplyWithAttachments(evt, attachments, slackbot.WithTyping)

}


func ListCodeBuildBuildsHandler(ctx context.Context, bot *slackbot.Bot, evt *slack.MessageEvent) {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(config.AWSRegion)},
	)

	svc := codebuild.New(sess)

	names, err := svc.ListBuilds(&codebuild.ListBuildsInput{SortOrder: aws.String("ASCENDING")})

	if err != nil {
		bot.Reply(evt, "Failed to list builds", slackbot.WithTyping)
	}

	builds, err := svc.BatchGetBuilds(&codebuild.BatchGetBuildsInput{Ids: names.Ids})

	if err != nil {
        bot.Reply(evt, "Failed to list build details", slackbot.WithTyping)
	}
	
	attachments := make([]slack.Attachment, 0)

	for _, build := range builds.Builds {
        // fmt.Printf("Project: %s\n", aws.StringValue(build.ProjectName))
        // fmt.Printf("Phase:   %s\n", aws.StringValue(build.CurrentPhase))
        // fmt.Printf("Status:  %s\n", aws.StringValue(build.BuildStatus))
		// fmt.Println("")


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
		
		attachments = append(attachments, slack.Attachment{
			Color: c,
			Fallback: aws.StringValue(build.ProjectName),
			Text: aws.StringValue(build.ProjectName) + " | " + statusText,
			Footer: "Started: " + build.StartTime.String() + " | ID: " + aws.StringValue(build.Id),
		})
	}
	
	bot.ReplyWithAttachments(evt, attachments, slackbot.WithTyping)
}

func StartCodeBuildBuildHandler(ctx context.Context, bot *slackbot.Bot, evt *slack.MessageEvent) {
	args := strings.Split(evt.Text, " ")
	if len(args) > 2 {
		sess, err := session.NewSession(&aws.Config{
			Region: aws.String(config.AWSRegion)},
		)

		svc := codebuild.New(sess)

		out, err := svc.StartBuild(&codebuild.StartBuildInput{ProjectName: aws.String(args[2])})
		build := *out.Build

		if err != nil {
			bot.Reply(evt, "Failed to build project", slackbot.WithTyping)
		} else {
			bot.Reply(evt, "Project build started " + *build.Id, slackbot.WithTyping)
			buildStatuses[*build.Id] = "IN_PROGRESS"
			bot.RTM.SendMessage(bot.RTM.NewOutgoingMessage("Project build has started for project " + args[2] + "\nStarted by <@" + evt.User + ">", config.AnnounceChannel))
		}
	} else {
		bot.Reply(evt, "You must specify a project", slackbot.WithTyping)
	}
}

func DeleteCodeBuildBuildHandler(ctx context.Context, bot *slackbot.Bot, evt *slack.MessageEvent) {
	args := strings.Split(evt.Text, " ")
	if len(args) > 2 {
		sess, err := session.NewSession(&aws.Config{
			Region: aws.String(config.AWSRegion)},
		)

		svc := codebuild.New(sess)

		deleteId := make([]*string, 1)
		deleteId[0] = &args[2]

		_, err = svc.BatchDeleteBuilds(&codebuild.BatchDeleteBuildsInput{Ids: deleteId})

		if err != nil {
			bot.Reply(evt, "Failed to delete build", slackbot.WithTyping)
		} else {
			bot.Reply(evt, "Build deleted", slackbot.WithTyping)
		}
	} else {
		bot.Reply(evt, "You must specify a build", slackbot.WithTyping)
	}
}

func ListElasticBeanstalkApplicationsHandler(ctx context.Context, bot *slackbot.Bot, evt *slack.MessageEvent) {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(config.AWSRegion)},
	)

	svc := elasticbeanstalk.New(sess)

	result, err := svc.DescribeApplications(&elasticbeanstalk.DescribeApplicationsInput{})
	
	if err != nil {
		bot.Reply(evt, "Failed to list applications", slackbot.WithTyping)
	}

	attachments := make([]slack.Attachment, 0)

	for _, app := range result.Applications {
		attachments = append(attachments, slack.Attachment{
			Color: "#7CD197",
			Fallback: aws.StringValue(app.ApplicationName),
			Text: aws.StringValue(app.ApplicationName),
			Footer: "Created: " + app.DateCreated.String() + " | Updated: " + app.DateUpdated.String(),
		})
	}

	bot.ReplyWithAttachments(evt, attachments, slackbot.WithTyping)
}

func ListElasticBeanstalkEnviromentsHandler(ctx context.Context, bot *slackbot.Bot, evt *slack.MessageEvent) {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(config.AWSRegion)},
	)

	svc := elasticbeanstalk.New(sess)

	result, err := svc.DescribeEnvironments(&elasticbeanstalk.DescribeEnvironmentsInput{})
	
	if err != nil {
		bot.Reply(evt, "Failed to list environments", slackbot.WithTyping)
	}

	attachments := make([]slack.Attachment, 0)

	var statusText string
	var c string

	for _, env := range result.Environments {
		switch aws.StringValue(env.HealthStatus) {
		case "Ok":
			c = "#7CD197"
			statusText = "Ok ✓"
		case "Degraded":
			c = "#C62828"
			statusText = "Degraded ×"
		case "Severe":
			c = "#C62828"
			statusText = "Fault ×"
		case "Warning":
			c = "#C62828"
			statusText = "Warning ×"
		default:
			c = "#F9A825"
			statusText = "Pending / Unknown"
		}
		
		attachments = append(attachments, slack.Attachment{
			Color: c,
			Fallback: aws.StringValue(env.EnvironmentName),
			Title: aws.StringValue(env.EnvironmentName),
			Text: "\nApp - " + aws.StringValue(env.ApplicationName) + "\nStatus - " + statusText,
			Footer: "Created: " + env.DateCreated.String() + " | Updated: " + env.DateUpdated.String(),
		})
	}

	bot.ReplyWithAttachments(evt, attachments, slackbot.WithTyping)
}

func RebuildElasticBeanstalkEnviromentHandler(ctx context.Context, bot *slackbot.Bot, evt *slack.MessageEvent) {
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
				bot.Reply(evt, "Failed to initialize a rebuild", slackbot.WithTyping)
			} else {
				bot.Reply(evt, "Now rebuilding enviroment... this can take a few minutes to complete.", slackbot.WithTyping)
			}
		}

	} else {
		bot.Reply(evt, "You must specify a enviroment name", slackbot.WithTyping)
	}
	
}