package dtos

type BasicGoodReadsBook struct {
	ID            string  `json:"id"`
	Title         string  `json:"title"`
	Author        string  `json:"author"`
	SeriesText    string  `json:"seriesText"`
	Link          string  `json:"link"`
	Cover         string  `json:"cover"`
	Isbn13        string  `json:"isbn13"`
	Asin          string  `json:"asin"`
	Rating        float32 `json:"rating"`
	PublishedYear int     `json:"published"`

	// BooksBooksBooks specific meta data
	IsFromSeriesSearch bool `json:"isFromSeriesSearch"`
}

type GoodReadsSearchBookResult struct {
	ImageURL      string               `json:"imageUrl"`
	BookID        string               `json:"bookId"`
	WorkID        string               `json:"workId"`
	BookURL       string               `json:"bookUrl"`
	FromSearch    bool                 `json:"from_search"`
	FromSrp       bool                 `json:"from_srp"`
	Qid           string               `json:"qid"`
	Rank          int                  `json:"rank"`
	Title         string               `json:"title"`
	BookTitleBare string               `json:"bookTitleBare"`
	NumPages      int                  `json:"numPages"`
	AvgRating     string               `json:"avgRating"`
	RatingsCount  int                  `json:"ratingsCount"`
	Author        GoodReadsAuthor      `json:"author"`
	KcrPreviewURL string               `json:"kcrPreviewUrl"`
	Description   GoodReadsDescription `json:"description"`
}

type GoodReadsAuthor struct {
	ID                int    `json:"id"`
	Name              string `json:"name"`
	IsGoodreadsAuthor bool   `json:"isGoodreadsAuthor"`
	ProfileURL        string `json:"profileUrl"`
	WorksListURL      string `json:"worksListUrl"`
}

type GoodReadsDescription struct {
	HTML           string `json:"html"`
	Truncated      bool   `json:"truncated"`
	FullContentURL string `json:"fullContentUrl"`
}
