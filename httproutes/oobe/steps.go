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
	"github.com/jakemakesstuff/spherical/jobs"
	"github.com/jakemakesstuff/spherical/utils/s3"
)

func boolPtr(x bool) *bool {
	return &x
}

type hostname struct {
	Protocol string `json:"protocol"`
	Hostname string `json:"hostname"`
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
			var h hostname
			b, ok := m["hostname"]
			if ok {
				err := json.Unmarshal(b, &h)
				if err != nil {
					ok = false
				}
			}
			if !ok {
				return "httproutes/oobe/steps:no_hostname"
			}
			err := db.UpdateConfig(ctx, "https", h.Protocol != "http")
			if err != nil {
				return "httproutes/oobe/steps:update_config_fail"
			}
			err = db.UpdateConfig(ctx, "hostname", h.Hostname)
			if err != nil {
				return "httproutes/oobe/steps:update_config_fail"
			}
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

func email(_ bool) installStage {
	return installStage{
		Step:        "email",
		ImageURL:    "/png/mailbox.png",
		ImageAlt:    "httproutes/oobe/steps:mailbox",
		Title:       "httproutes/oobe/steps:email_title",
		Description: "httproutes/oobe/steps:email_description",
		Options: []setupOption{
			{
				ID:          "mail_from",
				Type:        setupTypeInput,
				Name:        "httproutes/oobe/steps:mail_from_name",
				Description: "httproutes/oobe/steps:mail_from_description",
				Sticky:      false,
				Required:    false,
			},
			{
				ID:          "smtp_host",
				Type:        setupTypeInput,
				Name:        "httproutes/oobe/steps:smtp_host_name",
				Description: "httproutes/oobe/steps:smtp_host_description",
				Sticky:      false,
				Required:    true,
			},
			{
				ID:          "smtp_port",
				Type:        setupTypeNumber,
				Name:        "httproutes/oobe/steps:smtp_port_name",
				Description: "httproutes/oobe/steps:smtp_port_description",
				Sticky:      false,
				Required:    false,
			},
			{
				ID:          "smtp_username",
				Type:        setupTypeInput,
				Name:        "httproutes/oobe/steps:smtp_username_name",
				Description: "httproutes/oobe/steps:smtp_username_description",
				Sticky:      false,
				Required:    true,
			},
			{
				ID:          "smtp_password",
				Type:        setupTypeSecret,
				Name:        "httproutes/oobe/steps:smtp_password_name",
				Description: "httproutes/oobe/steps:smtp_password_description",
				Sticky:      false,
				Required:    true,
			},
			{
				ID:          "mail_to",
				Type:        setupTypeInput,
				Name:        "httproutes/oobe/steps:mail_to_name",
				Description: "httproutes/oobe/steps:mail_to_description",
				Sticky:      false,
				Required:    true,
			},
			{
				ID:          "smtp_secure",
				Type:        setupTypeBoolean,
				Name:        "httproutes/oobe/steps:smtp_secure_name",
				Description: "httproutes/oobe/steps:smtp_secure_description",
				Sticky:      false,
				Required:    true,
			},
		},
		NextButton: "httproutes/oobe/steps:test_email_configuration",
		Pass: func() bool {
			return config.Config().SMTPHost != ""
		},
		Run: func(ctx context.Context, m map[string]json.RawMessage) string {
			// Handle defaulting the SMTP port.
			x, ok := m["smtp_port"]
			if !ok || bytes.Equal(x, []byte("0")) || bytes.Equal(x, []byte("null")) {
				m["smtp_port"] = []byte("25")
			}

			// Handle writing to the database.
			allowed := map[string]func() any{
				"mail_from":     func() any { return "" },
				"smtp_host":     func() any { return "" },
				"smtp_port":     func() any { return 0 },
				"smtp_username": func() any { return "" },
				"smtp_password": func() any { return "" },
				"smtp_secure":   func() any { return false },
			}
			for k, v := range m {
				// Check if the key is allowed.
				var includedFactory func() any
				for v, factory := range allowed {
					if v == k {
						includedFactory = factory
						break
					}
				}

				if includedFactory != nil {
					// Get the value as the factory specified.
					r := includedFactory()
					err := json.Unmarshal(v, &r)
					if err != nil {
						continue
					}

					// Write to the database.
					err = db.UpdateConfig(ctx, k, r)
					if err != nil {
						return "httproutes/oobe/steps:update_config_fail"
					}
				}
			}

			// Handle sending mail.
			var mailTo string
			_ = json.Unmarshal(m["mail_to"], &mailTo)
			if mailTo != "" {
				err := jobs.MailerJob.RunAndBlock(ctx, jobs.MailEvent{
					To:          mailTo,
					Subject:     "E-mail Sending Test",
					ContentHTML: "<p>This is a test of sending mail with Spherical.</p>",
				})
				if err != nil {
					_ = db.UpdateConfig(ctx, "smtp_host", "")
					return "httproutes/oobe/steps:email_send_fail"
				}
			}

			// No errors!
			return ""
		},
	}
}

func serverInfo(_ bool) installStage {
	return installStage{
		Step:        "server_info",
		ImageURL:    "/png/cog.png",
		ImageAlt:    "httproutes/oobe/steps:cog",
		Title:       "httproutes/oobe/steps:server_info_title",
		Description: "httproutes/oobe/steps:server_info_description",
		Options: []setupOption{
			{
				ID:          "server_name",
				Type:        setupTypeInput,
				Name:        "httproutes/oobe/steps:server_name_name",
				Description: "httproutes/oobe/steps:server_name_description",
				Sticky:      false,
				Required:    true,
			},
			{
				ID:          "server_description",
				Type:        setupTypeTextbox,
				Name:        "httproutes/oobe/steps:server_description_name",
				Description: "httproutes/oobe/steps:server_description_description",
				Sticky:      false,
				Required:    true,
			},
			{
				ID:          "server_public",
				Type:        setupTypeBoolean,
				Name:        "httproutes/oobe/steps:server_public_name",
				Description: "httproutes/oobe/steps:server_public_description",
				Sticky:      false,
				Required:    true,
			},
			{
				ID:          "sign_ups_enabled",
				Type:        setupTypeBoolean,
				Name:        "httproutes/oobe/steps:sign_ups_enabled_name",
				Description: "httproutes/oobe/steps:sign_ups_enabled_description",
				Sticky:      false,
				Required:    true,
			},
		},
		NextButton: "httproutes/oobe/steps:save_server_info",
		Pass: func() bool {
			return config.Config().ServerName != ""
		},
		Run: func(ctx context.Context, m map[string]json.RawMessage) string {
			allowed := map[string]func() any{
				"server_name":        func() any { return "" },
				"server_description": func() any { return "" },
				"server_public":      func() any { return false },
				"sign_ups_enabled":   func() any { return false },
			}
			for k, v := range m {
				// Check if the key is allowed.
				var includedFactory func() any
				for v, factory := range allowed {
					if v == k {
						includedFactory = factory
						break
					}
				}

				if includedFactory != nil {
					// Get the value as the factory specified.
					r := includedFactory()
					err := json.Unmarshal(v, &r)
					if err != nil {
						continue
					}

					// Write to the database.
					err = db.UpdateConfig(ctx, k, r)
					if err != nil {
						return "httproutes/oobe/steps:update_config_fail"
					}
				}
			}
			return ""
		},
	}
}

func done(_ bool) installStage {
	return installStage{
		Step:        "done",
		ImageURL:    "/png/tick.png",
		ImageAlt:    "httproutes/oobe/steps:done_tick",
		Title:       "httproutes/oobe/steps:done_title",
		Description: "httproutes/oobe/steps:done_description",
		Options:     []setupOption{},
		NextButton:  "httproutes/oobe/steps:complete_installation",
		Pass: func() bool {
			return config.Config().Setup
		},
		Run: func(ctx context.Context, _ map[string]json.RawMessage) string {
			err := db.UpdateConfig(ctx, "setup", true)
			if err != nil {
				return "httproutes/oobe/steps:update_config_fail"
			}
			return ""
		},
	}
}

func init() {
	addStages(
		welcome, s3Conf, email, serverInfo, //ownerUser, hashValidators, // TODO
		done)
}
