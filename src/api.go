package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

type DiscordMessage struct {
	Content string          `json:"content,omitempty"`
	Embeds  []*DiscordEmbed `json:"embeds"`
}

type DiscordEmbed struct {
	Title       string              `json:"title,omitempty"`
	Description string              `json:"description,omitempty"`
	URL         string              `json:"url,omitempty"`
	Timestamp   string              `json:"timestamp,omitempty"`
	Color       uint32              `json:"color,omitempty"`
	Footer      *DiscordEmbedFooter `json:"footer,omitempty"`
	Author      *DiscordEmbedAuthor `json:"author,omitempty"`
}

type DiscordEmbedFooter struct {
	Text string `json:"text"`
}

type DiscordEmbedAuthor struct {
	Name string `json:"name"`
	URL  string `json:"url,omitempty"`
}

type EditComparison struct {
	Compare struct {
		FromTitle string `json:"fromtitle"`
		ToTitle   string `json:"totitle"`
		ToComment string `json:"tocomment"`
		FromSize  int    `json:"fromsize"`
		ToSize    int    `json:"tosize"`
		ToUser    string `json:"touser"`
	}
}

func FetchRecentChanges() ([]byte, error) {
	feedURL, err := GetFeedURL()

	if err != nil {
		return nil, err
	}

	return Get(feedURL)
}

func FetchEditComparison(oldID, newID int) (*EditComparison, error) {
	comparisonURL, err := GetPageComparisonURL(oldID, newID)

	if err != nil {
		return nil, err
	}

	data, err := Get(comparisonURL)

	if err != nil {
		return nil, err
	}

	var result EditComparison

	if err = json.Unmarshal(data, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

func PostWebhook(webhookURL string, data *DiscordMessage) error {
	parsedURL, err := url.Parse(webhookURL)

	if err != nil {
		return err
	}

	parsedURL.RawQuery = "wait=true"

	body, err := json.Marshal(data)

	if err != nil {
		return err
	}

	resp, err := http.Post(parsedURL.String(), "application/json", bytes.NewReader(body))

	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("discord: unexpected status code: %d", resp.StatusCode)
	}

	return nil
}
