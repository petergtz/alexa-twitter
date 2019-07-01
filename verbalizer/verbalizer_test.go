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
			`Von Peter Goetz: <lang xml:lang="en-US"><voice name="Kendra">Just submitted my <break strength="medium"/>hashtag<break strength="medium"/>golang<break strength="medium"/> 2 error handling proposal ` +
				`medium.com, title: "Thinking About New Ways of Error Handling in Go 2 - Peter Goetz - Medium", to the official feedback wiki: ` +
				`github.com, title: "Go2ErrorHandlingFeedback   golang/go Wiki   GitHub",.</voice></lang>`))
	})

	It("verbalizes a simple tweet with URLs and hashtags", func() {
		Whenever(statusesClient.Show(AnyInt64(), AnyPtrToTwitterStatusShowParams())).
			ThenReturn(tweet("tweet3.json"), nil, nil)

		Expect(t.Show(-1)).To(Equal(
			`Von Danny Trinh: <lang xml:lang="en-US"><voice name="Kendra">Tuesday energy: from Gentry Underwood on Navigator. As you add more people to a team, be mindful of all of the new relationships to foster:</voice></lang><break strength="strong"/>Angehängt zum Tweet ist ein Bild. `))
	})

	It("verbalizes a retweet", func() {
		Whenever(statusesClient.Show(AnyInt64(), AnyPtrToTwitterStatusShowParams())).
			ThenReturn(tweet("tweet4.json"), nil, nil)
		Expect(t.Show(-1)).To(Equal(
			`Von AF SMC, retweeted von Elon Musk: ` +
				`<lang xml:lang="en-US"><voice name="Kendra">The 3700 kg Integrated Payload Stack (IPS) for <break strength="medium"/>hashtag<break strength="medium"/>STP2<break strength="medium"/> has been completed! Have a look before it blasts off on the first <break strength="medium"/>hashtag<break strength="medium"/>DoD<break strength="medium"/> Falcon Heavy launch! <break strength="medium"/>hashtag<break strength="medium"/>SMC<break strength="medium"/> <break strength="medium"/>hashtag<break strength="medium"/>SpaceStartsHere<break strength="medium"/>` +
				`</voice></lang><break strength="strong"/>Angehängt zum Tweet ist ein Bild. `))
	})

	It("verbalizes a quoted tweet", func() {
		Whenever(statusesClient.Show(AnyInt64(), AnyPtrToTwitterStatusShowParams())).
			ThenReturn(tweet("tweet5.json"), nil, nil)
		Expect(t.Show(-1)).To(Equal(
			`Von AukeHoekstra: ` +
				`<lang xml:lang="en-US"><voice name="Kendra">This one tender equals:<break strength="strong"/><break strength="strong"/>all global sales in 2006<break strength="strong"/><break strength="strong"/>10x global sales in 1999 ` +
				`</voice></lang><break strength="strong"/>Zitiert damit den Tweet Von Bloomberg Technology: <lang xml:lang="en-US"><voice name="Kendra">India issued a new tender for solar power equipment manufacturing capacity totaling 2 gigawatts</voice></lang><break strength="strong"/>Angehängt zum Tweet ist ein Link zu bloom.bg, titel: "India Issues New Tender for Solar Power Equipment Factories - Bloomberg". `))
	})

	It("verbalizes a tweet with 2 attachments", func() {
		Whenever(statusesClient.Show(AnyInt64(), AnyPtrToTwitterStatusShowParams())).
			ThenReturn(tweet("tweet7.json"), nil, nil)
		Expect(t.Show(-1)).To(Equal(
			`Von Jan Böhmermann 🤨: <lang xml:lang="de-DE">In von 24 Stunden haben wir gemeinsam 618.216 EURO für <break strength="medium"/>hashtag<break strength="medium"/>FreeCarolaRackete<break strength="medium"/> und die private Seenotrettung gesammelt.<break strength="strong"/><break strength="strong"/>Das ist nicht nur dringend benötigtes Geld, sondern auch ein Signal – an die Lebensretter ✊️ und die politisch Verantwortlichen 🗽❤️😘<break strength="strong"/><break strength="strong"/></lang><break strength="strong"/>Angehängt zum Tweet ist ein Bild und ein Link zu www.leetchi.com, titel: "Pool : Leben retten ist kein Verbrechen: Lasst uns die Seenotretter retten! - Leetchi.com". `))
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
