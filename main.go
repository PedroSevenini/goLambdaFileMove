package main

import (
	"context"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/sns"
)

func handle(ctx context.Context, s3Event events.S3Event) {

	sess := session.Must(session.NewSession(
		&aws.Config{
			Region: aws.String(os.Getenv("AWS_REGION")),
			Credentials: credentials.NewStaticCredentialsFromCreds(
				credentials.Value{
					AccessKeyID:     os.Getenv("AWS_ACCESS_KEY"),
					SecretAccessKey: os.Getenv("AWS_ACCESS_SECRET"),
				},
			),
		},
	))
	svc := s3.New(sess)

	for _, record := range s3Event.Records {
		s3Record := record.S3
		fmt.Printf("[%s - %s] Bucket = %s, Key = %s \n", record.EventSource, record.EventTime, s3Record.Bucket.Name, s3Record.Object.Key)

		if strings.Contains(s3Record.Object.Key, "entrada/") {

			srcKey := "/" + s3Record.Bucket.Name + "/" + s3Record.Object.Key
			destKey := "/processado/" + path.Base(s3Record.Object.Key)
			_, err := svc.CopyObject(
				&s3.CopyObjectInput{
					Bucket:     aws.String(s3Record.Bucket.Name),
					CopySource: aws.String(srcKey),
					Key:        aws.String(destKey),
				},
			)
			fmt.Println(srcKey, destKey)
			if err != nil {
				fmt.Printf("Failed to copy object: %v", err)
				continue
			}

			_, _ = svc.DeleteObject(
				&s3.DeleteObjectInput{
					Bucket: aws.String(s3Record.Bucket.Name),
					Key:    aws.String(s3Record.Object.Key),
				},
			)

			SendSNS(os.Getenv("SNS_ARN"), sess, "Processsado arquivo: "+s3Record.Object.Key)
		}
	}
}

func SendSNS(arn string, sess *session.Session, msg string) {
	svc := sns.New(sess)

	params := &sns.PublishInput{
		Message:  aws.String(msg),
		TopicArn: aws.String(arn),
	}

	resp, err := svc.Publish(params)

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	fmt.Println(resp)

}

func envVariableSet() {
	os.Setenv("AWS_REGION", "us-east-1")
	os.Setenv("AWS_ACCESS_KEY", "")
	os.Setenv("AWS_ACCESS_SECRET", "")
}

func main() {
	envVariableSet()
	lambda.Start(handle)
}
