package main

import (
	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

var (
	consumerKey       = getenv("TWITTER_CONSUMER_KEY")
	consumerSecret    = getenv("TWITTER_CONSUMER_SECRET")
	accessToken       = getenv("TWITTER_ACCESS_TOKEN")
	accessTokenSecret = getenv("TWITTER_ACCESS_TOKEN_SECRET")
)

func getenv(name string) string {
	v := os.Getenv(name)
	if v == "" {
		log.Fatalf("missing required environment variable %s", name)
	}
	return v
}

func main() {
	// set up http client with oauth1
	config := oauth1.NewConfig(consumerKey, consumerSecret)
	token := oauth1.NewToken(accessToken, accessTokenSecret)
	client := twitter.NewClient(config.Client(oauth1.NoContext, token))

	// open a filtered stream
	params := &twitter.StreamFilterParams{
		Track:         []string{"wahoowa", "gohoos"},
		StallWarnings: twitter.Bool(true),
	}

	stream, err := client.Streams.Filter(params)
	defer stream.Stop()

	if err != nil {
		log.Fatalf("could not open stream %s", err)
	}

	// all we care about are tweets
	demux := twitter.NewSwitchDemux()
	demux.Tweet = func(tweet *twitter.Tweet) {
		if text := tweet.Text; !strings.Contains(text, "RT") {
			_, _, err = client.Statuses.Retweet(tweet.ID, nil)
			if err != nil {
				log.Printf("could not retweet %d", tweet.ID)
			} else {
				log.Printf("retweeted: %s", text)
			}
		}
	}

	// launch a separate goroutine forever
	go demux.HandleChan(stream.Messages)

	// wait for SIGINT and SIGTERM
	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	log.Println(<-ch)
}
