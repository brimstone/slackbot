package slackbot

import (
	"errors"
	"os"

	"github.com/brimstone/logger"
	"github.com/slack-go/slack"
)

type Bot struct {
	api      *slack.Client
	Users    map[string]slack.User
	Groups   map[string]slack.Channel
	Channels map[string]slack.Channel
}

func (b *Bot) updateUsers() {
	log := logger.New()
	b.Users = make(map[string]slack.User)
	us, err := b.api.GetUsers()
	if err != nil {
		panic(err)
	}
	for _, u := range us {
		if u.IsBot {
			continue
		}
		if u.Deleted {
			continue
		}
		//log.Printf("%#v\n", u)
		log.Debug("user",
			log.Field("ID", u.ID),
			log.Field("Name", u.Name),
			log.Field("RealName", u.RealName),
			log.Field("DisplayName", u.Profile.DisplayName),
			log.Field("Email", u.Profile.Email),
		)
		b.Users[u.ID] = u
	}
}

func (b *Bot) updateGroups() {
	log := logger.New()
	b.Groups = make(map[string]slack.Channel)
	// Get all groups since there isn't an API to get groups by name
	g, _, _ := b.api.GetConversations(&slack.GetConversationsParameters{
		Types: []string{"private_channel"},
	})
	log.Info("Updating groups",
		log.Field("groups", len(g)),
	)
	for _, group := range g {
		b.Groups[group.ID] = group
		log.Debug("Group",
			log.Field("ID", group.ID),
			log.Field("Name", group.Name),
			log.Field("Topic", group.Topic.Value),
		)
	}
}

func (b *Bot) updateChannels() {
	log := logger.New()
	b.Channels = make(map[string]slack.Channel)
	var cursor string
	for {
		channels, nextCursor, err := b.api.GetConversations(&slack.GetConversationsParameters{
			Cursor: cursor,
		})
		if err != nil {
			panic(err)
		}
		for _, channel := range channels {
			b.Channels[channel.ID] = channel
			log.Debug("Channel",
				log.Field("ID", channel.ID),
				log.Field("Name", channel.Name),
				log.Field("Topic", channel.Topic.Value),
			)
		}
		if nextCursor == "" {
			return
		}
		cursor = nextCursor
	}
}

func (b *Bot) FindUserByName(name string) (slack.User, error) {
	for _, u := range b.Users {
		if u.Name == name {
			return u, nil
		}
	}
	return slack.User{}, errors.New("User not found")
}

func (b *Bot) FindGroupByName(name string) (slack.Channel, error) {
	for _, g := range b.Groups {
		if g.Name == name {
			return g, nil
		}
	}
	return slack.Channel{}, errors.New("Group not found")
}

func (b *Bot) FindChannelByName(name string) (slack.Channel, error) {
	for _, c := range b.Channels {
		if c.Name == name {
			return c, nil
		}
	}
	return slack.Channel{}, errors.New("Channel not found")
}

/*
func (b *Bot) OpenIMChannel(uid string) (bool, bool, string, error) {
	return b.api.OpenIMChannel(uid)
}
*/

func (b *Bot) PostMessage(channelID string, options ...slack.MsgOption) (string, string, error) {
	return b.api.PostMessage(channelID, options...)
}

func (b *Bot) SetChannelTopic(channelID string, topic string) (string, error) {
	// https://api.slack.com/changelog/2020-01-deprecating-antecedents-to-the-conversations-api
	channel, err := b.api.SetTopicOfConversation(channelID, topic)
	if err != nil {
		return "", err
	}
	return channel.Topic.Value, nil
}

func (b *Bot) JoinChannel(channelID string) (*slack.Channel, string, []string, error) {
	return b.api.JoinConversation(channelID)
}

func NewBot() (*Bot, error) {
	slackToken := os.Getenv("SLACK_TOKEN")
	if slackToken == "" {
		return nil, errors.New("SLACK_TOKEN is required")
	}
	b := &Bot{}

	b.api = slack.New(slackToken)
	b.updateUsers()
	b.updateGroups()
	b.updateChannels()

	return b, nil
}
