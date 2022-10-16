package goodreads

import "strings"

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}

func stripOfFormatting(input string) string {
	formatted := strings.ReplaceAll(input, "\n", "")
	// formatted = strings.ReplaceAll(formatted, "\t", "")
	return formatted
}
