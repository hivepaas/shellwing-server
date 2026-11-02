package config

type DevMode struct {
	Enabled         bool `toml:"-" env:"-"`
	ForceAgentLocal bool `toml:"force_agent_local" env:"SHW_DEV_MODE_FORCE_AGENT_LOCAL"`
	// UserPassword, when set, becomes every user's password each time the app
	// starts: a dev server reseeded on each deploy keeps one password its team
	// knows, rather than the seed's, which is public.
	UserPassword string `toml:"user_password" env:"SHW_DEV_MODE_USER_PASSWORD"`
	// UserLimit, when set, is the limit of users with no license in force, in place of the free one: a seed has more
	// users than that. It works only in development, as everything here does.
	UserLimit int `toml:"user_limit" env:"SHW_DEV_MODE_USER_LIMIT"`
}
