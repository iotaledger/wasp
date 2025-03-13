package bot

import (
	"os"

	"github.com/ethereum/go-ethereum/log"
	"github.com/joho/godotenv"
	"github.com/slack-go/slack"
)

type ActiveBot struct {
	Token     string
	ChannelID string
	Client    *slack.Client
	Enabled   bool
}

type NoopBot struct{}

func (NoopBot) PostMessage(msg string, options ...slack.MsgOption) (string, string, error) {
	return "", "", nil
}

type Bot interface {
	PostMessage(msg string, options ...slack.MsgOption) (string, string, error)
}

var bot Bot

func Get() Bot {
	if bot == nil {
		bot = initSlackBot()
	}

	return bot
}

func initSlackBot() Bot {
	err := godotenv.Load(".env")
	if err != nil {
		log.Warn("Error loading .env file. No Slack bot will be enabled.")
		return &NoopBot{}
	}

	token := os.Getenv("SLACK_AUTH_TOKEN")
	channelID := os.Getenv("SLACK_CHANNEL_ID")

	if token == "" || channelID == "" {
		log.Warn("Slack bot disabled - SLACK_AUTH_TOKEN and SLACK_CHANNEL_ID are not set in .env file.")
		return &NoopBot{}
	}

	return &ActiveBot{
		Token:     token,
		ChannelID: channelID,
		Client:    slack.New(token, slack.OptionDebug(false)),
	}
}

func (b *ActiveBot) PostMessage(msg string, options ...slack.MsgOption) (string, string, error) {
	allOptions := make([]slack.MsgOption, 0, len(options)+1)
	allOptions = append(allOptions, slack.MsgOptionText(msg, false))
	allOptions = append(allOptions, options...)

	return b.Client.PostMessage(b.ChannelID, allOptions...)
}
