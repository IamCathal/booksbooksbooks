package thebookshop

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/iamcathal/booksbooksbooks/dtos"
)

func checkErr(err error) {
	if err != nil {
		logger.Sugar().Fatal(err)
	}
}

func urlEncodeBookSearch(bookInfo dtos.BasicGoodReadsBook) string {
	searchParam := url.Values{}
	searchString := fmt.Sprintf("%s / %s", bookInfo.Author, bookInfo.Title)
	searchParam.Add("search_query", searchString)
	searchParam.Add("section", "product")
	encoded := searchParam.Encode()
	return strings.ReplaceAll(encoded, "+", "%20")
}

func extractAuthorFromTitle(fullBookTitle string) (string, string) {
	fullBookTitle = strings.TrimSpace(fullBookTitle)
	splitUpBySlash := strings.Split(fullBookTitle, "/")
	if len(splitUpBySlash) == 2 {
		return strings.TrimSpace(splitUpBySlash[0]), strings.TrimSpace(splitUpBySlash[1])
	}
	if len(splitUpBySlash) > 2 {
		return strings.TrimSpace(splitUpBySlash[0]), strings.TrimSpace(strings.Join(splitUpBySlash[1:], "-"))
	}

	splitUpByDash := strings.Split(fullBookTitle, "-")
	if len(splitUpByDash) == 2 {
		return strings.TrimSpace(splitUpByDash[0]), strings.TrimSpace(splitUpByDash[1])
	}
	if len(splitUpByDash) > 2 {
		return strings.TrimSpace(splitUpByDash[0]), strings.TrimSpace(strings.Join(splitUpByDash[1:], "-"))
	}

	return strings.TrimSpace(fullBookTitle), strings.TrimSpace(fullBookTitle)
}
