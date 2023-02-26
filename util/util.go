package util

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/iamcathal/booksbooksbooks/controller"
	"github.com/iamcathal/booksbooksbooks/db"
	"github.com/iamcathal/booksbooksbooks/dtos"
)

var (
	NON_ENGLISH_CHARACTER = regexp.MustCompile(`([^A-Za-z0-9 ,'\\\/\)\(-\.\[\]])`)
)

func SendNewBookIsAvailableNotification(book dtos.TheBookshopBook, available bool) {
	availableText := ""
	if available {
		availableText = fmt.Sprintf("%s - %s is now available", book.Author, book.Title)
	} else {
		availableText = fmt.Sprintf("%s - %s is no longer available", book.Author, book.Title)
	}

	message := dtos.DiscordMsg{
		Username:   "BooksBooksBooks",
		Avatar_url: "https://cathaloc.dev/static/favicons/ms-icon-150x150.png",
		Embed: []dtos.DiscordEmbed{{
			Author: dtos.EmbedAuthor{
				Name:    "Powered by BooksBooksBooks",
				IconURL: "https://cathaloc.dev/static/favicons/ms-icon-150x150.png",
				URL:     "https://github.com/IamCathal/BooksBooksBooks",
			},
			Title:       availableText,
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
		controller.Cnt.DeliverWebhook(message)
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
	if bookIsNoLongerAvailableFlag := db.GetSendAlertWhenBookNoLongerAvailable(); bookIsNoLongerAvailableFlag {
		controller.Cnt.DeliverWebhook(message)
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
	controller.Cnt.DeliverWebhook(message)
}

func IsEnglishText(bookDetail string) bool {
	hasNonEnglishCharacters := NON_ENGLISH_CHARACTER.MatchString(bookDetail)
	if hasNonEnglishCharacters {
		return false
	}

	// A very crude BUT lightweight way of detecting a good amount
	// of non english books. Using an actual language detection
	// library would be like using an airplane to thread a needle
	experimentalNonEnglishSnippets := []string{
		" de ",
		" le ",
		" en ",
		" francais ",
		" del ",
		" el ",
		" los ",
		" las ",
		" und ",
		" der ",
		" des ",
		" dem ",
		" y ",
		" ein ",
		" eine ",
		" einer ",
		" l'",
		" d'",
		" la ",
		" c'est ",
	}
	for _, nonEnglishSnippet := range experimentalNonEnglishSnippets {
		if strings.Contains(strings.ToLower(bookDetail), nonEnglishSnippet) {
			return false
		}
	}

	return true
}
