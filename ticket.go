package main

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

// createNewTicket creates a new text channel for the ticket in the ModMail category.
func createNewTicket(s *discordgo.Session, user *discordgo.User) (string, error) {
    if cfg.ModMailCategoryID == "" || cfg.StaffRoleID == "" {
        return "", fmt.Errorf("modmail configuration not complete")
    }
    
	channelName := fmt.Sprintf("%s-ticket", strings.ToLower(user.Username))
	if len(channelName) > 100 { 
        channelName = channelName[:95] + "-ticket"
    }

    // FIX: Get the creation date from the Snowflake ID
    createdAt, err := discordgo.SnowflakeTimestamp(user.ID)
    if err != nil {
        createdAt = time.Now() // Fallback if ID is invalid
    }


	permissionOverwrites := []*discordgo.PermissionOverwrite{
		{
			ID:   cfg.GuildID, // @everyone role
			Type: discordgo.PermissionOverwriteTypeRole,
			Deny: discordgo.PermissionViewChannel,
		},
		{
			ID:   cfg.StaffRoleID,
			Type: discordgo.PermissionOverwriteTypeRole,
			Allow: discordgo.PermissionViewChannel | 
				   discordgo.PermissionSendMessages | 
				   discordgo.PermissionReadMessageHistory,
		},
	}

	ch, err := s.GuildChannelCreateComplex(cfg.GuildID, discordgo.GuildChannelCreateData{
		Name:                 channelName,
		Type:                 discordgo.ChannelTypeGuildText,
		ParentID:             cfg.ModMailCategoryID,
		Topic:                fmt.Sprintf("ModMail ticket for %s (%s)", user.String(), user.ID),
		PermissionOverwrites: permissionOverwrites,
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
            // FIX: Use the extracted timestamp
			{Name: "Joined Discord", Value: createdAt.Format("2 Jan 2006"), Inline: true},
		},
		Color: 0x00FF00, // Green
		Timestamp: time.Now().Format(time.RFC3339),
	})

	return ch.ID, nil
}

// forwardUserMessage forwards a message from the user's DM to the ticket channel as an embed.
func forwardUserMessage(s *discordgo.Session, m *discordgo.MessageCreate, ticketChannelID string) {
	embed := createMessageEmbed(m.Author, m.Content, "User Message", 0x00BFFF) // Deep Sky Blue

	// Check for attachments (images/files/links)
	if len(m.Attachments) > 0 {
		attachment := m.Attachments[0]
		if strings.Contains(attachment.ContentType, "image") {
			embed.Image = &discordgo.MessageEmbedImage{URL: attachment.URL}
		} else {
			embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
				Name:  "Attachment",
				Value: fmt.Sprintf("[%s](%s)", attachment.Filename, attachment.URL),
				Inline: false,
			})
		}
	}
	
	s.ChannelMessageSendEmbed(ticketChannelID, embed)
}

// forwardStaffReply forwards a staff member's message from the ticket channel to the user's DM as an embed.
func forwardStaffReply(s *discordgo.Session, m *discordgo.MessageCreate, userID string) {
	embed := createMessageEmbed(m.Author, m.Content, "Staff Reply", 0xFF8C00) // Dark Orange

	userChannel, err := s.UserChannelCreate(userID)
	if err != nil {
		log.Printf("Error creating DM channel for user %s: %v", userID, err)
		s.ChannelMessageSend(m.ChannelID, "⚠️ Could not DM the user. They may have DMs disabled.")
		return
	}
	
	// Check for attachments (images/files/links)
	if len(m.Attachments) > 0 {
		attachment := m.Attachments[0]
		if strings.Contains(attachment.ContentType, "image") {
			embed.Image = &discordgo.MessageEmbedImage{URL: attachment.URL}
		} else {
			embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
				Name:  "Attachment",
				Value: fmt.Sprintf("[%s](%s)", attachment.Filename, attachment.URL),
				Inline: false,
			})
		}
	}

	_, err = s.ChannelMessageSendEmbed(userChannel.ID, embed)
	if err != nil {
		log.Printf("Error sending staff reply to user %s: %v", userID, err)
		s.ChannelMessageSend(m.ChannelID, "⚠️ Could not send the message to the user.")
		return
	}

	s.MessageReactionAdd(m.ChannelID, m.ID, "✅")
}

// createMessageEmbed is a helper to build a consistent message embed structure.
func createMessageEmbed(author *discordgo.User, content string, title string, color int) *discordgo.MessageEmbed {
	embed := &discordgo.MessageEmbed{
		Title:       title,
		Description: content,
		Color:       color,
		Timestamp:   time.Now().Format(time.RFC3339),
		Author: &discordgo.MessageEmbedAuthor{
			Name:    author.String(),
			IconURL: author.AvatarURL("256"),
		},
		Footer: &discordgo.MessageEmbedFooter{
			Text: fmt.Sprintf("User ID: %s", author.ID),
		},
	}
	return embed
}

// logTranscript sends a log of the ticket to the log channel.
func logTranscript(s *discordgo.Session, channelID string, user *discordgo.User, reason string) {
	if cfg.LogChannelID == "" {
		return
	}

	logEmbed := &discordgo.MessageEmbed{
		Title:       "🔒 Ticket Closed/Deleted",
		Description: fmt.Sprintf("Ticket for **%s** has been logged.", user.String()),
		Fields: []*discordgo.MessageEmbedField{
			{Name: "User", Value: user.String(), Inline: true},
			{Name: "Channel ID", Value: channelID, Inline: true},
			{Name: "Reason", Value: reason, Inline: false},
		},
		Color: 0x808080, // Grey
		Timestamp: time.Now().Format(time.RFC3339),
	}

	s.ChannelMessageSendEmbed(cfg.LogChannelID, logEmbed)

	// Remove from active tickets map
	for uid, cid := range activeTickets {
		if cid == channelID {
			delete(activeTickets, uid)
			break
		}
	}
}
