package config

import "sync"

type Structure struct {
	// Setup is set to true when the out of box experience is complete.
	Setup bool `config:"setup"`

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

	// S3AccessKeyID is used to define an access key ID.
	S3AccessKeyID string `config:"s3_access_key_id"`

	// S3SecretAccessKey is used to define a secret access key.
	S3SecretAccessKey string `config:"s3_secret_access_key"`

	// S3Bucket is used to define the S3 bucket.
	S3Bucket string `config:"s3_bucket"`

	// S3Endpoint is used to define the S3 endpoint.
	S3Endpoint string `config:"s3_endpoint"`

	// S3Hostname is used to define the hostname of S3 uploads.
	S3Hostname string `config:"s3_hostname"`

	// S3Region is used to define the S3 region.
	S3Region string `config:"s3_region"`

	// ServerName is used to define the name of Spherical server.
	ServerName string `config:"server_name"`

	// ServerDescription is used to define the description of the Spherical server.
	ServerDescription string `config:"server_description"`

	// ServerBackground is used to define the background of the Spherical server.
	ServerBackground string `config:"server_background"`

	// ServerPublic defines if guilds or members within a server are public.
	ServerPublic bool `config:"server_public"`

	// SignUpsEnabled is used to define if this server is accepting sign ups.
	SignUpsEnabled bool `config:"sign_ups_enabled"`

	// Locale is used to define the servers default locale. If this is blank, we use the users locale.
	Locale string `config:"locale"`

	// JobCount is used to define the number of jobs. Defaults to 120.
	JobCount uint `config:"job_count"`

	// SchedulerSleepTime is how number of milliseconds the scheduler sleeps for. Defaults to 1000ms.
	SchedulerSleepTime uint `config:"scheduler_sleep_time"`

	// Consensus is the level of consensus that is required for hash validation. Defaults to 3.
	Consensus uint `config:"consensus"`

	// Informers are hash verifiers that should be informed of new hashes.
	Informers []string `config:"informers"`

	// Trusted is a list of trusted hash verifiers. These are used to decide if a message is genuine.
	Trusted []string `config:"trusted"`
}

var (
	configVar     Structure
	configVarLock sync.RWMutex
)

// Config is used to return the current configuration.
func Config() Structure {
	configVarLock.RLock()
	c := configVar
	configVarLock.RUnlock()
	return c
}
