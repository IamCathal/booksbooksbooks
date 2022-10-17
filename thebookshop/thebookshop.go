package thebookshop

import (
	"fmt"
	"io"
	"net/http"

	"github.com/PuerkitoBio/goquery"
	"github.com/iamcathal/booksbooksbooks/dtos"
)

var (
	THE_BOOKSHOP_BASE_URL = "https://thebookshop.ie"
)

func SearchForBook(bookInfo dtos.BasicGoodReadsBook) dtos.TheBookshopBook {
	searchTheBookshop(bookInfo)
	return dtos.TheBookshopBook{}
}

func searchTheBookshop(bookInfo dtos.BasicGoodReadsBook) *goquery.Selection {
	searchURL := fmt.Sprintf("%s/search.php?%s", THE_BOOKSHOP_BASE_URL, urlEncodeBookSearch(bookInfo))
	fmt.Println(searchURL)
	// doc, err := goquery.NewDocumentFromReader(getPage(searchURL))
	// checkErr(err)
	return nil
}

func getPage(pageURL string) io.ReadCloser {
	client := &http.Client{}
	req, err := http.NewRequest("GET", pageURL, nil)
	checkErr(err)

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/74.0.3729.169 Safari/537.36")
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Alt-Used", "cdn11.bigcommerce.com")
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("X-Requested-With", "XMLHttpRequest")
	req.Header.Set("Host", "cdn11.bigcommerce.com")
	req.Header.Set("TE", "trailers")
	req.Header.Set("Referer", "https://thebookshop.ie/")

	res, err := client.Do(req)
	checkErr(err)
	return res.Body
}
