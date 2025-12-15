package main

import (
	"encoding/json"
	"log"
	"os"
)

// Configuration struct to hold settings loaded from environment variables/file
type Config struct {
	BotToken          string
	GuildID           string // The main server ID where modmail operates
	ModMailCategoryID string // Category ID where ticket channels will be created
	LogChannelID      string // Channel ID for transcripts and logs
	StaffRoleID       string // Role ID that can interact with tickets
}

const configFileName = "config.json"

// LoadConfig initializes the configuration from environment variables AND a configuration file.
func LoadConfig() Config {
	cfg := Config{
		BotToken: os.Getenv("DISCORD_BOT_TOKEN"),
		GuildID:  os.Getenv("DISCORD_GUILD_ID"),
	}

	data, err := os.ReadFile(configFileName)
	if err == nil {
		// Found config file, try to unmarshal
		if err := json.Unmarshal(data, &cfg); err != nil {
			log.Printf("Error unmarshalling config file: %v. Using defaults/ENV.", err)
		} else {
			log.Println("Configuration loaded from config.json.")
		}
	} else if !os.IsNotExist(err) {
		log.Printf("Error reading config file: %v. Using defaults/ENV.", err)
	} else {
		log.Println("config.json not found. Configuration will be saved after setup.")
	}

	return cfg
}

// SaveConfig writes the current configuration to a JSON file.
func (c *Config) SaveConfig() {
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		log.Printf("Error marshalling config: %v", err)
		return
	}

	// Render allows writing to the current directory, which is non-ephemeral storage
    // for this purpose (until the next build/deploy).
	if err := os.WriteFile(configFileName, data, 0644); err != nil {
		log.Printf("Error writing config file: %v", err)
	} else {
		log.Println("Configuration successfully saved to config.json.")
	}
}
