package config

type Users struct {
	Admin UserAdmin `toml:"admin"`
	Demo  UserDemo  `toml:"demo"`

	// SelfSignupEnabled lets anybody create an account with an email address, which a link sent to it
	// verifies: what a customer does before starting a trial or buying a license. Off, an account is only
	// made by an admin's invitation, as on an installation that sells nothing.
	SelfSignupEnabled bool `toml:"self_signup_enabled" env:"SHW_USERS_SELF_SIGNUP_ENABLED"`
}

type UserAdmin struct {
	Email    string `toml:"email" env:"SHW_USER_ADMIN_EMAIL"`
	Username string `toml:"username" env:"SHW_USER_ADMIN_USERNAME"`
	Password string `toml:"password" env:"SHW_USER_ADMIN_PASSWORD"`
}

type UserDemo struct {
	UserID string `toml:"user_id" env:"SHW_USER_DEMO_USER_ID"`
}
