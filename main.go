package main

import (
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/BurntSushi/toml"

	"golang.org/x/text/language"

	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	. "github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/petergtz/alexa-twitter/locale"
	"github.com/petergtz/alexa-twitter/verbalizer"
	"github.com/petergtz/go-alexa"
)

func main() {
	l := createLoggerWith("debug")
	defer l.Sync()
	logger = l.Sugar()

	i18nBundle := i18n.NewBundle(language.English)
	i18nBundle.RegisterUnmarshalFunc("toml", toml.Unmarshal)

	i18nBundle.MustParseMessageFileBytes(locale.DeDe, "active.de.toml")
	i18nBundle.MustParseMessageFileBytes(locale.EnUs, "active.en.toml")

	var e error
	skipRequestValidation := false
	if os.Getenv("SKILL_SKIP_REQUEST_VALIDATION") != "" {
		skipRequestValidation, e = strconv.ParseBool(os.Getenv("SKILL_SKIP_REQUEST_VALIDATION"))
		if e != nil {
			logger.Fatalw("Invalid env var SKILL_SKIP_REQUEST_VALIDATION", "value", os.Getenv("SKILL_SKIP_REQUEST_VALIDATION"))
		}
		if skipRequestValidation {
			logger.Info("Skipping request validation. THIS SHOULD ONLY BE USED IN TESTING")
		}
	}

	if os.Getenv("APPLICATION_ID") == "" {
		logger.Fatal("env var APPLICATION_ID not provided.")
	}

	handler := &alexa.Handler{
		Skill: &Skill{
			i18nBundle:      i18nBundle,
			twitterProvider: NewTwitterProvider(os.Getenv("CONSUMER_KEY"), os.Getenv("CONSUMER_SECRET")),
		},
		Log:                   logger,
		ExpectedApplicationID: os.Getenv("APPLICATION_ID"),
		SkipRequestValidation: skipRequestValidation,
	}
	serveMux := http.NewServeMux()
	serveMux.HandleFunc("/", handler.Handle)

	port := os.Getenv("PORT")
	if port == "" {
		logger.Fatal("No env variable PORT specified")
	}
	addr := os.Getenv("SKILL_ADDR")
	if addr == "" {
		addr = "0.0.0.0"
		logger.Infow("No SKILL_ADDR provided. Using default.", "addr", addr)
	} else {
		logger.Infow("SKILL_ADDR provided.", "addr", addr)
	}

	httpServer := &http.Server{
		Handler:      serveMux,
		Addr:         addr + ":" + port,
		WriteTimeout: 60 * time.Minute,
		ReadTimeout:  60 * time.Minute,
		ErrorLog:     NewStdLog(l),
	}

	if os.Getenv("SKILL_USE_TLS") == "true" {
		logger.Infow("Starting webserver", "use-tls", true, "cert-path", os.Getenv("CERT"), "key-path", os.Getenv("KEY"), "port", port, "address", addr)
		e = httpServer.ListenAndServeTLS(os.Getenv("CERT"), os.Getenv("KEY"))
	} else {
		logger.Infow("Starting webserver", "use-tls", false, "port", port, "address", addr)
		e = httpServer.ListenAndServe()
	}
	logger.Fatal(e)
}

type TwitterVerbalizerProvider struct{ config *oauth1.Config }

func NewTwitterProvider(consumerKey, consumerSecret string) *TwitterVerbalizerProvider {
	return &TwitterVerbalizerProvider{config: oauth1.NewConfig(consumerKey, consumerSecret)}
}

func (tp *TwitterVerbalizerProvider) Get(accessTokenAndKey string) *verbalizer.Twitter {
	// TODO: do accessTokenAndKey input checking
	parts := strings.Split(accessTokenAndKey, ",")
	client := twitter.NewClient(tp.config.Client(oauth1.NoContext, oauth1.NewToken(parts[0], parts[1])))
	return verbalizer.NewTwitter(client.Timelines, client.Statuses)
}

