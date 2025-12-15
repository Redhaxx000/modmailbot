package main

import (
	"fmt"
	"log"

	"github.com/bwmarrin/discordgo"
)

var commands = []*discordgo.ApplicationCommand{
	{
		Name:        "modmail-setup",
		Description: "Verify the current ModMail setup (Admin only)",
	},
	{
		Name:        "modmail-set-config", 
		Description: "Set core configuration (Category, Log Channel, Staff Role)",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionChannel,
				Name:        "category",
				Description: "The CATEGORY where new tickets will be created.",
				Required:    true,
				ChannelTypes: []discordgo.ChannelType{discordgo.ChannelTypeGuildCategory},
			},
			{
				Type:        discordgo.ApplicationCommandOptionChannel,
				Name:        "log-channel",
				Description: "The TEXT CHANNEL for logging transcripts and closures.",
				Required:    true,
				ChannelTypes: []discordgo.ChannelType{discordgo.ChannelTypeGuildText},
			},
			{
				Type:        discordgo.ApplicationCommandOptionRole,
				Name:        "staff-role",
				Description: "The ROLE whose members can reply to tickets.",
				Required:    true,
			},
		},
	},
	{
		Name:        "claim",
		Description: "Claim a ModMail ticket",
	},
	{
		Name:        "close",
		Description: "Close the current ModMail ticket (preserves channel)",
	},
	{
		Name:        "delete",
		Description: "Close and permanently delete the current ModMail ticket",
	},
}

func registerCommands(s *discordgo.Session, guildID string) {
	log.Println("Registering commands...")
	for _, v := range commands {
		_, err := s.ApplicationCommandCreate(s.State.User.ID, guildID, v)
		if err != nil {
			log.Fatalf("Cannot create command '%s': %v", v.Name, err)
		}
	}
	log.Println("Commands registered.")
}

func deregisterCommands(s *discordgo.Session, guildID string) {
	registeredCommands, _ := s.ApplicationCommands(s.State.User.ID, guildID)
	log.Println("Deregistering commands...")
	for _, v := range registeredCommands {
		s.ApplicationCommandDelete(s.State.User.ID, guildID, v.ID)
	}
	log.Println("Commands deregistered.")
}

// --- Command Handlers ---

func handleSetupCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if !i.Member.Permissions&discordgo.PermissionAdministrator != 0 {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "❌ You must be an administrator to run this command.",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		return
	}
	
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: fmt.Sprintf("✅ **Current Config Status:**\n- Category ID: `%s`\n- Log Channel ID: `%s`\n- Staff Role ID: `%s`\n\nUse `/modmail-set-config` to change these settings.",
				cfg.ModMailCategoryID, cfg.LogChannelID, cfg.StaffRoleID),
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
}

func handleSetConfigCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	// 1. Check for Admin Permissions
	if !i.Member.Permissions&discordgo.PermissionAdministrator != 0 {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "❌ You must be an administrator to run this command.",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		return
	}

	options := i.ApplicationCommandData().Options
	var categoryID, logChannelID, staffRoleID string

	// 2. Extract Options
	for _, option := range options {
		switch option.Name {
		case "category":
			categoryID = option.ChannelValue(s).ID
		case "log-channel":
			logChannelID = option.ChannelValue(s).ID
		case "staff-role":
			staffRoleID = option.RoleValue(s, i.GuildID).ID
		}
	}
    
    // 3. Update Global Config and Save to File
    cfg.ModMailCategoryID = categoryID
    cfg.LogChannelID = logChannelID
    cfg.StaffRoleID = staffRoleID
    cfg.SaveConfig() // Save the updated configuration persistently
    
    // 4. Send Confirmation Response
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: fmt.Sprintf("✅ **ModMail Configuration Updated!**\n"+
				"* Category ID: `%s`\n"+
				"* Log Channel ID: `%s`\n"+
				"* Staff Role ID: `%s`\n\n"+
				"The changes are now persistent. The bot is ready to receive DMs!",
				categoryID, logChannelID, staffRoleID),
			Flags: discordgo.MessageFlagsEphemeral,
		},
	})
}


func handleClaimCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	channel, _ := s.State.Channel(i.ChannelID)
	if channel.ParentID != cfg.ModMailCategoryID {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "❌ This command can only be used in a ModMail ticket channel.",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		return
	}

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: fmt.Sprintf("✅ Ticket claimed by **%s**.", i.Member.User.String()),
		},
	})
}

func handleCloseCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	channel, _ := s.State.Channel(i.ChannelID)
	if channel.ParentID != cfg.ModMailCategoryID {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "❌ This command can only be used in a ModMail ticket channel.",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		return
	}
	
	var userID string
	for uid, cid := range activeTickets {
		if cid == i.ChannelID {
			userID = uid
			break
		}
	}
	
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "✅ Closing ticket... Logging transcript and archiving channel (channel remains visible).",
		},
	})
	
	if userID != "" {
		user, _ := s.User(userID)
		dmChannel, _ := s.UserChannelCreate(userID)
		s.ChannelMessageSend(dmChannel.ID, fmt.Sprintf(
			"🔒 Your support ticket has been closed by **%s**. It may be reopened if needed.", i.Member.User.String(),
		))
		logTranscript(s, i.ChannelID, user, "Closed by staff: "+i.Member.User.String())
	}
}

func handleDeleteCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	channel, _ := s.State.Channel(i.ChannelID)
	if channel.ParentID != cfg.ModMailCategoryID {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "❌ This command can only be used in a ModMail ticket channel.",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		return
	}
	
	var userID string
	for uid, cid := range activeTickets {
		if cid == i.ChannelID {
			userID = uid
			break
		}
	}
	
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "🗑️ Deleting ticket... Logging transcript and permanently removing channel.",
		},
	})
	
	if userID != "" {
		user, _ := s.User(userID)
		
		dmChannel, _ := s.UserChannelCreate(userID)
		s.ChannelMessageSend(dmChannel.ID, fmt.Sprintf(
			"🔒 Your support ticket has been closed and deleted by **%s**.", i.Member.User.String(),
		))
		
		logTranscript(s, i.ChannelID, user, "Deleted by staff: "+i.Member.User.String())
	}

	// Delete the channel immediately after logging/responding
	s.ChannelDelete(i.ChannelID)
}
