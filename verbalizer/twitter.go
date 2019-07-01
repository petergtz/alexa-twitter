package verbalizer

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"encoding/xml"

	"github.com/dghubble/go-twitter/twitter"
	"golang.org/x/net/html"
)

type Twitter struct {
	twitterTimelineClient TwitterTimelineClient
	twitterStatusesClient TwitterStatusesClient
}

type TwitterTimelineClient interface {
	HomeTimeline(params *twitter.HomeTimelineParams) ([]twitter.Tweet, *http.Response, error)
}

type TwitterStatusesClient interface {
	Show(id int64, params *twitter.StatusShowParams) (*twitter.Tweet, *http.Response, error)
}

func NewTwitter(twitterTimelineClient TwitterTimelineClient,
	twitterStatusesClient TwitterStatusesClient) *Twitter {
	return &Twitter{
		twitterTimelineClient: twitterTimelineClient,
		twitterStatusesClient: twitterStatusesClient,
	}
}

func (t *Twitter) Show(id int64) (string, error) {
	tweet, _, e := t.twitterStatusesClient.Show(id, &twitter.StatusShowParams{TweetMode: "extended"})
	if e != nil {
		return "", e
	}
	verbalized, e := t.verbalizedTweet(tweet)
	if e != nil {
		return "", e
	}
	return verbalized, nil
}

func (t *Twitter) HomeTimeline() (string, error) {
	tweets, _, e := t.twitterTimelineClient.HomeTimeline(&twitter.HomeTimelineParams{
		Count:     10,
		TweetMode: "extended",
	})
	if e != nil {
		return "", e
	}
	timeline := ""
	var errs []error
	for _, tweet := range tweets {
		verbalized, e := t.verbalizedTweet(&tweet)
		if e != nil {
			errs = append(errs, e)
			// TODO log issue
			continue
		}
		timeline += verbalized + `<break time="1300ms"/>`
	}
	if timeline == "" {
		return "", fmt.Errorf("One or more errors while verbalizing tweets from home timeline: %#v", errs)
	}
	return timeline, nil
}

