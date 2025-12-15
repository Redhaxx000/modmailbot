package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
)

// Global map to keep track of active tickets: UserID -> TicketChannelID
var activeTickets = make(map[string]string)
var cfg Config

func main() {
	// 1. Load Configuration
	cfg = LoadConfig()
	if cfg.BotToken == "" {
		log.Fatal("DISCORD_BOT_TOKEN environment variable not set.")
	}
    if cfg.GuildID == "" {
        log.Fatal("DISCORD_GUILD_ID environment variable not set.")
    }

	// 2. Create a new Discord session
	dg, err := discordgo.New("Bot " + cfg.BotToken)
	if err != nil {
		log.Fatalf("Error creating Discord session: %v", err)
	}

	// 3. Add event handlers
	dg.AddHandler(ready)
	dg.AddHandler(handleMessageCreate)
	dg.AddHandler(handleInteractionCreate)

	// Set necessary intents
	dg.Identify.Intents = discordgo.IntentsGuildMessages | 
						 discordgo.IntentsDirectMessages |
						 discordgo.IntentsMessageContent |
						 discordgo.IntentsGuilds

	// 4. Open a websocket connection to Discord
	err = dg.Open()
	if err != nil {
		log.Fatalf("Error opening connection: %v", err)
	}

	// 5. Register slash commands
	registerCommands(dg, cfg.GuildID)

	// 6. Start a simple web server for Render health checks
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // Default port if not set by Render
	}

	go func() {
		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprint(w, "ModMail Bot is running!")
		})
		log.Printf("Starting web server on port %s for Render health check...", port)
		if err := http.ListenAndServe(":"+port, nil); err != nil {
			log.Fatalf("Error starting web server: %v", err)
		}
	}()

	// 7. Wait for an interrupt signal
	fmt.Println("Bot is now running. Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, syscall.SIGKILL)
	<-sc

	// 8. Cleanly close down the Discord session
	deregisterCommands(dg, cfg.GuildID)
	dg.Close()
}

func ready(s *discordgo.Session, event *discordgo.Ready) {
	s.UpdateGameStatus(0, "DM me for support!")
	log.Printf("Bot is ready. User: %s#%s", event.User.Username, event.User.Discriminator)
}
