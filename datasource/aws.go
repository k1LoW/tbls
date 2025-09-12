package datasource

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/k1LoW/tbls/drivers/dynamo"
	"github.com/k1LoW/tbls/schema"
)

// AnalizeDynamodb analyze `dynamodb://`
func AnalyzeDynamodb(urlstr string) (*schema.Schema, error) {
	s := &schema.Schema{}
	u, err := url.Parse(urlstr)
	if err != nil {
		return s, err
	}

	values := u.Query()
	err = setEnvAWSCredentials(values)
	if err != nil {
		return s, err
	}

	region := u.Host

	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(region))
	if err != nil {
		return s, err
	}

	if os.Getenv("AWS_ENDPOINT_URL") != "" {
		cfg.BaseEndpoint = aws.String(os.Getenv("AWS_ENDPOINT_URL"))
	}

	client := dynamodb.NewFromConfig(cfg)
	ctx := context.Background()

	driver, err := dynamo.New(ctx, client)
	if err != nil {
		return s, err
	}

	s.Name = fmt.Sprintf("Amazon DynamoDB (%s)", region)
	err = driver.Analyze(s)
	if err != nil {
		return s, err
	}
	return s, nil
}

func setEnvAWSCredentials(values url.Values) error {
	for k := range values {
		if strings.HasPrefix(k, "aws_") {
			return os.Setenv(strings.ToUpper(k), values.Get(k))
		}
	}
	return nil
}
