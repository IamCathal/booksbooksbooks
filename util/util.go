package util

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/iamcathal/booksbooksbooks/dtos"
)

func SendNewBookIsAvailableMessage(book dtos.TheBookshopBook) {
	msgEmbeds := dtos.DiscordMsg{
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
			// Image: dtos.EmbedImage{
			// 	URL: book.Cover,
			// },
			// Footer: dtos.EmbedFooter{
			// 	Text:    "https://github.com/IamCathal/BooksBooksBooks",
			// 	IconURL: "https://github.githubassets.com/images/modules/logos_page/GitHub-Mark.png",
			// },
		}},
	}
	DeliverWebHook(msgEmbeds)
}

func DeliverWebHook(msg dtos.DiscordMsg) {
	webhookURL := os.Getenv("DISCORD_WEBHOOK_URL")
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
