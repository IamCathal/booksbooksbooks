package goodreads

import (
	"encoding/json"
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/iamcathal/booksbooksbooks/controller"
	"github.com/iamcathal/booksbooksbooks/db"
	"github.com/iamcathal/booksbooksbooks/dtos"
	"github.com/segmentio/ksuid"
	"go.uber.org/zap"
	"golang.org/x/net/html"
)

var (
	// There are five spaces between a books
	// title and its series information if
	// the series information is given
	TITLE_AND_SERIES_INFO_SEPERATOR = regexp.MustCompile("[ ]{3,}")
	NUMBER_MATCH                    = regexp.MustCompile("[0-9]{1,}")
	ONLY_NUMBERS                    = regexp.MustCompile(`([0-9]+)`)
	// Goodreads returns 30 books per page
	BOOK_COUNT_PER_PAGE = 30
	// Base URL that book links are built on
	GOODREADS_BASE_BOOK_URL = "https://www.goodreads.com"
	// Crude to check if a roughly  valid
	// shelf URL is being queried
	GOODREADS_SHELF_URL_PREFIX = GOODREADS_BASE_BOOK_URL + "/review/list/"
)

func CheckIsShelfURL(checkURL string) bool {
	hasPrefix := strings.HasPrefix(checkURL, GOODREADS_SHELF_URL_PREFIX)
	properURL, err := url.Parse(checkURL)
	if err != nil {
		logger.Sugar().Fatal(err)
	}

	shelfParam := properURL.Query().Get("shelf")

	return hasPrefix && shelfParam != ""
}

func processBook(fullTitle, author, cover, isbn13, asin, rating, link string) dtos.BasicGoodReadsBook {
	fullTitle = stripOfFormatting(fullTitle)
	author = stripOfFormatting(author)
	cover = stripOfFormatting(cover)
	isbn13 = stripOfFormatting(isbn13)
	asin = stripOfFormatting(asin)
	rating = stripOfFormatting(rating)
	link = GOODREADS_BASE_BOOK_URL + link

	value, err := strconv.ParseFloat(rating, 32)
	if err != nil {
		logger.Sugar().Fatal(err)
	}

	bookTitle, seriesInfo := extractTitleDetailsIfPossible(fullTitle)
	newBook := dtos.BasicGoodReadsBook{
		ID:         ksuid.New().String(),
		Title:      bookTitle,
		Author:     author,
		SeriesText: seriesInfo,
		Link:       link,
		Cover:      cover,
		Isbn13:     isbn13,
		Asin:       asin,
		Rating:     float32(value),
	}
	return newBook
}

func extractBooksFromHTML(doc *goquery.Document) []dtos.BasicGoodReadsBook {
	allBooks := []dtos.BasicGoodReadsBook{}
	doc.Find("#booksBody").Each(func(i int, bookReviews *goquery.Selection) {
		bookReviews.Find("tr").Each(func(k int, bookReviewRow *goquery.Selection) {
			title := bookReviewRow.Find("td[class='field title'] a").Text()
			author := bookReviewRow.Find("td[class='field author'] a").Text()
			cover, _ := bookReviewRow.Find("td[class='field cover'] img").Attr("src")
			isbn13 := bookReviewRow.Find("td[class='field isbn13'] div").Text()
			asin := bookReviewRow.Find("td[class='field asin'] div").Text()
			rating := bookReviewRow.Find("td[class='field avg_rating'] div").Text()
			link, _ := bookReviewRow.Find("td[class='field title'] a").Attr("href")

			currBook := processBook(title, author, cover, isbn13, asin, rating, link)
			allBooks = append(allBooks, currBook)
		})
	})
	return allBooks
}

func GetAvailableBooksFromSearchResult(searchResults []dtos.EnchancedSearchResult) []dtos.AvailableBook {
	availableBooks := []dtos.AvailableBook{}
	for _, searchResult := range searchResults {
		for _, titleMatch := range searchResult.TitleMatches {
			availableBook := dtos.AvailableBook{
				BookInfo:         searchResult.SearchBook,
				BookPurchaseInfo: titleMatch,
			}
			availableBooks = append(availableBooks, availableBook)
		}
		if addMoreAuthorBooks := db.AddOtherAuthorBooksIfFound(); addMoreAuthorBooks {
			for _, authorMatch := range searchResult.TitleMatches {
				availableBook := dtos.AvailableBook{
					BookInfo:         searchResult.SearchBook,
					BookPurchaseInfo: authorMatch,
				}
				availableBooks = append(availableBooks, availableBook)
			}
		}

	}
	return availableBooks
}

func stripOfFormatting(input string) string {
	formatted := strings.ReplaceAll(input, "\n", "")
	formatted = strings.TrimSpace(formatted)
	return formatted
}

