//go:build dynamo

package dynamo

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/SouhlInc/tbls/schema"
)

var region = "ap-northeast-1"
var ctx context.Context
var client *dynamodb.DynamoDB

func TestAnalyze(t *testing.T) {
	ctx, client := initClient(t)
	s := &schema.Schema{
		Name: fmt.Sprintf("Amazon DynamoDB (%s)", region),
	}
	driver, err := New(ctx, client)
	if err != nil {
		t.Errorf("%v", err)
	}
	err = driver.Analyze(s)
	if err != nil {
		t.Errorf("%v", err)
	}
	table, err := s.FindTableByName("Thread")
	if err != nil {
		t.Errorf("%v", err)
	}
	want := table.Name
	if want == "" {
		t.Errorf("got not empty string.")
	}
}

func initClient(t *testing.T) (context.Context, *dynamodb.DynamoDB) {
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
	config := aws.NewConfig().WithRegion(region).WithEndpoint("http://localhost:18000")
	client = dynamodb.New(sess, config)
	ctx = context.Background()
	return ctx, client
}
