package verbalizer_test

import (
	"encoding/json"
	"io/ioutil"

	"github.com/dghubble/go-twitter/twitter"
	. "github.com/onsi/ginkgo"
	"github.com/onsi/gomega/format"
	. "github.com/petergtz/alexa-twitter/verbalizer/matchers"
	. "github.com/petergtz/pegomock/ginkgo_compatible"

	. "github.com/onsi/gomega"
	"github.com/petergtz/alexa-twitter/verbalizer"
)

var _ = Describe("Verbalizer", func() {
	var (
		timelineClient *MockTwitterTimelineClient
		statusesClient *MockTwitterStatusesClient
		t              *verbalizer.Twitter
	)

	BeforeEach(func() {
		timelineClient = NewMockTwitterTimelineClient()
		statusesClient = NewMockTwitterStatusesClient()
		t = verbalizer.NewTwitter(timelineClient, statusesClient)

		format.TruncatedDiff = false
	})

	It("verbalizes a simple tweet with URLs and hashtags", func() {
		Whenever(statusesClient.Show(AnyInt64(), AnyPtrToTwitterStatusShowParams())).
			ThenReturn(tweet("tweet1.json"), nil, nil)

		Expect(t.Show(-1)).To(Equal(
			`<speak>Von Peter Goetz: <lang xml:lang="en-US">Just submitted my <break strength="medium"/>hashtag<break strength="medium"/>golang<break strength="medium"/> 2 error handling proposal ` +
				`medium.com, title: "Thinking About New Ways of Error Handling in Go 2 – Peter Goetz – Medium", to the official feedback wiki: ` +
				`github.com, title: "Go2ErrorHandlingFeedback   golang/go Wiki   GitHub",.</lang></speak>`))
	})

	It("verbalizes a retweet", func() {
		Whenever(statusesClient.Show(AnyInt64(), AnyPtrToTwitterStatusShowParams())).
			ThenReturn(tweet("tweet4.json"), nil, nil)
		Expect(t.Show(-1)).To(Equal(
			`<speak>Von AF SMC, retweeted von Elon Musk: ` +
				`<lang xml:lang="en-US">The 3700 kg Integrated Payload Stack (IPS) for <break strength="medium"/>hashtag<break strength="medium"/>STP2<break strength="medium"/> has been completed! Have a look before it blasts off on the first <break strength="medium"/>hashtag<break strength="medium"/>DoD<break strength="medium"/> Falcon Heavy launch! <break strength="medium"/>hashtag<break strength="medium"/>SMC<break strength="medium"/> <break strength="medium"/>hashtag<break strength="medium"/>SpaceStartsHere<break strength="medium"/> ` +
				`</lang><break strength="strong"/>Angehängt zum Tweet ist ein Bild.</speak>`))
	})
})

func tweet(filename string) *twitter.Tweet {
	b, e := ioutil.ReadFile(filename)
	Expect(e).NotTo(HaveOccurred())
	var tweet twitter.Tweet
	e = json.Unmarshal(b, &tweet)
	Expect(e).NotTo(HaveOccurred())
	return &tweet
}
