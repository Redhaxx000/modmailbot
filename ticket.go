// ticket.go

package main

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

// ... (createNewTicket function, starting line 39) ...
func createNewTicket(s *discordgo.Session, user *discordgo.User) (string, error) {
    if cfg.ModMailCategoryID == "" || cfg.StaffRoleID == "" {
        return "", fmt.Errorf("modmail configuration not complete")
    }
    
	channelName := fmt.Sprintf("%s-ticket", strings.ToLower(user.Username))
	if len(channelName) > 100 { 
        channelName = channelName[:95] + "-ticket"
    }

    // 1. GET THE CREATED AT TIMESTAMP FROM THE USER ID
    createdAt, err := discordgo.SnowflakeTimestamp(user.ID)
    if err != nil {
        createdAt = time.Now() // Fallback if ID is invalid
    }

	permissionOverwrites := []*discordgo.PermissionOverwrite{
        // ... (permissions remain the same) ...
    }

	ch, err := s.GuildChannelCreateComplex(cfg.GuildID, discordgo.GuildChannelCreateData{
        // ... (channel creation data remains the same) ...
	})

	if err != nil {
		return "", err
	}

	// Send an initial message in the ticket channel
	s.ChannelMessageSendEmbed(ch.ID, &discordgo.MessageEmbed{
		Title:       "🚨 New ModMail Ticket Opened",
		Description: fmt.Sprintf("A new support ticket has been opened by **%s**.", user.String()),
		Fields: []*discordgo.MessageEmbedField{
			{Name: "User ID", Value: user.ID, Inline: true},
            // CORRECTED: Use the extracted timestamp
			{Name: "Joined Discord", Value: createdAt.Format("2 Jan 2006"), Inline: true},
		},
		Color: 0x00FF00,
		Timestamp: time.Now().Format(time.RFC3339),
	})

	return ch.ID, nil
}
// ... (rest of ticket.go remains the same) ...
