package oobe

import (
	"context"
	"crypto/subtle"
	"encoding/json"
	"io"
	"net/http"
	"strconv"

	"github.com/jakemakesstuff/spherical/db"
	"github.com/jakemakesstuff/spherical/i18n"
)

type setupType string

const (
	// setupTypeHostname is used to define a setup option where the hostname and
	// protocol are sent in the structure {hostname: string, protocol: string}. This
	// should fail if not https and MustSecure is true.
	setupTypeHostname setupType = "hostname"

	// setupTypeInput is used to define a small input box.
	setupTypeInput setupType = "input"

	// setupTypeSecret is used to define a password box.
	setupTypeSecret setupType = "secret"

	// setupTypeNumber is used to define a number box.
	setupTypeNumber setupType = "number"

	// setupTypeTextbox is used to define a large textbox.
	setupTypeTextbox setupType = "textbox"

	// setupTypeBoolean is used to define a boolean within the setup.
	setupTypeBoolean setupType = "boolean"
)

type setupOption struct {
	// ID is used to define the ID of the option.
	ID string `json:"id"`

	// Type defines the type of the setup option.
	Type setupType `json:"type"`

	// Name is used to define the name of the option.
	Name string `json:"name"`

	// Description is used to define the description of the option. It being blank means none.
	Description string `json:"description"`

	// Sticky is an option that should persist across the remainder.
	Sticky bool `json:"sticky"`

	// MustSecure is used with the hostname structure to determine if the option has to
	// be within a secure context to succeed.
	MustSecure *bool `json:"must_secure,omitempty"`

	// Regexp is used for text inputs to validate the contents before they hit the server.
	Regexp *string `json:"regexp,omitempty"`

	// Required is used to define if the option is required.
	Required bool `json:"required"`
}

type installData struct {
	// ImageURL is the image that should be displayed on the page.
	ImageURL string `json:"image_url"`

	// ImageAlt is the image alt text that should be displayed on the page.
	ImageAlt string `json:"image_alt"`

	// Title is the title within the currently set language.
	Title string `json:"title"`

	// Description is the description within the currently set language.
	Description string `json:"description"`

	// Step is the step that the client should send back with the content.
	Step string `json:"step"`

	// Options is used to define the options contained within the current page.
	Options []setupOption `json:"options"`

	// NextButton is the text for the next button within the currently set language.
	NextButton string `json:"next_button"`
}

type installStage struct {
	// Step is the ID of the step.
	Step string

	// ImageURL is the image that should be displayed on the page.
	ImageURL string

	// ImageAlt is used to define the i18n string for the image alt.
	ImageAlt string

	// Title is used to define the i18n string for the title.
	Title string

	// Description is used to define the i18n string for the description.
	Description string

	// Options is used to define the options contained within the current page.
	// All text elements should be the i18n string.
	Options []setupOption

	// NextButton is used to define the i18n string for next.
	NextButton string

	// Pass defines if this stage should be passed.
	Pass func() bool

	// Run defines the runner for the options. If string is blank, the next thing is ran.
	Run func(context.Context, map[string]json.RawMessage) string
}

func (i installStage) render(r *http.Request) installData {
	opts := make([]setupOption, len(i.Options))
	for i, v := range i.Options {
		v.Name = i18n.GetWithRequest(r, v.Name)
		v.Description = i18n.GetWithRequest(r, v.Description)
		opts[i] = v
	}
	return installData{
		ImageURL:    i.ImageURL,
		ImageAlt:    i18n.GetWithRequest(r, i.ImageAlt),
		Title:       i18n.GetWithRequest(r, i.Title),
		Description: i18n.GetWithRequest(r, i.Description),
		Step:        i.Step,
		Options:     opts,
		NextButton:  i18n.GetWithRequest(r, i.NextButton),
	}
}

var (
	stages        = []func(bool) installStage{}
	stagesIndexes = map[string]int{}
)

func addStages(localStages ...func(bool) installStage) {
	for _, v := range localStages {
		stages = append(stages, v)
		stagesIndexes[v(false).Step] = len(stages) - 1
	}
}

func sendJson(w http.ResponseWriter, x any, statusCode int) {
	b, _ := json.Marshal(x)
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Length", strconv.Itoa(len(b)))
	w.WriteHeader(statusCode)
	_, _ = w.Write(b)
}

type messageData struct {
	Message string `json:"message"`
}

type stateBody struct {
	Type string                     `json:"type"`
	Body map[string]json.RawMessage `json:"body"`
}

const hundredkb = 100000

func installState(w http.ResponseWriter, r *http.Request) {
	// Check if this is a GET request. If it is, render stage 0.
	isDev := r.Context().Value("dev").(bool)
	if r.Method == "GET" {
		sendJson(w, stages[0](isDev).render(r), 200)
		return
	}

	// Parse the body.
	b, err := io.ReadAll(io.LimitReader(r.Body, hundredkb))
	if err != nil {
		sendJson(w, messageData{Message: i18n.GetWithRequest(
			r, "httproutes/oobe/state:body_read_fail")}, 400)
		return
	}

	// Parse the body as JSON.
	var body stateBody
	err = json.Unmarshal(b, &body)
	if err != nil {
		sendJson(w, messageData{Message: err.Error()}, 400)
		return
	}

	// Get the setup key.
	var setupKey string
	b = body.Body["setup_key"]
	if b != nil {
		_ = json.Unmarshal(b, &setupKey)
	}
	if setupKey == "" {
		sendJson(w, messageData{Message: i18n.GetWithRequest(
			r, "httproutes/oobe/state:setup_key_nonexistent")}, 400)
		return
	}
	dbSetupKey, err := db.SetupKey(r.Context())
	if err != nil {
		sendJson(w, messageData{Message: i18n.GetWithRequest(
			r, "httproutes/oobe/state:setup_key_db")}, 400)
		return
	}
	if subtle.ConstantTimeCompare([]byte(setupKey), []byte(dbSetupKey)) != 1 {
		sendJson(w, messageData{Message: i18n.GetWithRequest(
			r, "httproutes/oobe/state:setup_keys_unequal")}, 400)
		return
	}

	// Get the index of the type.
	index, ok := stagesIndexes[body.Type]
	if !ok {
		// Specified type does not exist.
		sendJson(
			w, messageData{Message: i18n.GetWithRequest(
				r, "httproutes/oobe/state:event_nonexistent")},
			400)
		return
	}
	stg := stages[index](isDev)

	// Run the function.
	msg := stg.Run(r.Context(), body.Body)
	if msg != "" {
		// Error.
		sendJson(w, messageData{Message: i18n.GetWithRequest(r, msg)}, 400)
		return
	}

	// Find the next option.
	for _, optFactory := range stages[index+1:] {
		// See if this option is done.
		opt := optFactory(isDev)
		done := opt.Pass()
		if !done {
			// Serve this option.
			sendJson(w, opt.render(r), 200)
			return
		}
	}

	// We are done!
	err = db.UpdateConfig(r.Context(), "setup", true)
	if msg != "" {
		// DB write error.
		sendJson(w, messageData{Message: i18n.GetWithRequest(
			r, "httproutes/oobe/state:setup_write_fail")}, 400)
		return
	}
	http.Redirect(w, r, "/", http.StatusPermanentRedirect)
}
