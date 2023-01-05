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
