package resellerbot

import (
	"html"
	"strings"

	"github.com/mhsanaei/3x-ui/v3/internal/database/model"

	"github.com/mymmrac/telego"
	tu "github.com/mymmrac/telego/telegoutil"
)

var menuKeys = map[string]string{"buy": "menuBuyNew", "trial": "menuTrial", "invoices": "menuInvoices", "subs": "menuSubs", "account": "menuAccount", "support": "menuSupport", "tutorial": "menuTutorial"}

func menuHidden(action string, cfg *model.BotConfig) bool {
	if cfg == nil {
		return false
	}
	p := cfg.Presentation
	return action == "trial" && p.HideTrial || action == "support" && p.HideSupport || action == "tutorial" && p.HideTutorial
}

func shopMenuLabel(lang, action string, cfg *model.BotConfig) string {
	if cfg != nil {
		if label := strings.TrimSpace(cfg.Presentation.Labels[lang+":"+action]); label != "" {
			return label
		}
	}
	return t(lang, menuKeys[action])
}

func shopButtonStyle(action string, cfg *model.BotConfig) string {
	if cfg != nil {
		switch style := cfg.Presentation.ButtonStyles[action]; style {
		case "default":
			return ""
		case "primary", "success", "danger":
			return style
		}
	}
	switch action {
	case "buy", "subs":
		return "success"
	case "trial", "invoices", "account", "tutorial":
		return "primary"
	case "support":
		return "danger"
	}
	return ""
}

func shopMenuKeyboard(lang string, cfg *model.BotConfig) *telego.ReplyKeyboardMarkup {
	rows := [][]string{{"buy"}, {"trial"}, {"invoices", "subs"}, {"account"}, {"support", "tutorial"}}
	if cfg != nil && cfg.Presentation.MenuLayout == "compact" {
		rows = [][]string{{"buy", "subs"}, {"trial", "invoices"}, {"account"}, {"support", "tutorial"}}
	}
	kb := &telego.ReplyKeyboardMarkup{ResizeKeyboard: true, IsPersistent: true}
	for _, row := range rows {
		buttons := []telego.KeyboardButton{}
		for _, action := range row {
			if !menuHidden(action, cfg) {
				button := tu.KeyboardButton(shopMenuLabel(lang, action, cfg))
				button.Style = shopButtonStyle(action, cfg)
				buttons = append(buttons, button)
			}
		}
		if len(buttons) > 0 {
			kb.Keyboard = append(kb.Keyboard, buttons)
		}
	}
	return kb
}

func shopMenuAction(lang, text string, cfg *model.BotConfig) string {
	for _, action := range []string{"buy", "trial", "invoices", "subs", "account", "support", "tutorial"} {
		if !menuHidden(action, cfg) && text == shopMenuLabel(lang, action, cfg) {
			return action
		}
	}
	action := menuAction(lang, text)
	if menuHidden(action, cfg) {
		return ""
	}
	return action
}

func shopWelcome(lang string, cfg *model.BotConfig) string {
	if cfg == nil {
		return t(lang, "welcome")
	}
	p := cfg.Presentation
	parts := []string{}
	if p.BrandName != "" {
		parts = append(parts, "<b>"+html.EscapeString(p.BrandName)+"</b>")
	}
	welcome := cfg.WelcomeMessage
	if lang == langEN {
		welcome = p.WelcomeEN
	}
	if welcome == "" {
		welcome = t(lang, "welcome")
	}
	parts = append(parts, html.EscapeString(welcome))
	if p.Announcement != "" {
		parts = append(parts, html.EscapeString(p.Announcement))
	}
	return strings.Join(parts, "\n\n")
}
