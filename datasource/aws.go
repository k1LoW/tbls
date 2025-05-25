package datasource

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/SouhlInc/tbls/drivers/dynamo"
	"github.com/SouhlInc/tbls/schema"
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

	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	config := aws.NewConfig().WithRegion(region)
	if os.Getenv("AWS_ENDPOINT_URL") != "" {
		config = config.WithEndpoint(os.Getenv("AWS_ENDPOINT_URL"))
	}

	client := dynamodb.New(sess, config)
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
