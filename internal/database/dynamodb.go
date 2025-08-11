package database

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"

	"flavaflav/internal/models"
)

// DynamoDBService handles all DynamoDB operations
type DynamoDBService struct {
	client    *dynamodb.Client
	tableName string
}

// NewDynamoDBService creates a new DynamoDB service instance
func NewDynamoDBService(region, tableName string) (*DynamoDBService, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(region))
	if err != nil {
		return nil, fmt.Errorf("unable to load SDK config: %v", err)
	}

	client := dynamodb.NewFromConfig(cfg)

	return &DynamoDBService{
		client:    client,
		tableName: tableName,
	}, nil
}

// Member operations

// CreateMember creates a new member in the database
func (db *DynamoDBService) CreateMember(ctx context.Context, member *models.Member) error {
	item, err := attributevalue.MarshalMap(member)
	if err != nil {
		return fmt.Errorf("failed to marshal member: %v", err)
	}

	_, err = db.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(db.tableName),
		Item:      item,
	})
	if err != nil {
		return fmt.Errorf("failed to create member: %v", err)
	}

	return nil
}

// GetMember retrieves a member by Discord ID
func (db *DynamoDBService) GetMember(ctx context.Context, discordID string) (*models.Member, error) {
	result, err := db.client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(db.tableName),
		Key: map[string]types.AttributeValue{
			"discord_id": &types.AttributeValueMemberS{Value: discordID},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get member: %v", err)
	}

	if result.Item == nil {
		return nil, fmt.Errorf("member not found")
	}

	var member models.Member
	err = attributevalue.UnmarshalMap(result.Item, &member)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal member: %v", err)
	}

	return &member, nil
}

// UpdateMember updates an existing member
func (db *DynamoDBService) UpdateMember(ctx context.Context, member *models.Member) error {
	item, err := attributevalue.MarshalMap(member)
	if err != nil {
		return fmt.Errorf("failed to marshal member: %v", err)
	}

	_, err = db.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(db.tableName),
		Item:      item,
	})
	if err != nil {
		return fmt.Errorf("failed to update member: %v", err)
	}

	return nil
}

// DeleteMember removes a member from the database
func (db *DynamoDBService) DeleteMember(ctx context.Context, discordID string) error {
	_, err := db.client.DeleteItem(ctx, &dynamodb.DeleteItemInput{
		TableName: aws.String(db.tableName),
		Key: map[string]types.AttributeValue{
			"discord_id": &types.AttributeValueMemberS{Value: discordID},
		},
	})
	if err != nil {
		return fmt.Errorf("failed to delete member: %v", err)
	}

	return nil
}

// GetAllMembers retrieves all members from the database
func (db *DynamoDBService) GetAllMembers(ctx context.Context) ([]*models.Member, error) {
	result, err := db.client.Scan(ctx, &dynamodb.ScanInput{
		TableName:        aws.String(db.tableName),
		FilterExpression: aws.String("attribute_exists(discord_id) AND attribute_not_exists(link_id) AND attribute_not_exists(list_type)"),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to scan members: %v", err)
	}

	var members []*models.Member
	for _, item := range result.Items {
		var member models.Member
		err = attributevalue.UnmarshalMap(item, &member)
		if err != nil {
			continue // Skip invalid items
		}
		members = append(members, &member)
	}

	return members, nil
}

// Link History operations

// CreateLinkHistory creates a new link history record
func (db *DynamoDBService) CreateLinkHistory(ctx context.Context, linkHistory *models.LinkHistory) error {
	item, err := attributevalue.MarshalMap(linkHistory)
	if err != nil {
		return fmt.Errorf("failed to marshal link history: %v", err)
	}

	_, err = db.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(db.tableName),
		Item:      item,
	})
	if err != nil {
		return fmt.Errorf("failed to create link history: %v", err)
	}

	return nil
}

// GetLinkHistoryByMember retrieves link history for a specific member
func (db *DynamoDBService) GetLinkHistoryByMember(ctx context.Context, discordID string) ([]*models.LinkHistory, error) {
	result, err := db.client.Query(ctx, &dynamodb.QueryInput{
		TableName:              aws.String(db.tableName),
		IndexName:              aws.String("discord_id-timestamp-index"),
		KeyConditionExpression: aws.String("discord_id = :discord_id"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":discord_id": &types.AttributeValueMemberS{Value: discordID},
		},
		ScanIndexForward: aws.Bool(false), // Most recent first
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query link history: %v", err)
	}

	var history []*models.LinkHistory
	for _, item := range result.Items {
		var linkHistory models.LinkHistory
		err = attributevalue.UnmarshalMap(item, &linkHistory)
		if err != nil {
			continue // Skip invalid items
		}
		history = append(history, &linkHistory)
	}

	return history, nil
}

// Distribution List operations

// CreateDistributionList creates a new distribution list
func (db *DynamoDBService) CreateDistributionList(ctx context.Context, distributionList *models.DistributionList) error {
	item, err := attributevalue.MarshalMap(distributionList)
	if err != nil {
		return fmt.Errorf("failed to marshal distribution list: %v", err)
	}

	_, err = db.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(db.tableName),
		Item:      item,
	})
	if err != nil {
		return fmt.Errorf("failed to create distribution list: %v", err)
	}

	return nil
}

// GetDistributionList retrieves a distribution list by type
func (db *DynamoDBService) GetDistributionList(ctx context.Context, listType string) (*models.DistributionList, error) {
	result, err := db.client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(db.tableName),
		Key: map[string]types.AttributeValue{
			"list_type": &types.AttributeValueMemberS{Value: listType},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get distribution list: %v", err)
	}

	if result.Item == nil {
		return nil, fmt.Errorf("distribution list not found")
	}

	var distributionList models.DistributionList
	err = attributevalue.UnmarshalMap(result.Item, &distributionList)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal distribution list: %v", err)
	}

	return &distributionList, nil
}

// UpdateDistributionList updates an existing distribution list
func (db *DynamoDBService) UpdateDistributionList(ctx context.Context, distributionList *models.DistributionList) error {
	item, err := attributevalue.MarshalMap(distributionList)
	if err != nil {
		return fmt.Errorf("failed to marshal distribution list: %v", err)
	}

	_, err = db.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(db.tableName),
		Item:      item,
	})
	if err != nil {
		return fmt.Errorf("failed to update distribution list: %v", err)
	}

	return nil
}
