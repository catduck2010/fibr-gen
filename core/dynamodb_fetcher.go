package core

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

// DynamoDBClient defines the interface needed for scanning.
type DynamoDBClient interface {
	Scan(ctx context.Context, params *dynamodb.ScanInput, optFns ...func(*dynamodb.Options)) (*dynamodb.ScanOutput, error)
}

// DynamoDBDataFetcher implements DataFetcher using AWS DynamoDB.
// It maps viewName to a DynamoDB Table Name.
type DynamoDBDataFetcher struct {
	Client DynamoDBClient
}

// NewDynamoDBDataFetcher creates a new fetcher with the given AWS config.
func NewDynamoDBDataFetcher(cfg aws.Config) *DynamoDBDataFetcher {
	return &DynamoDBDataFetcher{
		Client: dynamodb.NewFromConfig(cfg),
	}
}

// Fetch scans the DynamoDB table specified by viewName.
// It applies simple equality filtering based on params if provided.
// Note: Currently assumes all filter values are Strings.
func (f *DynamoDBDataFetcher) Fetch(viewName string, params map[string]string) ([]map[string]interface{}, error) {
	// viewName corresponds to Table Name
	tableName := viewName

	var filterExpression *string
	var expressionAttributeNames map[string]string
	var expressionAttributeValues map[string]types.AttributeValue

	if len(params) > 0 {
		expr := ""
		expressionAttributeNames = make(map[string]string)
		expressionAttributeValues = make(map[string]types.AttributeValue)
		idx := 0
		for k, v := range params {
			if idx > 0 {
				expr += " AND "
			}
			// Use #k for name, :v for value to avoid reserved words conflicts
			kName := fmt.Sprintf("#k%d", idx)
			vName := fmt.Sprintf(":v%d", idx)

			expr += fmt.Sprintf("%s = %s", kName, vName)
			expressionAttributeNames[kName] = k
			// Assuming String value for filter.
			// TODO: Support other types if needed (e.g. check if v is number)
			expressionAttributeValues[vName] = &types.AttributeValueMemberS{Value: v}
			idx++
		}
		filterExpression = aws.String(expr)
	}

	input := &dynamodb.ScanInput{
		TableName:                 aws.String(tableName),
		FilterExpression:          filterExpression,
		ExpressionAttributeNames:  expressionAttributeNames,
		ExpressionAttributeValues: expressionAttributeValues,
	}

	paginator := dynamodb.NewScanPaginator(f.Client, input)
	var items []map[string]interface{}

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(context.TODO())
		if err != nil {
			return nil, fmt.Errorf("failed to scan table %s: %w", tableName, err)
		}

		var pageItems []map[string]interface{}
		if err := attributevalue.UnmarshalListOfMaps(page.Items, &pageItems); err != nil {
			return nil, fmt.Errorf("failed to unmarshal items: %w", err)
		}
		items = append(items, pageItems...)
	}

	return items, nil
}
