package dghubble_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Dghubble", func() {
	Describe("Statuses.Show", func() {
		It("works", func() {
			config := oauth1.NewConfig(os.Getenv("CONSUMER_KEY"), os.Getenv("CONSUMER_SECRET"))
			twitterClient := twitter.NewClient(config.Client(oauth1.NoContext, oauth1.NewToken(os.Getenv("ACCESS_TOKEN"), os.Getenv("ACCESS_TOKEN_SECRET"))))

			tweet, resp, e := twitterClient.Statuses.Show(1062290266336436224, &twitter.StatusShowParams{TweetMode: "extended"})
			Expect(resp.StatusCode, e).To(Equal(http.StatusOK))
			Expect(tweet.FullText).To(Equal("Just submitted my #golang 2 error handling proposal https://t.co/sinLcoRIaL to the official feedback wiki: https://t.co/pCiqPm9S17."))
			printTweet(tweet)

			tweet, resp, e = twitterClient.Statuses.Show(1141055529843929091, &twitter.StatusShowParams{TweetMode: "extended"})
			Expect(resp.StatusCode, e).To(Equal(http.StatusOK))
			printTweet(tweet)

			tweet, resp, e = twitterClient.Statuses.Show(1141054757173252096, &twitter.StatusShowParams{
				IncludeMyRetweet: twitter.Bool(true),
			})
			Expect(resp.StatusCode, e).To(Equal(http.StatusOK))
			printTweet(tweet)
		})
	})

	Describe("Timelines.HomeTimeline", func() {
		FIt("works", func() {
			config := oauth1.NewConfig(os.Getenv("CONSUMER_KEY"), os.Getenv("CONSUMER_SECRET"))
			twitterClient := twitter.NewClient(config.Client(oauth1.NoContext, oauth1.NewToken(os.Getenv("ACCESS_TOKEN"), os.Getenv("ACCESS_TOKEN_SECRET"))))

			tweets, resp, e := twitterClient.Timelines.HomeTimeline(&twitter.HomeTimelineParams{
				TweetMode: "extended",
			})
			Expect(resp.StatusCode, e).To(Equal(http.StatusOK))
			for _, tweet := range tweets {
				printTweet(&tweet)
			}
		})
	})

})

func printTweet(tweet *twitter.Tweet) {
	b, e := json.MarshalIndent(tweet, "", "  ")
	Expect(e).NotTo(HaveOccurred())
	fmt.Println(string(b))
}
