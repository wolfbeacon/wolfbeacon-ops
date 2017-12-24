package main
import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/codebuild"
)

func ReloadCodeBuildBuilds() bool {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(config.AWSRegion)},
	)
	svc := codebuild.New(sess)
	names, err := svc.ListBuilds(&codebuild.ListBuildsInput{SortOrder: aws.String("ASCENDING")})
	if err != nil {
		return false
	}
	builds, err := svc.BatchGetBuilds(&codebuild.BatchGetBuildsInput{Ids: names.Ids})
	
	if err != nil {
		return false
	}

	codeBuildBuilds = builds.Builds

	return true
}