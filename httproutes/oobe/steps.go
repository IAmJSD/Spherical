package oobe

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jakemakesstuff/spherical/config"
	"github.com/jakemakesstuff/spherical/db"
	"github.com/jakemakesstuff/spherical/utils/s3"
)

func boolPtr(x bool) *bool {
	return &x
}

func welcome(dev bool) installStage {
	return installStage{
		Step:        "welcome",
		ImageURL:    "/png/wave.png",
		ImageAlt:    "httproutes/oobe/steps:wave",
		Title:       "httproutes/oobe/steps:welcome_title",
		Description: "httproutes/oobe/steps:welcome_description",
		Options: []setupOption{
			{
				ID:          "setup_key",
				Type:        setupTypeInput,
				Name:        "httproutes/oobe/steps:setup_key_name",
				Description: "httproutes/oobe/steps:setup_key_description",
				Sticky:      true,
				Required:    true,
				MustSecure:  boolPtr(dev),
			},
			{
				ID:       "hostname",
				Type:     setupTypeHostname,
				Name:     "httproutes/oobe/steps:hostname",
				Required: true,
			},
		},
		NextButton: "httproutes/oobe/steps:setup_next",
		Pass: func() bool {
			// Always true because the stage passing checks run after the GET and setup key.
			return true
		},
		Run: func(ctx context.Context, m map[string]json.RawMessage) string {
			// Special case: We have already validated the setup key in the main
			// handler. Just get the hostname information.
			return ""
		},
	}
}

func s3Conf(_ bool) installStage {
	return installStage{
		Step:        "s3",
		ImageURL:    "/png/cloud.png",
		ImageAlt:    "httproutes/oobe/steps:cloud_alt",
		Title:       "httproutes/oobe/steps:s3_title",
		Description: "httproutes/oobe/steps:s3_description",
		Options: []setupOption{
			{
				ID:          "s3_access_key_id",
				Type:        setupTypeInput,
				Name:        "httproutes/oobe/steps:access_key_id_name",
				Description: "httproutes/oobe/steps:access_key_id_description",
				Sticky:      false,
				Required:    true,
			},
			{
				ID:          "s3_secret_access_key",
				Type:        setupTypeSecret,
				Name:        "httproutes/oobe/steps:secret_access_key_name",
				Description: "httproutes/oobe/steps:secret_access_key_description",
				Sticky:      false,
				Required:    true,
			},
			{
				ID:          "s3_bucket",
				Type:        setupTypeInput,
				Name:        "httproutes/oobe/steps:s3_bucket_name",
				Description: "httproutes/oobe/steps:s3_bucket_description",
				Sticky:      false,
				Required:    true,
			},
			{
				ID:          "s3_endpoint",
				Type:        setupTypeInput,
				Name:        "httproutes/oobe/steps:s3_endpoint_name",
				Description: "httproutes/oobe/steps:s3_endpoint_description",
				Sticky:      false,
				Required:    true,
			},
			{
				ID:          "s3_region",
				Type:        setupTypeInput,
				Name:        "httproutes/oobe/steps:s3_region_name",
				Description: "httproutes/oobe/steps:s3_region_description",
				Sticky:      false,
				Required:    true,
			},
			{
				ID:          "s3_hostname",
				Type:        setupTypeInput,
				Name:        "httproutes/oobe/steps:s3_hostname_name",
				Description: "httproutes/oobe/steps:s3_hostname_description",
				Sticky:      false,
				Required:    false,
			},
		},
		NextButton: "httproutes/oobe/steps:test_s3_configuration",
		Pass: func() bool {
			return config.Config().S3AccessKeyID != ""
		},
		Run: func(ctx context.Context, m map[string]json.RawMessage) string {
			// Handle writing to the database.
			allowed := []string{
				"s3_access_key_id", "s3_secret_access_key", "s3_bucket", "s3_endpoint",
				"s3_hostname", "s3_region",
			}
			requiredCount := 0
			for k, v := range m {
				// Get the value as a string.
				var s string
				err := json.Unmarshal(v, &s)
				if err != nil {
					continue
				}

				// Check if the key is allowed.
				included := false
				for _, v := range allowed {
					if v == k {
						included = true
						break
					}
				}

				if included {
					// Add 1 to the required count.
					requiredCount++

					// Write to the database.
					err := db.UpdateConfig(ctx, k, s)
					if err != nil {
						return "httproutes/oobe/steps:update_config_fail"
					}
				}
			}

			// Defer reverting the key used for skipping in the event of error.
			revert := false
			defer func() {
				if revert {
					// Reset s3_access_key_id so that skip ignores it.
					_ = db.UpdateConfig(ctx, "s3_access_key_id", "")
				}
			}()

			// Defines handling deletes.
			s := uuid.New().String()
			path := s + "_test.txt"
			deleteSuccess := false
			doDelete := func() {
				if deleteSuccess {
					// Delete was already ran successfully. This is defer.
					return
				}
				err := s3.Delete(ctx, path)
				deleteSuccess = err == nil
			}
			defer doDelete()

			// Test uploading a file and reading it.
			url, err := s3.Upload(ctx, path, strings.NewReader("Hello World!"),
				-1, "text/plain", "", "public-read")
			if err != nil {
				revert = true
				return "httproutes/oobe/steps:s3_upload_fail"
			}
			ctx, cancel := context.WithTimeout(ctx, time.Second*5)
			defer cancel()
			req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
			if err != nil {
				revert = true
				return "httproutes/oobe/steps:resulting_url_parse_fail"
			}
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				revert = true
				return "httproutes/oobe/steps:s3_http_get_failed"
			}
			b, err := io.ReadAll(resp.Body)
			if err != nil || !bytes.Equal(b, []byte("Hello World!")) {
				revert = true
				return "httproutes/oobe/steps:s3_http_get_failed"
			}

			// Run the delete handler. Error if we cannot for some reason.
			doDelete()
			if !deleteSuccess {
				revert = true
				return "httproutes/oobe/steps:s3_delete_failed"
			}

			// No errors!
			return ""
		},
	}
}

func done(_ bool) installStage {
	return installStage{
		Step:        "done",
		ImageURL:    "/png/tick.png",
		ImageAlt:    "httproutes/oobe/state:done_tick",
		Title:       "httproutes/oobe/state:done_title",
		Description: "httproutes/oobe/state:done_description",
		Options:     []setupOption{},
		NextButton:  "httproutes/oobe/state:complete_installation",
		Pass: func() bool {
			return config.Config().Setup
		},
		Run: func(ctx context.Context, _ map[string]json.RawMessage) string {
			err := db.UpdateConfig(ctx, "setup", true)
			if err != nil {
				return "httproutes/oobe/state:update_config_fail"
			}
			return ""
		},
	}
}

func init() {
	addStages(
		welcome, s3Conf,
		//email, serverInfo, ownerUser, hashValidators, // TODO
		done)
}
