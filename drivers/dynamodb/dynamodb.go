package dynamodb

import (
	"context"

	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/k1LoW/tbls/schema"
)

type Dynamodb struct {
	ctx    context.Context
	client *dynamodb.DynamoDB
}

func NewDynamodb(ctx context.Context, client *dynamodb.DynamoDB) (*Dynamodb, error) {
	return &Dynamodb{
		ctx:    ctx,
		client: client,
	}, nil
}

func (d *Dynamodb) Analize(s *schema.Schema) error {
	return nil
}

func (d *Dynamodb) Info() (*schema.Driver, error) {
	d := &schema.Driver{
		Name:            "dynamodb",
		DatabaseVersion: "",
	}
	return d, nil
}
