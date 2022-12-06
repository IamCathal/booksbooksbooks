package util

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/iamcathal/booksbooksbooks/db"
	"github.com/iamcathal/booksbooksbooks/dtos"
)

func SendNewBookIsAvailableNotification(book dtos.TheBookshopBook) {
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
	if onlyWhenFreeShippingKicksIn := db.GetSendAlertOnlyWhenFreeShippingKicksIn(); !onlyWhenFreeShippingKicksIn {
		DeliverWebHook(message)
	}
}

func SendBookIsNoLongerAvailableNotification(book dtos.TheBookshopBook) {
	message := dtos.DiscordMsg{
		Username:   "BooksBooksBooks",
		Avatar_url: "https://cathaloc.dev/static/favicons/ms-icon-150x150.png",
		Embed: []dtos.DiscordEmbed{{
			Author: dtos.EmbedAuthor{
				Name:    "Powered by BooksBooksBooks",
				IconURL: "https://cathaloc.dev/static/favicons/ms-icon-150x150.png",
				URL:     "https://github.com/IamCathal/BooksBooksBooks",
			},
			Title:     fmt.Sprintf("%s - %s is no longer available", book.Author, book.Title),
			Color:     0xe60728,
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
	if onlyWhenFreeShippingKicksIn := db.GetSendAlertOnlyWhenFreeShippingKicksIn(); !onlyWhenFreeShippingKicksIn {
		DeliverWebHook(message)
	}
}

func SendFreeShippingTotalHasKickedInNotification(totalCostOfBooks float64) {
	message := dtos.DiscordMsg{
		Username:   "BooksBooksBooks",
		Avatar_url: "https://cathaloc.dev/static/favicons/ms-icon-150x150.png",
		Embed: []dtos.DiscordEmbed{{
			Author: dtos.EmbedAuthor{
				Name:    "Powered by BooksBooksBooks",
				IconURL: "https://cathaloc.dev/static/favicons/ms-icon-150x150.png",
				URL:     "https://github.com/IamCathal/BooksBooksBooks",
			},
			Title:       fmt.Sprintf("Available books total cost exceeds €20 (it's €%.2f). You can now get free shipping", totalCostOfBooks),
			Description: "http://localhost:2945/available",
			Color:       0x6a2ebf,
			Timestamp:   time.Now().Format(time.RFC3339),
		}},
	}
	DeliverWebHook(message)
}

func DeliverWebHook(msg dtos.DiscordMsg) {
	webhookURL := db.GetDiscordWebhookURL()
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

func FindBooksThatAreNowNotAvailable(yesterdaysBooks, todaysBooks []dtos.AvailableBook) []dtos.AvailableBook {
	booksThatAreNowNotAvailable := []dtos.AvailableBook{}
	yesterdaysBooksMap := make(map[string]bool)

	for _, book := range yesterdaysBooks {
		yesterdaysBooksMap[book.BookInfo.ID] = true
	}

	for _, book := range todaysBooks {
		if _, exists := yesterdaysBooksMap[book.BookInfo.ID]; !exists {
			booksThatAreNowNotAvailable = append(booksThatAreNowNotAvailable, book)
		}
	}

	return booksThatAreNowNotAvailable
}
