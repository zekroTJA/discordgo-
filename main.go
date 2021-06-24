package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/bytedance/sonic"
	"github.com/joho/godotenv"
)

var (
	ci = flag.String("ci", "", "enabled CD test mode; sets channelid for test message")

	sc = make(chan os.Signal, 1)
)

func main() {
	flag.Parse()

	godotenv.Load()
	dg, err := discordgo.New("Bot " + os.Getenv("DISCORD_TOKEN"))
	check(err)

	dg.UnmarshalFunc = sonic.Unmarshal
	dg.MarshalFunc = sonic.Marshal

	// Register the messageCreate func as a callback for MessageCreate events.
	dg.AddHandler(messageCreate)

	if *ci != "" {
		dg.AddHandler(readyTest)
	}

	// In this example, we only care about receiving message events.
	dg.Identify.Intents = discordgo.IntentsGuildMessages

	// Open a websocket connection to Discord and begin listening.
	err = dg.Open()
	check(err)

	// Wait here until CTRL-C or other term signal is received.
	fmt.Println("Bot is now running.  Press CTRL-C to exit.")
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	// Cleanly close down the Discord session.
	dg.Close()
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Ignore all messages created by the bot itself
	// This isn't required in this specific example but it's a good practice.
	if m.Author.ID == s.State.User.ID {
		return
	}
	// If the message is "ping" reply with "Pong!"
	if m.Content == "ping" {
		s.ChannelMessageSend(m.ChannelID, "Pong!")
	}

	// If the message is "pong" reply with "Ping!"
	if m.Content == "pong" {
		s.ChannelMessageSend(m.ChannelID, "Ping!")
	}
}

func readyTest(s *discordgo.Session, e *discordgo.Ready) {
	const content = "Hello from CD!"
	msg, err := s.ChannelMessageSend(*ci, content)
	check(err)

	msgRec, err := s.ChannelMessage(msg.ChannelID, msg.ID)
	check(err)

	if msgRec.Content != content {
		panic("Message content missmatch")
	}

	fmt.Printf("%+v\n", msgRec)
	sc <- syscall.SIGTERM
}