type Skill struct {
	i18nBundle      *i18n.Bundle
	twitterProvider *TwitterVerbalizerProvider
}

func (h *Skill) ProcessRequest(requestEnv *alexa.RequestEnvelope) *alexa.ResponseEnvelope {
	logger.Infow("Request", "Type", requestEnv.Request.Type, "Intent", requestEnv.Request.Intent,
		"SessionAttributes", requestEnv.Session.Attributes, "locale", requestEnv.Request.Locale)

	l := i18n.NewLocalizer(h.i18nBundle, requestEnv.Request.Locale)

	if requestEnv.Session.User.AccessToken == "" {
		return &alexa.ResponseEnvelope{Version: "1.0",
			Response: &alexa.Response{
				OutputSpeech:     plainText("Bevor Du die neuesten Tweets aus Deiner Timeline hören kannst, verbinde bitte zuerst Alexa mit Deinem Twitter Account in der Alexa App."),
				Card:             &alexa.Card{Type: "LinkAccount"},
				ShouldSessionEnd: true,
			},
			SessionAttributes: requestEnv.Session.Attributes,
		}
	}

	switch requestEnv.Request.Type {

	case "LaunchRequest":
		timeline, e := h.twitterProvider.Get(requestEnv.Session.User.AccessToken).HomeTimeline()
		if e != nil {
			logger.Error(e)
			return internalError(l)
		}
		logger.Debug(timeline)
		return &alexa.ResponseEnvelope{Version: "1.0", Response: &alexa.Response{
			OutputSpeech: ssmlText("Hier sind die neuesten Tweets aus Deiner Timeline: " + timeline),
		}}

	case "IntentRequest":
		intent := requestEnv.Request.Intent
		switch intent.Name {
		case "TimelineIntent":
			timeline, e := h.twitterProvider.Get(requestEnv.Session.User.AccessToken).HomeTimeline()
			if e != nil {
				logger.Error(e)
				return internalError(l)
			}
			logger.Debug(timeline)
			return &alexa.ResponseEnvelope{Version: "1.0", Response: &alexa.Response{
				OutputSpeech: ssmlText(timeline),
			}}
		case "AMAZON.HelpIntent":
			return &alexa.ResponseEnvelope{Version: "1.0", Response: &alexa.Response{OutputSpeech: plainText(
				"Du benutzt gerade Tweety, einen Skill zur Interaktion mit Deinem Twitter Account. Sage z.B. \"Was gibt es neues\".")}}
		case "AMAZON.FallbackIntent":
			return &alexa.ResponseEnvelope{Version: "1.0",
				Response: &alexa.Response{OutputSpeech: plainText(
					"Tweety kann hiermit nicht weiterhelfen. Aber Du kannst z.B. sagen \"Was gibt es neues\"."),
				},
			}
		case "AMAZON.CancelIntent", "AMAZON.StopIntent":
			return &alexa.ResponseEnvelope{Version: "1.0",
				Response: &alexa.Response{ShouldSessionEnd: true},
			}
		default:
			return internalError(l)
		}

	case "SessionEndedRequest":
		return &alexa.ResponseEnvelope{Version: "1.0"}

	default:
		return &alexa.ResponseEnvelope{Version: "1.0"}
	}
}

func plainText(text string) *alexa.OutputSpeech {
	return &alexa.OutputSpeech{Type: "PlainText", Text: text}
}

func ssmlText(text string) *alexa.OutputSpeech {
	return &alexa.OutputSpeech{
		Type: "SSML",
		SSML: "<speak>" + text + "</speak>",
	}
}

func internalError(l *i18n.Localizer) *alexa.ResponseEnvelope {
	return &alexa.ResponseEnvelope{Version: "1.0",
		Response: &alexa.Response{
			OutputSpeech: plainText(l.MustLocalize(&LocalizeConfig{DefaultMessage: &Message{
				ID:    "InternalError",
				Other: "Es ist ein interner Fehler aufgetreten bei der Benutzung von Tweety.",
			}})),
			ShouldSessionEnd: false,
		},
	}
}
