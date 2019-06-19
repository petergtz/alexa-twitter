package verbalizer

import (
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

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
	verbalized, e := verbalizedTweet(tweet)
	if e != nil {
		return "", e
	}
	return `<speak>` + verbalized + `</speak>`, nil
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
	for _, tweet := range tweets {
		verbalized, e := verbalizedTweet(&tweet)
		if e != nil {
			return "", e
		}
		timeline += verbalized + `<break strength="x-strong"/>`
	}
	return "<speak>" + timeline + "</speak>", nil
}

func verbalizedTweet(tweet *twitter.Tweet) (string, error) {
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
		if mediaEntity.Indices.End() == len(text) {
			text = strings.Replace(text, mediaEntity.URL, "", 1)
			attachment = `<break strength="strong"/>Angehängt zum Tweet ist ein Bild.`
		}
	}

	for _, u := range tweet.Entities.Urls {
		parsedURL, e := url.Parse(u.ExpandedURL)
		if e != nil {
			// TODO: logger.Error(e)
			continue
		}

		resp, e := http.Get(u.ExpandedURL)
		if e != nil {
			// TODO: logger.Error(e)
			continue
		}
		c, e := ioutil.ReadAll(resp.Body)
		resp.Body.Close()
		title := strings.ReplaceAll(getTitle(string(c)), "·", " ")
		switch tweet.Lang {
		case "en":
			text = strings.Replace(text, u.URL, parsedURL.Hostname()+", title: \""+title+"\",", 1)
		case "de":
			text = strings.Replace(text, u.URL, parsedURL.Hostname()+", titel: \""+title+"\",", 1)
		}
	}
	for _, hashtag := range tweet.Entities.Hashtags {
		text = strings.Replace(text, "#"+hashtag.Text, `<break strength="medium"/>hashtag<break strength="medium"/>`+hashtag.Text+`<break strength="medium"/>`, 1)
	}
	switch tweet.Lang {
	case "en":
		text = `<lang xml:lang="en-US">` + text + "</lang>"
	case "de":
		text = `<lang xml:lang="de-DE">` + text + "</lang>"
	}

	if retweeted {
		return "Von " + tweet.User.Name + ", retweeted von " + retweeter + ": " + text + attachment, nil
	}
	return "Von " + tweet.User.Name + ": " + text, nil
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
