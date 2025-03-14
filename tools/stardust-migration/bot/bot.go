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

	LastUpdateMessageTS string
	FirstMessageTS      string
}

type NoopBot struct{}

func (b NoopBot) PostWelcomeMessage(msg string) (string, string, error) {
	return "", "", nil
}

func (b NoopBot) PostStatusUpdate(msg string, options ...slack.MsgOption) (string, string, error) {
	return "", "", nil
}

func (b NoopBot) InvalidateStatusUpdateReference() {
}

func (NoopBot) PostMessage(msg string, options ...slack.MsgOption) (string, string, error) {
	return "", "", nil
}

type Bot interface {
	PostWelcomeMessage(msg string) (string, string, error)
	PostMessage(msg string, options ...slack.MsgOption) (string, string, error)
	PostStatusUpdate(msg string, options ...slack.MsgOption) (string, string, error)
	InvalidateStatusUpdateReference()
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

func (b *ActiveBot) PostWelcomeMessage(msg string) (string, string, error) {
	_, firstMsgTS, err := b.Client.PostMessage(b.ChannelID, slack.MsgOptionText(msg, false))
	if err != nil {
		return "", "", err
	}

	b.FirstMessageTS = firstMsgTS
	return "", firstMsgTS, nil
}

func (b *ActiveBot) PostMessage(msg string, options ...slack.MsgOption) (string, string, error) {
	allOptions := make([]slack.MsgOption, 0, len(options)+2)
	allOptions = append(allOptions, slack.MsgOptionText(msg, false))
	allOptions = append(allOptions, slack.MsgOptionTS(b.FirstMessageTS))
	allOptions = append(allOptions, options...)

	return b.Client.PostMessage(b.ChannelID, allOptions...)
}

// PostStatusUpdate writes an initial message. Any reoccouring call will update the posted message, not write a new one
// This is used for Status Updates, so we don't write a new message each minute, flooding the channel
func (b *ActiveBot) PostStatusUpdate(msg string, options ...slack.MsgOption) (string, string, error) {
	if len(b.LastUpdateMessageTS) == 0 {
		allOptions := make([]slack.MsgOption, 0, len(options)+2)
		allOptions = append(allOptions, slack.MsgOptionText(msg, false))
		allOptions = append(allOptions, slack.MsgOptionTS(b.FirstMessageTS))
		allOptions = append(allOptions, options...)

		a, ts, err := b.Client.PostMessage(b.ChannelID, allOptions...)
		if err != nil {
			return "", "", err
		}

		b.LastUpdateMessageTS = ts
		return a, ts, nil
	}

	allOptions := make([]slack.MsgOption, 0, len(options)+2)
	allOptions = append(allOptions, slack.MsgOptionText(msg, false))
	allOptions = append(allOptions, slack.MsgOptionUpdate(b.LastUpdateMessageTS))
	allOptions = append(allOptions, options...)

	return b.Client.PostMessage(b.ChannelID, allOptions...)
}

// InvalidateStatusUpdateReference will reset the status update message reference, causing the next StatusUpdate to be posted as a new message again
func (b *ActiveBot) InvalidateStatusUpdateReference() {
	b.LastUpdateMessageTS = ""
}
