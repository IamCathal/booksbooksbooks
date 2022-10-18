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

func SearchForBooks(searchBooks []dtos.BasicGoodReadsBook) dtos.AllBookshopBooksSearchResults {
	searchResults := make(dtos.AllBookshopBooksSearchResults)

	for _, bookInfo := range searchBooks {
		searchResults[bookInfo.Title] = SearchForBook(bookInfo)
	}

	return searchResults
}

func SearchForBook(bookInfo dtos.BasicGoodReadsBook) dtos.BookShopBookSearchResult {
	allBooks := searchTheBookshop(bookInfo)
	// for i, book := range allBooks {
	// 	fmt.Printf("[%d] %+v\n", i, book)
	// }
	return dtos.BookShopBookSearchResult{
		SearchResultBooks: allBooks,
	}
}

func searchTheBookshop(bookInfo dtos.BasicGoodReadsBook) []dtos.TheBookshopBook {
	searchURL := fmt.Sprintf("%s/search.php?%s", THE_BOOKSHOP_BASE_URL, urlEncodeBookSearch(bookInfo))
	doc, err := goquery.NewDocumentFromReader(getPage(searchURL))
	checkErr(err)
	fmt.Printf("Search for %s\n", searchURL)
	allBooks := []dtos.TheBookshopBook{}

	doc.Find("ul[class='productGrid']").Each(func(i int, bookReviews *goquery.Selection) {
		bookReviews.Find("li[class='product']").Each(func(k int, bookProduct *goquery.Selection) {

			bookTitle := bookProduct.Find("h4[class='card-title']").Text()
			bookLink, ok := bookProduct.Find("h4[class='card-title'] a").Attr("href")
			if !ok {
				panic(fmt.Sprintf("no bookLink found for %+v with query %s", bookInfo, searchURL))
			}

			bookPrice := bookProduct.Find("span[data-product-price-without-tax='']").Text()
			// fmt.Printf("Title: '%s' Price: '%s' Link: %s\n", bookTitle, bookPrice, bookLink)

			author, title := extractAuthorFromTitle(bookTitle)

			foundBook := dtos.TheBookshopBook{
				Title:  title,
				Author: author,
				Price:  bookPrice,
				Link:   bookLink,
			}
			allBooks = append(allBooks, foundBook)
		})
	})

	return allBooks
}

func getPage(pageURL string) io.ReadCloser {
	fmt.Println(pageURL)
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