func extractTitleDetailsIfPossible(fullTitle string) (string, string) {
	splitFullTitle := TITLE_AND_SERIES_INFO_SEPERATOR.Split(fullTitle, 2)
	if len(splitFullTitle) == 2 {
		return splitFullTitle[0], splitFullTitle[1]
	}
	return fullTitle, ""
}

func extractLoadedCount(loadedCountText string) (int, int) {
	loadedCountText = strings.TrimSpace(loadedCountText)
	splitBySpace := strings.Split(loadedCountText, " ")
	if len(splitBySpace) == 4 {
		return strToInt(splitBySpace[0]), strToInt(splitBySpace[2])
	}
	logger.Sugar().Fatal(splitBySpace)
	return 0, 0
}

func strToInt(str string) int {
	intVersion, err := strconv.Atoi(str)
	if err != nil {
		logger.Sugar().Fatal(err)
	}
	return intVersion
}

func strToFloat(floatString string) float64 {
	floatVal, err := strconv.ParseFloat(floatString, 64)
	if err != nil {
		logger.Sugar().Fatalf("failed to parse floatString: %s", floatString)
	}
	return floatVal
}

func totalPagesToCrawl(totalBooks int) int {
	fullPages, nonFullPageIfMoreThanOne := divmod(totalBooks, BOOK_COUNT_PER_PAGE)
	if (nonFullPageIfMoreThanOne) >= 1 {
		return fullPages + 1
	}
	return fullPages
}

func divmod(big, little int) (int, int) {
	quotient := big / little
	remainder := big % little
	return quotient, remainder
}

func extractPureTitle(fullTitle string) string {
	title := fullTitle
	if strings.Contains(title, "(") {
		return strings.TrimSpace(title[:strings.Index(title, "(")])
	}
	return strings.TrimSpace(title)
}

func ensureAllAverageRatingsAreOfTypeString(jsonData []byte) []byte {
	stringJson := string(jsonData)
	stringJson = strings.ReplaceAll(stringJson, `avgRating":0.0,`, `avgRating":"0.0",`)
	stringJson = strings.ReplaceAll(stringJson, `avgRating":0,`, `avgRating":"0.0",`)
	return []byte(stringJson)
}

func getSeriesLink(seriesTitle string, htmlPage *html.Node) string {
	doc := goquery.NewDocumentFromNode(htmlPage)
	bookSeriesLink := ""
	doc.Find("h3.Text__italic > a").Each(func(i int, bookSeriesElem *goquery.Selection) {
		bookSeriesLink, _ := bookSeriesElem.Attr("href")
		fmt.Printf("Found a new series link: %s\n", bookSeriesLink)
	})
	if bookSeriesLink == "" {
		logger.Sugar().Panicf("failed to retrieve series link for %s", seriesTitle)
	}
	fmt.Printf("Returning the series link found from indiv page %s\n", bookSeriesLink)
	return bookSeriesLink
}

func extractSeriesTitleAndAuthorFromFullSeriesTitle(fullTitle string) (string, string) {
	splitTitle := strings.Split(fullTitle, " Series")
	if len(splitTitle) != 2 {
		logger.Sugar().Warnf("could not split this series title: '%s'", fullTitle)
		return fullTitle, ""
	}
	return strings.TrimSpace(splitTitle[0]), strings.TrimSpace(splitTitle[1])
}

func extractPrimaryAndTotalWorks(fullWorksText string) (int, int) {
	extractedNumbers := ONLY_NUMBERS.FindAllString(fullWorksText, 2)
	return strToInt(extractedNumbers[0]), strToInt(extractedNumbers[1])
}

