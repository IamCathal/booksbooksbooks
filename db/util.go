package db

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/go-redis/redis/v9"
	"github.com/iamcathal/booksbooksbooks/dtos"
)

func getKeyForRecentCrawl(shelfURL string) string {
	urlObj, err := url.Parse(shelfURL)
	if err != nil {
		panic(err)
	}
	name := strings.Split(urlObj.Path, "-")
	return fmt.Sprintf("%s-%s", name[len(name)-1], urlObj.Query().Get("shelf"))
}

func removeDuplicateRecentCrawls(recentCrawls []dtos.RecentCrawl) []dtos.RecentCrawl {
	seenShelves := make(map[string]bool)
	noDuplicateRecentCrawls := []dtos.RecentCrawl{}

	for _, crawl := range recentCrawls {
		_, exists := seenShelves[crawl.ShelfURL]
		if !exists {
			seenShelves[crawl.ShelfURL] = true
			noDuplicateRecentCrawls = append(noDuplicateRecentCrawls, crawl)
		}
	}
	return noDuplicateRecentCrawls
}

func isNotRedisNil(err error) bool {
	return err != redis.Nil
}
