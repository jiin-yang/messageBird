package webhook

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/rs/zerolog/log"
	"golang.org/x/net/html"
	"io"
	"net/http"
	"strings"
)

type Client interface {
	SendMessage(message SendMessageRequest) (*SendMessageResponseFromWebhook, error)
}

type client struct {
	url string
}

type NewClientOptions struct {
	URL string
}

func NewWebhookClient(opts *NewClientOptions) Client {
	return &client{url: opts.URL}
}

func (c client) SendMessage(requestMsg SendMessageRequest) (*SendMessageResponseFromWebhook, error) {
	requestBody := SendMessageRequest{
		To:      requestMsg.To,
		Content: requestMsg.Content,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		log.Error().
			Err(err).
			Str("method", "SendMessage-Webhook Client").
			Str("to", requestMsg.To).
			Msg("Failed to marshal request body to JSON")
		return nil, err
	}

	req, err := http.NewRequest("POST", c.url, bytes.NewBuffer(jsonData))
	if err != nil {
		log.Error().
			Err(err).
			Str("method", "SendMessage-Webhook Client").
			Str("url", c.url).
			Msg("Failed to create HTTP request")
		return nil, err
	}

	httpClient := &http.Client{}
	resp, err := httpClient.Do(req)
	if err != nil {
		log.Error().
			Err(err).
			Str("method", "SendMessage-Webhook Client").
			Str("url", c.url).
			Msg("Failed to send HTTP request")
		return nil, err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Warn().
				Err(err).
				Str("method", "SendMessage-Webhook Client").
				Str("url", c.url).
				Msg("Failed to close response body")
		}
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Error().
			Err(err).
			Str("method", "SendMessage-Webhook Client").
			Str("url", c.url).
			Msg("Failed to read response body")
		return nil, err
	}

	if resp.StatusCode >= 400 {
		log.Warn().
			Int("status_code", resp.StatusCode).
			Msg("Received non-success HTTP response")

		if isHTMLResponse(resp.Header.Get("Content-Type")) {
			htmlTitle, parseErr := extractHTMLTitle(body)
			if parseErr != nil {
				log.Error().
					Err(parseErr).
					Msg("Failed to extract error message from HTML response")
			} else {
				return nil, errors.New(htmlTitle)
			}
		}

		return nil, errors.New("HTTP error occurred with non-HTML response")
	}

	var response SendMessageResponseFromWebhook
	if err = json.Unmarshal(body, &response); err != nil {
		log.Error().
			Err(err).
			Str("method", "SendMessage-Webhook Client").
			Str("url", c.url).
			RawJSON("response_body", body).
			Msg("Failed to parse response JSON")
		return nil, err
	}

	log.Info().
		Str("method", "SendMessage-Webhook Client").
		Str("to", requestMsg.To).
		Int("status_code", resp.StatusCode).
		Msg("Message sent successfully")

	return &response, nil
}

func isHTMLResponse(contentType string) bool {
	return strings.Contains(contentType, "text/html")
}

func extractHTMLTitle(body []byte) (string, error) {
	doc, err := html.Parse(bytes.NewReader(body))
	if err != nil {
		return "", err
	}

	var traverse func(*html.Node) string
	traverse = func(n *html.Node) string {
		if n.Type == html.ElementNode && n.Data == "title" && n.FirstChild != nil {
			return n.FirstChild.Data
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			title := traverse(c)
			if title != "" {
				return title
			}
		}
		return ""
	}

	title := traverse(doc)
	if title == "" {
		return "", errors.New("no <title> element found in HTML response")
	}
	return title, nil
}
