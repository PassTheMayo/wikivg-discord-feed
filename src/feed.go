package main

import (
	"log"
	"sort"
	"time"

	"github.com/mmcdole/gofeed"
)

var (
	fp *gofeed.Parser = gofeed.NewParser()
)

func ProcessAllFeedGoroutine() {
	for {
		go ProcessFeed()

		time.Sleep(config.CheckInterval)
	}
}

func ProcessFeed() {
	recentChangesData, err := FetchRecentChanges()

	if err != nil {
		log.Println(err)

		return
	}

	feed, err := fp.ParseString(string(recentChangesData))

	if err != nil {
		log.Println(err)

		return
	}

	sort.SliceStable(feed.Items, func(i, j int) bool {
		if feed.Items[i].PublishedParsed == nil || feed.Items[j].PublishedParsed == nil {
			return false
		}

		return feed.Items[i].PublishedParsed.Before(*feed.Items[j].PublishedParsed)
	})

	for _, item := range feed.Items {
		oldID, newID, err := ParseEditDetails(item.Link)

		if err != nil {
			log.Println(err)

			continue
		}

		edit, err := FetchEditComparison(oldID, newID)

		if err != nil {
			log.Println(err)

			continue
		}

		if err = SendEditWebhook(edit, item, oldID, newID); err != nil {
			log.Println(err)

			continue
		}
	}

	if err := WriteLastFetchTimestamp(lastCheck); err != nil {
		log.Fatal(err)
	}
}
