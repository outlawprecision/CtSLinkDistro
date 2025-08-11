package models

// Config represents application configuration
type Config struct {
	AWS struct {
		Region          string `json:"region"`
		DynamoDBTable   string `json:"dynamodb_table"`
		AccessKeyID     string `json:"access_key_id"`
		SecretAccessKey string `json:"secret_access_key"`
	} `json:"aws"`

	Discord struct {
		BotToken  string `json:"bot_token"`
		GuildID   string `json:"guild_id"`
		ChannelID string `json:"channel_id"`
	} `json:"discord"`

	Web struct {
		Port string `json:"port"`
		Host string `json:"host"`
	} `json:"web"`

	Rules struct {
		SilverEligibilityDays int `json:"silver_eligibility_days"`
		GoldEligibilityDays   int `json:"gold_eligibility_days"`
		MaxAbsenceCount       int `json:"max_absence_count"`
	} `json:"rules"`
}

// DefaultConfig returns a configuration with default values
func DefaultConfig() *Config {
	return &Config{
		AWS: struct {
			Region          string `json:"region"`
			DynamoDBTable   string `json:"dynamodb_table"`
			AccessKeyID     string `json:"access_key_id"`
			SecretAccessKey string `json:"secret_access_key"`
		}{
			Region:        "us-east-1",
			DynamoDBTable: "flavaflav",
		},
		Discord: struct {
			BotToken  string `json:"bot_token"`
			GuildID   string `json:"guild_id"`
			ChannelID string `json:"channel_id"`
		}{},
		Web: struct {
			Port string `json:"port"`
			Host string `json:"host"`
		}{
			Port: "8080",
			Host: "localhost",
		},
		Rules: struct {
			SilverEligibilityDays int `json:"silver_eligibility_days"`
			GoldEligibilityDays   int `json:"gold_eligibility_days"`
			MaxAbsenceCount       int `json:"max_absence_count"`
		}{
			SilverEligibilityDays: 30,
			GoldEligibilityDays:   90,
			MaxAbsenceCount:       3,
		},
	}
}
