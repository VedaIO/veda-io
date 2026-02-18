package app

// BlockedAppDetail represents the details of a blocked application.
type BlockedAppDetail struct {
	Name    string `json:"name"`
	ExePath string `json:"exe_path"`
}
