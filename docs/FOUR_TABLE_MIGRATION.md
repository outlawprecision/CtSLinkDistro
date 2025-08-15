# Four-Table Architecture Migration

## Overview
This document describes the migration from a single DynamoDB table to a four-table architecture for better data organization and scalability.

## Architecture Changes

### Previous Architecture (Single Table)
- **Single Table**: `flavaflav-users-{env}`
  - Stored all data types: members, inventory, distributions, and lists
  - Used different partition keys for different entity types
  - Led to conflicts and the 500 error when adding inventory

### New Architecture (Four Tables)
1. **Members Table**: `flavaflav-members-{env}`
   - Stores guild member profiles only
   - Primary Key: `discord_id`
   
2. **Inventory Table**: `flavaflav-inventory-{env}`
   - Stores available mastery links
   - Primary Key: `link_id` (sanitized, no spaces)
   - GSI: `availability-quality-index`, `type-quality-index`
   
3. **Distributions Table**: `flavaflav-distributions-{env}`
   - Stores distribution history (audit trail)
   - Primary Key: `distribution_id`
   - GSI: `member-date-index`, `date-index`
   
4. **Lists Table**: `flavaflav-lists-{env}`
   - Stores distribution lists for managing eligible members
   - Primary Key: `list_id`
   - GSI: `active-quality-index`

## Key Fixes Implemented

### 1. Link ID Sanitization
- **Problem**: Link IDs contained spaces (e.g., `gold_Melee Damage_123456`)
- **Solution**: Sanitize link types by replacing spaces and special characters
- **File**: `internal/models/inventory.go`
```go
sanitizedLinkType := strings.ReplaceAll(linkType, " ", "_")
sanitizedLinkType := strings.ReplaceAll(sanitizedLinkType, "/", "_")
sanitizedLinkType := strings.ReplaceAll(sanitizedLinkType, "-", "_")
sanitizedLinkType := strings.ReplaceAll(sanitizedLinkType, "%", "pct")
```

### 2. DynamoDB Client Update
- **File**: `internal/db/dynamodb.go`
- Now accepts four table names
- Separate operations for each table type
- Proper GSI support for complex queries

### 3. Lambda Configuration
- **File**: `cmd/lambda/main.go`
- Accepts four environment variables for table names
- Backward compatibility with single table (fallback)

### 4. CloudFormation Template
- **File**: `cloudformation/flavaflav-infrastructure.yaml`
- Defines four separate DynamoDB tables
- Proper IAM permissions for all tables
- GSIs for efficient querying

## Deployment Steps

### 1. Deploy CloudFormation Stack
```bash
aws cloudformation update-stack \
  --stack-name flavaflav-infrastructure-dev \
  --template-body file://cloudformation/flavaflav-infrastructure.yaml \
  --parameters ParameterKey=Environment,ParameterValue=dev \
  --capabilities CAPABILITY_NAMED_IAM
```

### 2. Build and Deploy Lambda
```bash
# Build Lambda
make build-lambda

# Package Lambda
make package-lambda

# Deploy Lambda (will be done automatically by CloudFormation)
```

### 3. Update Frontend Configuration
The frontend should continue to work as-is, but ensure the API endpoints are correct in `web/static/config.js`.

## Benefits of New Architecture

1. **Clean Separation**: Each table has a single, clear purpose
2. **No Key Conflicts**: Each entity type has its own table with appropriate keys
3. **Better Performance**: Optimized indexes for common queries
4. **Scalability**: Tables can scale independently
5. **Maintainability**: Easier to understand and modify
6. **Cost Efficiency**: Pay-per-request billing, only pay for what you use

## Migration Notes

- **No Data Migration Required**: If starting fresh or no production data
- **Backward Compatibility**: Lambda can fall back to single table if new env vars not set
- **Audit Trail**: Distributions table preserves complete history

## Environment Variables

### Lambda Function
```
DYNAMODB_MEMBERS_TABLE=flavaflav-members-dev
DYNAMODB_INVENTORY_TABLE=flavaflav-inventory-dev
DYNAMODB_DISTRIBUTIONS_TABLE=flavaflav-distributions-dev
DYNAMODB_LISTS_TABLE=flavaflav-lists-dev
```

### Discord Bot (if used)
Same environment variables as Lambda

## Testing

1. **Add Member**: Test creating a new member
2. **Add Inventory**: Test adding links (should no longer get 500 error)
3. **View Inventory**: Check that inventory displays correctly
4. **Distribution**: Test distributing links to members
5. **History**: Verify distribution history is recorded

## Rollback Plan

If issues arise:
1. Keep old CloudFormation stack as backup
2. Revert Lambda code to use single table
3. Update environment variables to point to old table

## Future Improvements

1. Add data migration script if needed for existing data
2. Implement caching layer for frequently accessed data
3. Add CloudWatch dashboards for monitoring
4. Consider adding DynamoDB streams for real-time updates
