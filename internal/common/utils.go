package common

import (
	"regexp"
)

func ExtractYear(title string) string {
	re := regexp.MustCompile(`\b(19\d{2}|20\d{2})\b`)
	matches := re.FindStringSubmatch(title)
	if len(matches) > 0 {
		return matches[0]
	}
	return ""
}

func RemoveYearFromTitle(title string) string {
	re := regexp.MustCompile(`\s+(19\d{2}|20\d{2})$`)
	return re.ReplaceAllString(title, "")
}
