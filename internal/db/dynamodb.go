package db

import (
	"context"
	"fmt"

	"flavaflav/internal/models"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

// DynamoDBClient wraps the AWS DynamoDB client
type DynamoDBClient struct {
	client    *dynamodb.Client
	tableName string
}

// NewDynamoDBClient creates a new DynamoDB client
func NewDynamoDBClient(tableName string) (*DynamoDBClient, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %v", err)
	}

	return &DynamoDBClient{
		client:    dynamodb.NewFromConfig(cfg),
		tableName: tableName,
	}, nil
}

// Member operations

// CreateMember creates a new member in DynamoDB
func (db *DynamoDBClient) CreateMember(ctx context.Context, member *models.Member) error {
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
func (db *DynamoDBClient) GetMember(ctx context.Context, discordID string) (*models.Member, error) {
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
func (db *DynamoDBClient) UpdateMember(ctx context.Context, member *models.Member) error {
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

// GetAllMembers retrieves all members
func (db *DynamoDBClient) GetAllMembers(ctx context.Context) ([]*models.Member, error) {
	result, err := db.client.Scan(ctx, &dynamodb.ScanInput{
		TableName:        aws.String(db.tableName),
		FilterExpression: aws.String("attribute_exists(discord_id) AND attribute_not_exists(link_id) AND attribute_not_exists(distribution_id)"),
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

// Inventory operations

// CreateInventoryLink creates a new inventory link
func (db *DynamoDBClient) CreateInventoryLink(ctx context.Context, link *models.InventoryLink) error {
	item, err := attributevalue.MarshalMap(link)
	if err != nil {
		return fmt.Errorf("failed to marshal inventory link: %v", err)
	}

	_, err = db.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(db.tableName),
		Item:      item,
	})
	if err != nil {
		return fmt.Errorf("failed to create inventory link: %v", err)
	}

	return nil
}

// GetInventoryLink retrieves a specific inventory link by ID
func (db *DynamoDBClient) GetInventoryLink(ctx context.Context, linkID string) (*models.InventoryLink, error) {
	result, err := db.client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(db.tableName),
		Key: map[string]types.AttributeValue{
			"link_id": &types.AttributeValueMemberS{Value: linkID},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get inventory link: %v", err)
	}

	if result.Item == nil {
		return nil, fmt.Errorf("inventory link not found")
	}

	var link models.InventoryLink
	err = attributevalue.UnmarshalMap(result.Item, &link)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal inventory link: %v", err)
	}

	return &link, nil
}

// UpdateInventoryLink updates an existing inventory link
func (db *DynamoDBClient) UpdateInventoryLink(ctx context.Context, link *models.InventoryLink) error {
	item, err := attributevalue.MarshalMap(link)
	if err != nil {
		return fmt.Errorf("failed to marshal inventory link: %v", err)
	}

	_, err = db.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(db.tableName),
		Item:      item,
	})
	if err != nil {
		return fmt.Errorf("failed to update inventory link: %v", err)
	}

	return nil
}

// GetAvailableInventoryLinks retrieves all available inventory links
func (db *DynamoDBClient) GetAvailableInventoryLinks(ctx context.Context) ([]*models.InventoryLink, error) {
	result, err := db.client.Scan(ctx, &dynamodb.ScanInput{
		TableName:        aws.String(db.tableName),
		FilterExpression: aws.String("attribute_exists(link_id) AND is_available = :available"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":available": &types.AttributeValueMemberBOOL{Value: true},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to scan inventory links: %v", err)
	}

	var links []*models.InventoryLink
	for _, item := range result.Items {
		var link models.InventoryLink
		err = attributevalue.UnmarshalMap(item, &link)
		if err != nil {
			continue // Skip invalid items
		}
		links = append(links, &link)
	}

	return links, nil
}

// GetAvailableInventoryLinksByQuality retrieves available links by quality
func (db *DynamoDBClient) GetAvailableInventoryLinksByQuality(ctx context.Context, quality string) ([]*models.InventoryLink, error) {
	result, err := db.client.Scan(ctx, &dynamodb.ScanInput{
		TableName:        aws.String(db.tableName),
		FilterExpression: aws.String("attribute_exists(link_id) AND is_available = :available AND quality = :quality"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":available": &types.AttributeValueMemberBOOL{Value: true},
			":quality":   &types.AttributeValueMemberS{Value: quality},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to scan inventory links by quality: %v", err)
	}

	var links []*models.InventoryLink
	for _, item := range result.Items {
		var link models.InventoryLink
		err = attributevalue.UnmarshalMap(item, &link)
		if err != nil {
			continue // Skip invalid items
		}
		links = append(links, &link)
	}

	return links, nil
}

// Distribution operations

// CreateDistribution creates a new distribution record
func (db *DynamoDBClient) CreateDistribution(ctx context.Context, distribution *models.Distribution) error {
	item, err := attributevalue.MarshalMap(distribution)
	if err != nil {
		return fmt.Errorf("failed to marshal distribution: %v", err)
	}

	_, err = db.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(db.tableName),
		Item:      item,
	})
	if err != nil {
		return fmt.Errorf("failed to create distribution: %v", err)
	}

	return nil
}

// GetDistributionsByMember retrieves all distributions for a specific member
func (db *DynamoDBClient) GetDistributionsByMember(ctx context.Context, memberID string) ([]*models.Distribution, error) {
	result, err := db.client.Scan(ctx, &dynamodb.ScanInput{
		TableName:        aws.String(db.tableName),
		FilterExpression: aws.String("attribute_exists(distribution_id) AND member_id = :member_id"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":member_id": &types.AttributeValueMemberS{Value: memberID},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to scan distributions: %v", err)
	}

	var distributions []*models.Distribution
	for _, item := range result.Items {
		var distribution models.Distribution
		err = attributevalue.UnmarshalMap(item, &distribution)
		if err != nil {
			continue // Skip invalid items
		}
		distributions = append(distributions, &distribution)
	}

	return distributions, nil
}

// GetAllDistributions retrieves all distribution records
func (db *DynamoDBClient) GetAllDistributions(ctx context.Context) ([]*models.Distribution, error) {
	result, err := db.client.Scan(ctx, &dynamodb.ScanInput{
		TableName:        aws.String(db.tableName),
		FilterExpression: aws.String("attribute_exists(distribution_id)"),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to scan all distributions: %v", err)
	}

	var distributions []*models.Distribution
	for _, item := range result.Items {
		var distribution models.Distribution
		err = attributevalue.UnmarshalMap(item, &distribution)
		if err != nil {
			continue // Skip invalid items
		}
		distributions = append(distributions, &distribution)
	}

	return distributions, nil
}

// Distribution List operations

// CreateDistributionList creates a new distribution list
func (db *DynamoDBClient) CreateDistributionList(ctx context.Context, list *models.DistributionList) error {
	item, err := attributevalue.MarshalMap(list)
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

// GetDistributionList retrieves a distribution list by ID
func (db *DynamoDBClient) GetDistributionList(ctx context.Context, listID string) (*models.DistributionList, error) {
	result, err := db.client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(db.tableName),
		Key: map[string]types.AttributeValue{
			"list_id": &types.AttributeValueMemberS{Value: listID},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get distribution list: %v", err)
	}

	if result.Item == nil {
		return nil, fmt.Errorf("distribution list not found")
	}

	var list models.DistributionList
	err = attributevalue.UnmarshalMap(result.Item, &list)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal distribution list: %v", err)
	}

	return &list, nil
}

// UpdateDistributionList updates an existing distribution list
func (db *DynamoDBClient) UpdateDistributionList(ctx context.Context, list *models.DistributionList) error {
	item, err := attributevalue.MarshalMap(list)
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

// GetActiveDistributionLists retrieves all active distribution lists
func (db *DynamoDBClient) GetActiveDistributionLists(ctx context.Context) ([]*models.DistributionList, error) {
	result, err := db.client.Scan(ctx, &dynamodb.ScanInput{
		TableName:        aws.String(db.tableName),
		FilterExpression: aws.String("attribute_exists(list_id) AND is_active = :active"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":active": &types.AttributeValueMemberBOOL{Value: true},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to scan distribution lists: %v", err)
	}

	var lists []*models.DistributionList
	for _, item := range result.Items {
		var list models.DistributionList
		err = attributevalue.UnmarshalMap(item, &list)
		if err != nil {
			continue // Skip invalid items
		}
		lists = append(lists, &list)
	}

	return lists, nil
}
