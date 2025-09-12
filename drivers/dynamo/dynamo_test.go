//go:build dynamo

package dynamo

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/k1LoW/tbls/schema"
)

var region = "ap-northeast-1"
var ctx context.Context
var client *dynamodb.Client

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

func initClient(t *testing.T) (context.Context, *dynamodb.Client) {
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(region),
		config.WithEndpointResolverWithOptions(aws.EndpointResolverWithOptionsFunc(
			func(service, region string, options ...interface{}) (aws.Endpoint, error) {
				return aws.Endpoint{
					URL: "http://localhost:18000",
				}, nil
			})),
	)
	if err != nil {
		t.Fatalf("unable to load SDK config, %v", err)
	}
	client = dynamodb.NewFromConfig(cfg)
	ctx = context.Background()
	return ctx, client
}
