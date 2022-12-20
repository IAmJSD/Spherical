package apiv1

import (
	"net/http"

	"github.com/jakemakesstuff/spherical/config"
	"github.com/jakemakesstuff/spherical/utils/httpresponse"
)

func nodeInfoHn(w http.ResponseWriter, r *http.Request) {
	c := config.Config()
	if c.Locale == "" {
		c.Locale = "system"
	}
	if c.ServerBackground == "" {
		c.ServerBackground = "/png/default_bg.png"
	}
	httpresponse.MarshalToAccept(w, r, http.StatusOK, map[string]any{
		"server_name":        c.ServerName,
		"server_description": c.ServerDescription,
		"server_background":  c.ServerBackground,
		"locale":             c.Locale,
		"public":             c.ServerPublic,
		"sign_ups_enabled":   c.SignUpsEnabled,
	})
}