func (t *Twitter) verbalizedTweet(tweet *twitter.Tweet) (string, error) {

	retweeted := false
	var retweeter string
	if tweet.RetweetedStatus != nil {
		retweeter = tweet.User.Name
		tweet = tweet.RetweetedStatus
		retweeted = true
	}
	text := tweet.FullText

	attachment := ""
	for _, mediaEntity := range tweet.Entities.Media {
		text = strings.TrimRight(strings.Replace(text, mediaEntity.URL, "", 1), " .,")
	}
	if len(tweet.Entities.Media) == 1 {
		attachment += `<break strength="strong"/>Angehängt zum Tweet ist ein Bild`
	} else if len(tweet.Entities.Media) > 1 {
		attachment += `<break strength="strong"/>Angehängt zum Tweet sind mehrere Bilder`
	}
	text = xmlEscapeText(text)

	for _, user := range tweet.Entities.UserMentions {
		text = strings.ReplaceAll(text, "@"+user.ScreenName, xmlEscapeText(user.Name))
	}

	for _, u := range tweet.Entities.Urls {
		parsedURL, e := url.Parse(u.ExpandedURL)
		if e != nil {
			// TODO: logger.Error(e)
			continue
		}

		if parsedURL.Hostname() == "twitter.com" {
			fragments := regexp.MustCompile(`/.*/status/(\d+)`).FindStringSubmatch(parsedURL.Path)
			if len(fragments) == 2 {
				tweetID, e := strconv.ParseInt(fragments[1], 10, 64)
				if e != nil {
					// TODO: logger.Error(e)
					continue
				}
				if tweetID == tweet.QuotedStatusID {
					text = strings.ReplaceAll(text, u.URL, "")
					continue
				}
				referencedTweet, resp, e := t.twitterStatusesClient.Show(tweetID, &twitter.StatusShowParams{TweetMode: "extended"})
				if resp.StatusCode != http.StatusOK {
					// TODO: logger.Error(e)
					continue
				}

				verbalizedTweet, e := t.verbalizedTweet(referencedTweet)
				if e != nil {
					// TODO: logger.Error(e)
					continue
				}

				text += "Tweet: " + verbalizedTweet
			}
		}

		r, e := http.NewRequest("GET", u.ExpandedURL, nil)
		if e != nil {
			panic(e)
		}
		r.Header.Set("accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8")
		r.Header.Set("accept-language", "en-US,en;q=0.9,fr;q=0.8,ro;q=0.7,ru;q=0.6,la;q=0.5,pt;q=0.4,de;q=0.3")
		r.Header.Set("cache-control", "max-age=0")
		r.Header.Set("upgrade-insecure-requests", "1")
		r.Header.Set("user-agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_13_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/67.0.3396.99 Safari/537.36")
		resp, e := http.DefaultClient.Do(r)
		if e != nil {
			// TODO: logger.Error(e)
			fmt.Println("ERROR:", e)
			continue
		}
		c, e := ioutil.ReadAll(resp.Body)
		if e != nil {
			panic(e)
		}
		resp.Body.Close()
		title := xmlEscapeText(strings.ReplaceAll(getTitle(string(c)), "·", " "))
		if strings.Index(text, u.URL)+len(u.URL) == len(text) {
			attachment = attach(attachment, "Link zu "+parsedURL.Hostname()+", titel: \""+title+"\"")
			text = strings.TrimRight(strings.Replace(text, u.URL, "", 1), " ,.")
		} else {
			switch tweet.Lang {
			case "en":
				text = strings.Replace(text, u.URL, parsedURL.Hostname()+", title: \""+title+"\",", 1)
			case "de":
				text = strings.Replace(text, u.URL, parsedURL.Hostname()+", titel: \""+title+"\",", 1)
			}
		}
	}
	for _, hashtag := range tweet.Entities.Hashtags {
		text = strings.Replace(text, "#"+hashtag.Text, `<break strength="medium"/>hashtag<break strength="medium"/>`+xmlEscapeText(hashtag.Text)+`<break strength="medium"/>`, 1)
	}
	switch tweet.Lang {
	case "en":
		text = `<lang xml:lang="en-US"><voice name="Kendra">` + text + "</voice></lang>"
	case "de":
		text = `<lang xml:lang="de-DE">` + text + "</lang>"
	}

	if attachment != "" {
		attachment += ". "
	}
	if tweet.QuotedStatus != nil {
		verbalizedQuotedTweet, e := t.verbalizedTweet(tweet.QuotedStatus)
		if e != nil {
			return "", errors.New("Could not verbalize quoted tweet: " + e.Error())
		}
		return "Von " + xmlEscapeText(tweet.User.Name) + ": " + text + attachment + `<break strength="strong"/>Zitiert damit den Tweet ` + verbalizedQuotedTweet, nil
	}
	if retweeted {
		return "Von " + xmlEscapeText(tweet.User.Name) + ", retweeted von " + xmlEscapeText(retweeter) + ": " + text + attachment, nil
	}
	return "Von " + xmlEscapeText(tweet.User.Name) + ": " + text + attachment, nil
}

func attach(a, b string) string {
	if a != "" {
		return a + " und ein " + b
	} else {
		return "<break strength=\"strong\"/>Angehängt zum Tweet ist ein " + b
	}
}

func xmlEscapeText(text string) string {
	var tempBuffer bytes.Buffer
	xml.EscapeText(&tempBuffer, []byte(text))
	return strings.ReplaceAll(tempBuffer.String(), `&#xA;`, `<break strength="strong"/>`)
}

func getTitle(HTMLString string) (title string) {
	r := strings.NewReader(HTMLString)
	z := html.NewTokenizer(r)

	var i int
	for {
		tt := z.Next()

		i++
		if i > 100 { // Title should be one of the first tags
			return
		}

		switch {
		case tt == html.ErrorToken:
			// End of the document, we're done
			return
		case tt == html.StartTagToken:
			t := z.Token()

			// Check if the token is an <title> tag
			if t.Data != "title" {
				continue
			}

			// fmt.Printf("%+v\n%v\n%v\n%v\n", t, t, t.Type.String(), t.Attr)
			tt := z.Next()

			if tt == html.TextToken {
				t := z.Token()
				title = t.Data
				return
				// fmt.Printf("%+v\n%v\n", t, t.Data)
			}
		}
	}
}