func extractSeriesInfo(seriesPageLink string) dtos.Series {
	seriesInfo := dtos.Series{
		ID:   ksuid.New().String(),
		Link: seriesPageLink,
	}
	authorInMainTitle := false

	seriesPage := controller.Cnt.GetPage(seriesPageLink)

	doc := goquery.NewDocumentFromNode(seriesPage)
	doc.Find("div[class='responsiveSeriesHeader']").Each(func(i int, worksInfo *goquery.Selection) {
		worksInfo.Children().Each(func(i int, child *goquery.Selection) {
			switch i {
			case 0:
				seriesTitle, author := extractSeriesTitleAndAuthorFromFullSeriesTitle(child.Text())
				if author != "" {
					authorInMainTitle = true
					seriesInfo.Author = author
				}
				seriesInfo.Title = seriesTitle
			case 1:
				primaryWorks, totalWorks := extractPrimaryAndTotalWorks(child.Text())
				seriesInfo.PrimaryWorks = primaryWorks
				seriesInfo.TotalWorks = totalWorks
			default:
				break
			}

		})
	})

	currBookSeriesText := ""

	doc.Find("div[class='listWithDividers__item']").Each(func(i int, bookRow *goquery.Selection) {
		currBookInSeries := dtos.SeriesBook{}
		currBookInSeries.BookInfo.ID = ksuid.New().String()
		currBookInSeries.BookInfo.SeriesText = seriesInfo.Title
		bookRow.Find("h3").Each(func(k int, bookSeriesElement *goquery.Selection) {
			if k == 0 {
				if isLumpedTogetherBook := strings.HasPrefix(bookSeriesElement.Text(), "Shelve"); isLumpedTogetherBook {
					currBookInSeries.BookSeriesText = currBookSeriesText
				} else {
					currBookInSeries.BookSeriesText = bookSeriesElement.Text()
					currBookSeriesText = currBookInSeries.BookSeriesText
				}
				currBookInSeries.RealBookOrder = i + 1
			}
		})

		bookRow.Find("div[class='responsiveBook__media'] > a").Each(func(i int, linkElem *goquery.Selection) {
			link, _ := linkElem.Attr("href")
			currBookInSeries.BookInfo.Link = GOODREADS_BASE_BOOK_URL + link
		})
		bookRow.Find("div[class='responsiveBook__media'] > a > img").Each(func(i int, imgElem *goquery.Selection) {
			cover, _ := imgElem.Attr("src")
			currBookInSeries.BookInfo.Cover = cover
			title, _ := imgElem.Attr("alt")
			currBookInSeries.BookInfo.Title = title
		})
		bookRow.Find("img[class='responsiveBook__img']").Each(func(i int, imgElem *goquery.Selection) {
			cover, _ := imgElem.Attr("src")
			currBookInSeries.BookInfo.Cover = cover
		})
		bookRow.Find("span[itemprop='author'] > span[itemprop='name'] > a").Each(func(i int, authorLink *goquery.Selection) {
			currBookInSeries.BookInfo.Author = authorLink.Text()
		})
		bookRow.Find("div[class='communityRating']").Each(func(i int, communitRatingElems *goquery.Selection) {
			rating, publishedYear := extractCommunityRatingElementsFromText(communitRatingElems.Text())
			currBookInSeries.BookInfo.Rating = rating
			currBookInSeries.BookInfo.PublishedYear = publishedYear
		})

		seriesInfo.Books = append(seriesInfo.Books, currBookInSeries)
	})

	if !authorInMainTitle {
		// TODO get most common authors or all authors
		if len(seriesInfo.Books) == 0 {
			fmt.Printf("\n\nbad bad no books found for series im gonna fail\n\n")
			jsonOut, err := json.Marshal(seriesInfo)
			if err != nil {
				panic(err)
			}
			fmt.Println(string(jsonOut))
			panic("bad bad bad bad")
		}
		seriesInfo.Author = seriesInfo.Books[0].BookInfo.Author
	}

	return seriesInfo
}

func FilterSeriesTitleFromSeriesText(seriesText string) string {
	formatted := strings.ReplaceAll(seriesText, "(", "")
	formatted = strings.ReplaceAll(formatted, ")", "")
	formatted = strings.ReplaceAll(formatted, "#", "")
	formatted = strings.ReplaceAll(formatted, ",", "")
	formatted = strings.ReplaceAll(formatted, ".", "")
	formatted = NUMBER_MATCH.ReplaceAllString(formatted, "")
	return strings.TrimSpace(formatted)
}

func sleepIfLongerThanAllotedTimeSinceLastRequest() {
	logger.Sugar().Debugw(fmt.Sprintf("Time since last goodreads request was %+v Default is %+v", time.Since(lastRequestMade), SLEEP_DURATION),
		zap.String("dignostics", "goodReadsEngine"))
	if time.Since(lastRequestMade) > SLEEP_DURATION {
		lastRequestMade = time.Now()
		logger.Sugar().Debugw(fmt.Sprintf("[goodreads] Time since last request more than %d, not sleeping", SLEEP_DURATION),
			zap.String("dignostics", "goodReadsEngine"))
		return
	}
	timeDifference := SLEEP_DURATION - time.Since(lastRequestMade)
	logger.Sugar().Debugw(fmt.Sprintf("Time since last request was less than %d, sleeping for %+v", SLEEP_DURATION, timeDifference),
		zap.String("dignostics", "goodReadsEngine"))
	controller.Cnt.Sleep(timeDifference)
	lastRequestMade = time.Now()
}

func extractCommunityRatingElementsFromText(ratingRawText string) (float32, int) {
	splitRatings := strings.Split(ratingRawText, "·")
	rating := strToFloat(strings.TrimSpace(splitRatings[0]))
	publishYear := "0"
	for _, elem := range splitRatings {
		if strings.Contains(elem, "published") {
			publishYear = strings.TrimSpace(strings.Split(elem, " ")[2])
		}
	}

	return float32(rating), strToInt(publishYear)
}

func getFirstNBookCovers(books []dtos.BasicGoodReadsBook, n int) []string {
	bookCovers := []string{}

	for _, book := range books {
		bookCovers = append(bookCovers, book.Cover)
	}
	if len(bookCovers) >= n {
		return bookCovers[:n]
	}
	return bookCovers
}
