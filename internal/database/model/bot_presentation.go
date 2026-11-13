package model

// BotPresentation is the owner's buyer-facing shop identity and menu preferences.
type BotPresentation struct {
	BrandName       string            `json:"brandName"`
	MiniAppURL      string            `json:"miniAppURL"`
	ButtonStyles    map[string]string `json:"buttonStyles"`
	WelcomeEN       string            `json:"welcomeEN"`
	Announcement    string            `json:"announcement"`
	SupportURL      string            `json:"supportURL"`
	TutorialURL     string            `json:"tutorialURL"`
	DefaultLanguage string            `json:"defaultLanguage"`
	MenuLayout      string            `json:"menuLayout"`
	HideTrial       bool              `json:"hideTrial"`
	HideSupport     bool              `json:"hideSupport"`
	HideTutorial    bool              `json:"hideTutorial"`
	Labels          map[string]string `json:"labels"`
}
