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
	httpresponse.MarshalToAccept(w, r, http.StatusOK, map[string]any{
		"server_name":        c.ServerName,
		"server_description": c.ServerDescription,
		"locale":             c.Locale,
		"public":             c.ServerPublic,
		"sign_ups_enabled":   c.SignUpsEnabled,
	})
}
