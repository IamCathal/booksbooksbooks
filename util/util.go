package util

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/iamcathal/booksbooksbooks/db"
	"github.com/iamcathal/booksbooksbooks/dtos"
)

func SendNewBookIsAvailableMessage(book dtos.TheBookshopBook) {
	message := dtos.DiscordMsg{
		Username:   "BooksBooksBooks",
		Avatar_url: "https://cathaloc.dev/static/favicons/ms-icon-150x150.png",
		Embed: []dtos.DiscordEmbed{{
			Author: dtos.EmbedAuthor{
				Name:    "Powered by BooksBooksBooks",
				IconURL: "https://cathaloc.dev/static/favicons/ms-icon-150x150.png",
				URL:     "https://github.com/IamCathal/BooksBooksBooks",
			},
			Title:       fmt.Sprintf("%s - %s is now available", book.Author, book.Title),
			Description: book.Link,
			Fields: []dtos.EmbedField{
				{
					Name:   "Price",
					Value:  book.Price,
					Inline: false,
				},
			},
			Color:     0x03fc90,
			Timestamp: time.Now().Format(time.RFC3339),
			Thumbnail: dtos.EmbedImage{
				URL: book.Cover,
			},
		}},
	}

	messageFormat := db.GetDiscordMessageFormat()
	if messageFormat == "big" {
		message.Embed[0].Thumbnail = dtos.EmbedImage{}
		message.Embed[0].Image = dtos.EmbedImage{
			URL: book.Cover,
		}
	}

	DeliverWebHook(message, getDefaultWebhookURL())
}

func DeliverWebHook(msg dtos.DiscordMsg, webhookURL string) {
	if webhookURL == "" {
		return
	}

	msgEmbedByte, err := json.Marshal(msg)
	if err != nil {
		log.Fatal(err)
	}
	res, err := http.Post(webhookURL, "application/json", bytes.NewBuffer(msgEmbedByte))
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()
}

func getDefaultWebhookURL() string {
	return os.Getenv("DISCORD_WEBHOOK_URL")
}
