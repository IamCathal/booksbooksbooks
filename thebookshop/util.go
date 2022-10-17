package thebookshop

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/iamcathal/booksbooksbooks/dtos"
)

func checkErr(err error) {
	if err != nil {
		panic(err)
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
		return splitUpBySlash[0], splitUpBySlash[1]
	}

	splitUpByDash := strings.Split(fullBookTitle, "-")
	if len(splitUpByDash) >= 2 {
		return splitUpByDash[0], splitUpByDash[1]
	}

	return splitUpByDash[0], splitUpByDash[0]
	// panic(fmt.Sprintf("failed to split up %s", fullBookTitle))
}
