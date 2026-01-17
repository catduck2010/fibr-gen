package core

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type MockDynamoDBClient struct {
	ScanFunc func(ctx context.Context, params *dynamodb.ScanInput, optFns ...func(*dynamodb.Options)) (*dynamodb.ScanOutput, error)
}

func (m *MockDynamoDBClient) Scan(ctx context.Context, params *dynamodb.ScanInput, optFns ...func(*dynamodb.Options)) (*dynamodb.ScanOutput, error) {
	return m.ScanFunc(ctx, params, optFns...)
}

func TestDynamoDBDataFetcher_Fetch(t *testing.T) {
	mockClient := &MockDynamoDBClient{
		ScanFunc: func(ctx context.Context, params *dynamodb.ScanInput, optFns ...func(*dynamodb.Options)) (*dynamodb.ScanOutput, error) {
			// Verify params
			if *params.TableName != "test_view" {
				t.Errorf("TableName = %v, want test_view", *params.TableName)
			}
			
			// We expect a filter expression because params were provided
			if params.FilterExpression == nil {
				t.Error("FilterExpression is nil")
			}
			
			// Verify logic for params map[string]string{"id": "123"}
			// names should contain one entry mapping to "id"
			foundKey := false
			for _, v := range params.ExpressionAttributeNames {
				if v == "id" {
					foundKey = true
				}
			}
			if !foundKey {
				t.Error("ExpressionAttributeNames should contain 'id'")
			}

			return &dynamodb.ScanOutput{
				Items: []map[string]types.AttributeValue{
					{
						"id":   &types.AttributeValueMemberS{Value: "123"},
						"name": &types.AttributeValueMemberS{Value: "Test Name"},
					},
				},
				Count: 1,
			}, nil
		},
	}

	fetcher := &DynamoDBDataFetcher{Client: mockClient}
	results, err := fetcher.Fetch("test_view", map[string]string{"id": "123"})
	if err != nil {
		t.Fatalf("Fetch error: %v", err)
	}

	if len(results) != 1 {
		t.Fatalf("results count = %d, want 1", len(results))
	}
	if results[0]["name"] != "Test Name" {
		t.Errorf("name = %v, want Test Name", results[0]["name"])
	}
}
