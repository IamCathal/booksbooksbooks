package thebookshop

import (
	"sync"
	"time"

	"github.com/iamcathal/booksbooksbooks/dtos"
)

var (
	activeSearchLock      sync.Mutex
	lastSearchTime        time.Time
	TIME_BETWEEN_SEARCHES = time.Duration(900 * time.Millisecond)
)

func init() {
	lastSearchTime = time.Now()
}

// func searchForBookWithThrottling(bookInfo dtos.BasicGoodReadsBook) []dtos.TheBookshopBook {
// 	activeSearchLock.Lock()
// 	defer activeSearchLock.Unlock()
// 	for {
// 		if time.Since(lastSearchTime) > TIME_BETWEEN_SEARCHES {
// 			fmt.Printf("Last search was %v ago\n", time.Since(lastSearchTime))
// 			lastSearchTime = time.Now()
// 			return searchTheBookshop(bookInfo)
// 		}
// 	}
// }

func asyncSearch(booksInfo []dtos.BasicGoodReadsBook) {

}
