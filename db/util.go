package db

import (
	"fmt"
	"net/url"
	"strings"
)

func getKeyForRecentCrawl(shelfURL string) string {
	// https://www.goodreads.com/review/list/151819645-cathal?ref=nav_mybooks&shelf=to-read
	urlObj, err := url.Parse(shelfURL)
	if err != nil {
		panic(err)
	}
	name := strings.Split(urlObj.Path, "-")
	return fmt.Sprintf("%s-%s", name[len(name)-1], urlObj.Query().Get("shelf"))
}
