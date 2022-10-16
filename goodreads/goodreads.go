package goodreads

import (
	"fmt"
	"net/http"

	"github.com/PuerkitoBio/goquery"
	"github.com/iamcathal/booksbooksbooks/dtos"
)

func GetBooksFromShelf(shelfURL string) []dtos.BasicGoodReadsBook {
	// is a shelf URL ?

	extractedBooks := extractBooksFromShelfPage(shelfURL)

	return extractedBooks
}

func extractBooksFromShelfPage(shelfURL string) []dtos.BasicGoodReadsBook {
	doc, err := goquery.NewDocumentFromResponse(getPage(shelfURL))
	checkErr(err)

	doc.Find("tbody#booksBody").Each(func(i int, bookReviews *goquery.Selection) {
		fmt.Println("found a bookbody")
		bookReviews.Find("tr").Each(func(k int, bookReviewRow *goquery.Selection) {

			title := bookReviewRow.Find("td#field title, a").Text()

			fmt.Println(fmt.Sprintf("[%d] %v\n", k, stripOfFormatting(title)))

		})
	})

	return []dtos.BasicGoodReadsBook{}
}

func getPage(pageURL string) *http.Response {
	// return string(` This is a random-length HTML comment: kbyakgavvzypvfwnxmeecoqynzidcugnvhjxuvhmlgildzhssfuytcfbaotkuzcoiuzkcfjvmkpfmepbhqiyoyrapyqfqohjfvdxwdybndkahjvoaaotinteszswjrtkkxznqhmwhnhhqjjbidjxhswxjibvlhkqvlvomftkfaayntgjkgzmszeqvhjokstosxlqkxzgkogufnesxtkjgawsqeisymxysuvhauwqpsigwraoincxhirkexncsgzlhuexmsioyiizxuaodwiqfaqadorhgxnsudtxmtixicqamkancmwvuvfibnlmusntrymolgrboxsghehwzkxmaemqdydvaizrozoxzwsyjwrqxxpnoaxmdugfjyiasxboinufsyjjtkouejdfmtqxpldvhcgoshperpzqrnvkblolryxapvaphbtqnvbgitcuztnptumcvyxalgulcibpgb `)
	res, err := http.Get(pageURL)
	checkErr(err)

	return res
}
