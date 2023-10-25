package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/mmcdole/gofeed"
)

const (
	BaseAPI     string = "https://wiki.vg/api.php"
	NamespaceID int    = 0
)

var (
	SectionReplacementRegExp = regexp.MustCompile(`^\/\* ?(.*) ?\*\/ ?(.*)$`)
)

func GetFeedURL() (string, error) {
	query := &url.Values{}
	query.Set("action", "feedrecentchanges")
	query.Set("feedformat", "rss")
	query.Set("from", strconv.FormatInt(lastCheck.Unix(), 10))
	query.Set("namespace", strconv.FormatInt(int64(NamespaceID), 10))

	requestURL, err := url.Parse(BaseAPI)

	if err != nil {
		return "", err
	}

	requestURL.RawQuery = query.Encode()

	return requestURL.String(), nil
}

func GetPageComparisonURL(oldID, newID int) (string, error) {
	query := &url.Values{}
	query.Set("action", "compare")
	query.Set("format", "json")
	query.Set("fromrev", strconv.FormatInt(int64(oldID), 10))
	query.Set("torev", strconv.FormatInt(int64(newID), 10))
	query.Set("prop", "title|user|comment|size")

	requestURL, err := url.Parse(BaseAPI)

	if err != nil {
		return "", err
	}

	requestURL.RawQuery = query.Encode()

	return requestURL.String(), nil
}

func ParseEditDetails(link string) (int, int, error) {
	parsedURL, err := url.Parse(link)

	if err != nil {
		return 0, 0, err
	}

	q := parsedURL.Query()

	if !q.Has("oldid") || !q.Has("diff") {
		return 0, 0, fmt.Errorf("%s: missing oldid or diff query", link)
	}

	oldID, err := strconv.ParseInt(q.Get("oldid"), 10, 64)

	if err != nil {
		return 0, 0, err
	}

	diffID, err := strconv.ParseInt(q.Get("diff"), 10, 64)

	if err != nil {
		return 0, 0, err
	}

	return int(oldID), int(diffID), nil
}

func ReadLastFetchTimestamp() time.Time {
	now := time.Now()

	data, err := os.ReadFile("timestamp.txt")

	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return now
		}

		log.Fatal(err)
	}

	value, err := strconv.ParseInt(string(data), 10, 64)

	if err != nil {
		log.Println(err)

		return now
	}

	t := time.Unix(value, 0)

	if t.Compare(time.Now()) > 0 {
		return now
	}

	return t
}

func WriteLastFetchTimestamp(value time.Time) error {
	lastCheck = time.Now()

	return os.WriteFile("timestamp.txt", []byte(strconv.FormatInt(value.Unix(), 10)), 0777)
}

func SendEditWebhook(edit *EditComparison, item *gofeed.Item, oldID, newID int) error {
	comment := edit.Compare.ToComment

	if SectionReplacementRegExp.MatchString(comment) {
		matches := SectionReplacementRegExp.FindAllStringSubmatch(comment, -1)

		if len(matches) > 0 && len(matches[0]) > 2 {
			if len(matches[0]) == 2 || len(strings.Trim(matches[0][2], " ")) < 1 {
				comment = matches[0][1]
			} else {
				comment = fmt.Sprintf("%s: %s", strings.Trim(matches[0][1], " "), strings.Trim(matches[0][2], " "))
			}
		}
	}

	return PostWebhook(config.WebhookURL, &DiscordMessage{
		Embeds: []*DiscordEmbed{
			{
				Title:       fmt.Sprintf("Page '%s' edited", edit.Compare.ToTitle),
				Description: comment,
				URL:         fmt.Sprintf("https://wiki.vg/index.php?title=%s&diff=%d&oldid=%d", url.QueryEscape(edit.Compare.ToTitle), newID, oldID),
				Color:       3066993,
				Author: &DiscordEmbedAuthor{
					Name: edit.Compare.ToUser,
					URL:  fmt.Sprintf("https://wiki.vg/User:%s", edit.Compare.ToUser),
				},
				Footer: &DiscordEmbedFooter{
					Text: fmt.Sprintf("%+d bytes", edit.Compare.ToSize-edit.Compare.FromSize),
				},
				Timestamp: item.PublishedParsed.Format(time.RFC3339),
			},
		},
	})
}

func Get(url string) ([]byte, error) {
	resp, err := http.Get(url)

	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("rest: unexpected status code: %d", resp.StatusCode)
	}

	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}
