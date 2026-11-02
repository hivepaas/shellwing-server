package cacheentity

// DeviceAuthCode is what a person approved in their browser, until the app comes for its tokens.
type DeviceAuthCode struct {
	UserID        string `json:"userId"`
	CodeChallenge string `json:"codeChallenge"`
	RedirectURI   string `json:"redirectUri"`
	DeviceName    string `json:"deviceName"`
	Platform      string `json:"platform"`
}
