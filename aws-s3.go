package main

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

type ListS3BucketsTask struct {}

/**
 * Task to get a list of AWS S3 buckets
 * Return: ["list-s3-buckets", "error|ok"] + if error ["error mssage"] + if ok ["bucket-name", ...]
 */
func (t ListS3BucketsTask) Run(args []string, lastResult []string) []string {
	result := make([]string, 0)
	result = append(result, "list-s3-buckets")
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(config.AWSRegion)},
	)
	if err != nil {
		result = append(result, "error")
		result = append(result, err.Error())
		return result
	}
	svc := s3.New(sess)
	response, err := svc.ListBuckets(&s3.ListBucketsInput{})
	if err != nil {
		result = append(result, "error")
		result = append(result, err.Error())
		return result
	}
	result = append(result, "ok")
	for _, bucket := range(response.Buckets) {
		result = append(result, aws.StringValue(bucket.Name))
	}
	return result
}