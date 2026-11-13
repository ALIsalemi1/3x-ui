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

// isRetiredGraphicalShopButton recognizes obsolete reply-keyboard labels that
// may still be cached by Telegram clients after the Mini App was removed.
// These are *input only*; never add them to a new keyboard.
func isRetiredGraphicalShopButton(text string) bool {
	label := strings.TrimSpace(text)
	for _, emoji := range []string{"🛍️", "🛍", "🛒", "🎨", "🌈"} {
		label = strings.TrimSpace(strings.TrimPrefix(label, emoji))
		label = strings.TrimSpace(strings.TrimSuffix(label, emoji))
	}
	switch label {
	case "ورود به فروشگاه گرافیکی", "Open graphical shop", "Open graphical store":
		return true
	default:
		return false
	}
}

func shopMenuLabel(lang, action string, cfg *model.BotConfig) string {
	if cfg != nil {
		if label := strings.TrimSpace(cfg.Presentation.Labels[lang+":"+action]); label != "" && !isRetiredGraphicalShopButton(label) {
			return label
		}
	}
	return t(lang, menuKeys[action])
}

// Telegram only supports the three native background styles. Extended
// accent choices use a visible emoji in the button label instead.
func shopAccentIndicator(style string) string {
	switch style {
	case "pink":
		return "🩷"
	case "yellow":
		return "🟡"
	case "purple":
		return "🟣"
	case "orange":
		return "🟠"
	case "cyan":
		return "🩵"
	case "brown":
		return "🟤"
	case "black":
		return "⚫"
	default:
		return ""
	}
}

// Default accents make the standard menu colorful while preserving owners'
// explicit choices and supported native Telegram button backgrounds.
func shopDefaultAccent(action string) string {
	switch action {
	case "trial":
		return "yellow"
	case "invoices":
		return "pink"
	case "subs":
		return "purple"
	case "account":
		return "cyan"
	case "tutorial":
		return "orange"
	default:
		return ""
	}
}

func shopMenuDisplayLabel(lang, action string, cfg *model.BotConfig) string {
	label := shopMenuLabel(lang, action, cfg)
	style := ""
	if cfg != nil {
		style = cfg.Presentation.ButtonStyles[action]
	}
	if style == "" || style == "auto" {
		style = shopDefaultAccent(action)
	}
	indicator := shopAccentIndicator(style)
	if indicator == "" || strings.HasPrefix(label, indicator+" ") {
		return label
	}
	return indicator + " " + label
}

func stripShopAccent(text string) string {
	for _, style := range []string{"pink", "yellow", "purple", "orange", "cyan", "brown", "black"} {
		indicator := shopAccentIndicator(style) + " "
		if strings.HasPrefix(text, indicator) {
			return strings.TrimPrefix(text, indicator)
		}
	}
	return text
}

func shopButtonStyle(action string, cfg *model.BotConfig) string {
	if cfg != nil {
		switch style := cfg.Presentation.ButtonStyles[action]; style {
		case "default", "pink", "yellow", "purple", "orange", "cyan", "brown", "black":
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
	rows := [][]string{{"buy"}, {"subs", "invoices"}, {"trial", "account"}, {"support", "tutorial"}}
	if cfg != nil && cfg.Presentation.MenuLayout == "compact" {
		rows = [][]string{{"buy", "subs"}, {"trial", "invoices"}, {"account"}, {"support", "tutorial"}}
	}
	kb := &telego.ReplyKeyboardMarkup{ResizeKeyboard: true, IsPersistent: true}
	for _, row := range rows {
		buttons := []telego.KeyboardButton{}
		for _, action := range row {
			if !menuHidden(action, cfg) {
				button := tu.KeyboardButton(shopMenuDisplayLabel(lang, action, cfg))
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
	normalized := stripShopAccent(text)
	for _, action := range []string{"buy", "trial", "invoices", "subs", "account", "support", "tutorial"} {
		if !menuHidden(action, cfg) && (text == shopMenuDisplayLabel(lang, action, cfg) || normalized == shopMenuLabel(lang, action, cfg)) {
			return action
		}
	}
	action := menuAction(lang, normalized)
	if menuHidden(action, cfg) {
		return ""
	}
	return action
}

// shopWelcome builds an HTML-safe, attractive landing message.
func shopWelcome(lang string, cfg *model.BotConfig) string {
	if cfg == nil {
		cfg = &model.BotConfig{}
	}
	p := cfg.Presentation
	brand := strings.TrimSpace(p.BrandName)
	if brand == "" {
		brand = "VPN SHOP"
	}
	welcome := cfg.WelcomeMessage
	if lang == langEN {
		welcome = p.WelcomeEN
	}
	welcome = strings.TrimSpace(welcome)
	if welcome == "" {
		welcome = t(lang, "welcome")
	}
	hint := "👇 از دکمه\u200cهای پایین، بخش موردنظرت رو انتخاب کن."
	notice := "اطلاعیه"
	if lang == langEN {
		hint = "👇 Choose a section using the buttons below."
		notice = "Announcement"
	}
	parts := []string{
		"<b>✨ " + html.EscapeString(brand) + "</b>",
		"━━━━━━━━━━━━━━",
		html.EscapeString(welcome),
	}
	if announcement := strings.TrimSpace(p.Announcement); announcement != "" {
		parts = append(parts, "<b>📣 "+notice+"</b>\n"+html.EscapeString(announcement))
	}
	parts = append(parts, "━━━━━━━━━━━━━━\n"+hint)
	return strings.Join(parts, "\n\n")
}
