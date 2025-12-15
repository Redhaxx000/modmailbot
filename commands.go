// commands.go

package main

import (
	"fmt"
	"log"

	"github.com/bwmarrin/discordgo"
)

// ... (commands array remains the same) ...
// ... (registerCommands and deregisterCommands remain the same) ...

// --- Command Handlers ---

func handleSetupCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	// CORRECTED: Permission check syntax
	if i.Member.Permissions&discordgo.PermissionAdministrator == 0 {
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
	// CORRECTED: Permission check syntax
	if i.Member.Permissions&discordgo.PermissionAdministrator == 0 {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "❌ You must be an administrator to run this command.",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		return
	}
    // ... (rest of the handleSetConfigCommand remains the same) ...
	options := i.ApplicationCommandData().Options
	var categoryID, logChannelID, staffRoleID string

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
    
    cfg.ModMailCategoryID = categoryID
    cfg.LogChannelID = logChannelID
    cfg.StaffRoleID = staffRoleID
    cfg.SaveConfig()
    
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

// ... (handleClaimCommand, handleCloseCommand, handleDeleteCommand remain the same) ...
