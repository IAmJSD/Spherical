package config

import "sync"

type Structure struct {
	// Setup is set to true when the out of box experience is complete.
	Setup bool `config:"setup"`

	// Locale is used to define the servers default locale. If this is blank, we use the users locale.
	Locale string `config:"locale"`

	// Hostname defines the host that this instance is running from.
	Hostname string `config:"hostname"`

	// HTTPS defines if this is an HTTPS host. This should only be false on dev systems.
	HTTPS bool `config:"https"`

	// MailFrom defines the from header for e-mails.
	MailFrom string `config:"mail_from"`

	// SMTPHost is used to define the SMTP hostname.
	SMTPHost string `config:"smtp_host"`

	// SMTPPort is used to define the SMTP port.
	SMTPPort int `config:"smtp_port"`

	// SMTPUsername is used to define the SMTP username. If blank, will use MailFrom.
	SMTPUsername string `config:"smtp_username"`

	// SMTPPassword is used to define the SMTP password.
	SMTPPassword string `config:"smtp_password"`

	// SMTPSecure is used to define if SMTP should be secure.
	SMTPSecure bool `config:"smtp_secure"`

	// JobCount is used to define the number of jobs. Defaults to 120.
	JobCount uint `config:"job_count"`

	// SchedulerSleepTime is how number of milliseconds the scheduler sleeps for. Defaults to 1000ms.
	SchedulerSleepTime uint `config:"scheduler_sleep_time"`
}

var (
	configVar     Structure
	configVarLock = sync.RWMutex{}
)

// Config is used to return the current configuration.
func Config() Structure {
	configVarLock.RLock()
	c := configVar
	configVarLock.RUnlock()
	return c
}
