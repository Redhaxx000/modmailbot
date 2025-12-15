package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/bwmarrin/discordgo"
)

// handleMessageCreate routes incoming messages either from a user DM or a staff reply.
func handleMessageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Ignore bot messages or if config is not set up
	if m.Author.ID == s.State.User.ID || cfg.ModMailCategoryID == "" {
		return
	}

	channel, err := s.State.Channel(m.ChannelID)
	if err != nil {
		log.Printf("Error fetching channel %s: %v", m.ChannelID, err)
		return
	}

	// --- CASE 1: Incoming User DM ---
	if channel.Type == discordgo.ChannelTypeDM {
		ticketChannelID, ok := activeTickets[m.Author.ID]
		if ok {
			forwardUserMessage(s, m, ticketChannelID)
		} else {
			// No active ticket, create a new one.
			newChannelID, err := createNewTicket(s, m.Author)
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, "Sorry, I couldn't create a support ticket. Staff configuration may be incomplete.")
				log.Printf("Error creating new ticket for user %s: %v", m.Author.ID, err)
				return
			}
			activeTickets[m.Author.ID] = newChannelID
			forwardUserMessage(s, m, newChannelID)
			
			s.ChannelMessageSend(m.ChannelID, fmt.Sprintf(
				"Thank you! A new support ticket has been opened. A staff member will respond shortly.",
			))
		}
		return
	}

	// --- CASE 2: Staff Reply in a Ticket Channel ---
	if channel.Type == discordgo.ChannelTypeGuildText && channel.ParentID == cfg.ModMailCategoryID {
		// Find the user ID linked to this ticket channel
		var userID string
		for uid, cid := range activeTickets {
			if cid == m.ChannelID {
				userID = uid
				break
			}
		}

		if userID != "" {
			member, err := s.State.Member(cfg.GuildID, m.Author.ID)
			if err != nil {
				log.Printf("Error fetching member %s: %v", m.Author.ID, err)
				return
			}

			if isStaff(member) {
				forwardStaffReply(s, m, userID)
			}
		}
	}
}

// Check if a member has the required staff role.
func isStaff(member *discordgo.Member) bool {
    if cfg.StaffRoleID == "" {
        return false
    }
	for _, roleID := range member.Roles {
		if roleID == cfg.StaffRoleID {
			return true
		}
	}
	return false
}

// handleInteractionCreate handles all slash command interactions.
func handleInteractionCreate(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.Type == discordgo.InteractionApplicationCommand {
		switch i.ApplicationCommandData().Name {
		case "modmail-setup":
			handleSetupCommand(s, i)
		case "modmail-set-config":
			handleSetConfigCommand(s, i)
		case "claim":
			handleClaimCommand(s, i)
		case "close":
			handleCloseCommand(s, i)
		case "delete":
			handleDeleteCommand(s, i)
		}
	}
}
