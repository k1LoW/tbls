package dynamo

import (
	"context"
	"fmt"

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

func (d *Dynamodb) Analyze(s *schema.Schema) error {
	input := &dynamodb.ListTablesInput{}

	// tables
	tables := []*schema.Table{}
	for {
		list, err := d.client.ListTablesWithContext(d.ctx, input)
		if err != nil {
			return err
		}

		for _, t := range list.TableNames {
			input := &dynamodb.DescribeTableInput{
				TableName: t,
			}
			desc, err := d.client.DescribeTableWithContext(d.ctx, input)
			if err != nil {
				return err
			}

			table := &schema.Table{
				Name: *t,
			}
			fmt.Printf("%#v\n", desc)

			tables = append(tables, table)
		}

		input.ExclusiveStartTableName = list.LastEvaluatedTableName

		if list.LastEvaluatedTableName == nil {
			break
		}
	}

	s.Tables = tables

	return nil
}

func (d *Dynamodb) Info() (*schema.Driver, error) {
	driver := &schema.Driver{
		Name:            "dynamodb",
		DatabaseVersion: "",
	}
	return driver, nil
}
