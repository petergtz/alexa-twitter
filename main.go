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

const helpText = "Um einen Artikel vorgelesen zu bekommen, " +
	"sage z.B. \"Suche nach Käsekuchen.\" oder \"Was ist Käsekuchen?\". " +
	"Du kannst jederzeit zum Inhaltsverzeichnis springen, indem Du \"Inhaltsverzeichnis\" sagst. " +
	"Oder sage \"Springe zu Abschnitt 3.2\", um direkt zu diesem Abschnitt zu springen."

const quickHelpText = "Suche zunächst nach einem Begriff. " +
	"Sage z.B. \"Suche nach Käsekuchen.\" oder \"Was ist Käsekuchen?\"."

func (h *Skill) ProcessRequest(requestEnv *alexa.RequestEnvelope) *alexa.ResponseEnvelope {
	logger.Infow("Request", "Type", requestEnv.Request.Type, "Intent", requestEnv.Request.Intent,
		"SessionAttributes", requestEnv.Session.Attributes, "locale", requestEnv.Request.Locale)

	l := i18n.NewLocalizer(h.i18nBundle, requestEnv.Request.Locale)

	switch requestEnv.Request.Type {

	case "LaunchRequest":
		return &alexa.ResponseEnvelope{Version: "1.0", Response: &alexa.Response{OutputSpeech: plainText("Und los geht's mit dem Gezwitscher! Sage z.B. \"was gibt's neues?\" um die neusten Tweets aus Deiner Twitter Home Timeline zu hören.")}}

	case "IntentRequest":
		intent := requestEnv.Request.Intent
		switch intent.Name {
		case "TimelineIntent":
			timeline, e := h.twitterProvider.Get(requestEnv.Session.User.AccessToken).HomeTimeline()
			if e != nil {
				logger.Error(e)
				return internalError(l)
			}
			return &alexa.ResponseEnvelope{Version: "1.0", Response: &alexa.Response{
				OutputSpeech: ssmlText(timeline),
			}}
		case "AMAZON.YesIntent",
			"AMAZON.ResumeIntent",
			"AMAZON.RepeatIntent",
			"AMAZON.NextIntent",
			"AMAZON.NoIntent",
			"AMAZON.PauseIntent",
			"AMAZON.HelpIntent",
			"AMAZON.FallbackIntent":
			return &alexa.ResponseEnvelope{Version: "1.0", Response: &alexa.Response{OutputSpeech: plainText("nicht implementiert")}}
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
	return &alexa.ResponseEnvelope{Version: "1.0"}
}

func quickHelp(sessionAttributes map[string]interface{}) *alexa.ResponseEnvelope {
	return &alexa.ResponseEnvelope{Version: "1.0",
		Response:          &alexa.Response{OutputSpeech: plainText(quickHelpText)},
		SessionAttributes: sessionAttributes,
	}
}

func plainText(text string) *alexa.OutputSpeech {
	return &alexa.OutputSpeech{Type: "PlainText", Text: text}
}

func ssmlText(text string) *alexa.OutputSpeech {
	return &alexa.OutputSpeech{Type: "SSML", SSML: text}
}

func internalError(l *i18n.Localizer) *alexa.ResponseEnvelope {
	return &alexa.ResponseEnvelope{Version: "1.0",
		Response: &alexa.Response{
			OutputSpeech: plainText(l.MustLocalize(&LocalizeConfig{DefaultMessage: &Message{
				ID:    "InternalError",
				Other: "Es ist ein interner Fehler aufgetreten bei der Benutzung von Wikipedia.",
			}})),
			ShouldSessionEnd: false,
		},
	}
}
