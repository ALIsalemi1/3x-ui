package resellerbot

import (
	"bytes"
	"image/png"
	"strings"
	"testing"

	"github.com/mhsanaei/3x-ui/v3/internal/database/model"
)

func TestShopMenuUsesCustomLabelsAndRoutesThem(test *testing.T) {
	cfg := &model.BotConfig{Presentation: model.BotPresentation{MenuLayout: "compact", HideTrial: true, Labels: map[string]string{"en:buy": "Browse plans"}}}
	kb := shopMenuKeyboard("en", cfg)
	if kb.Keyboard[0][0].Text != "Browse plans" || kb.Keyboard[0][1].Text != shopMenuDisplayLabel("en", "subs", cfg) {
		test.Fatalf("wrong compact menu: %+v", kb.Keyboard)
	}
	for _, row := range kb.Keyboard {
		for _, button := range row {
			if button.Text == t("en", "menuTrial") {
				test.Fatal("hidden trial is visible")
			}
		}
	}
	if got := shopMenuAction("en", "Browse plans", cfg); got != "buy" {
		test.Fatalf("custom button routes to %q", got)
	}
	if got := shopMenuAction("en", t("en", "menuTrial"), cfg); got != "" {
		test.Fatalf("stale hidden trial still routes to %q", got)
	}
	if got := shopMenuAction("en", t("en", "menuBuyNew"), cfg); got != "buy" {
		test.Fatalf("old keyboard no longer routes: %q", got)
	}
}

func TestShopWelcomeEscapesBrandAndUsesEnglishGreeting(test *testing.T) {
	cfg := &model.BotConfig{WelcomeMessage: "سلام", Presentation: model.BotPresentation{
		BrandName: "A & B <shop>", WelcomeEN: "Hello <friend>", Announcement: "Summer & winter",
	}}
	got := shopWelcome(langEN, cfg)
	if !strings.Contains(got, "<b>✨ A &amp; B &lt;shop&gt;</b>") ||
		!strings.Contains(got, "Hello &lt;friend&gt;") ||
		!strings.Contains(got, "<b>📣 Announcement</b>\nSummer &amp; winter") ||
		!strings.Contains(got, "👇 Choose a section") {
		test.Fatalf("invalid HTML or wrong welcome: %q", got)
	}
	if strings.Contains(shopWelcome(langFA, cfg), "Hello") {
		test.Fatal("English greeting leaked into Persian")
	}
}

func TestShopButtonsHaveDistinctTelegramColors(test *testing.T) {
	kb := shopMenuKeyboard(langFA, nil)
	expected := map[string]string{
		shopMenuDisplayLabel(langFA, "buy", nil):   "success",
		shopMenuDisplayLabel(langFA, "trial", nil):    "primary",
		shopMenuDisplayLabel(langFA, "invoices", nil): "primary",
		shopMenuDisplayLabel(langFA, "subs", nil):     "success",
		shopMenuDisplayLabel(langFA, "account", nil):  "primary",
		shopMenuDisplayLabel(langFA, "support", nil):  "danger",
		shopMenuDisplayLabel(langFA, "tutorial", nil): "primary",
	}
	for _, row := range kb.Keyboard {
		for _, button := range row {
			style, exists := expected[button.Text]
			if !exists || button.Style != style {
				test.Fatalf("unexpected button style for %q: got %q, want %q", button.Text, button.Style, style)
			}
			delete(expected, button.Text)
		}
	}
	if len(expected) != 0 {
		test.Fatalf("missing styled buttons: %v", expected)
	}
}

func TestShopBannerIsValidPhoto(test *testing.T) {
	config, err := png.DecodeConfig(bytes.NewReader(vpnShopBanner))
	if err != nil {
		test.Fatalf("embedded VPN SHOP banner is not a PNG: %v", err)
	}
	if config.Width != 720 || config.Height != 280 {
		test.Fatalf("unexpected banner dimensions: %dx%d", config.Width, config.Height)
	}
}

// Custom colors affect actual Telegram reply buttons, not only the preview.
func TestShopMenuButtonStylesOverrideDefaultsAndKeepRouting(test *testing.T) {
	cfg := &model.BotConfig{Presentation: model.BotPresentation{
		ButtonStyles: map[string]string{
			"buy":     "danger",
			"trial":   "success",
			"account": "default",
			"support": "primary",
		},
		Labels: map[string]string{"fa:buy": "🛒 خرید ویژه"},
	}}
	kb := shopMenuKeyboard(langFA, cfg)
	styles := make(map[string]string)
	for _, row := range kb.Keyboard {
		for _, button := range row {
			styles[button.Text] = button.Style
		}
	}
	for label, want := range map[string]string{
		"🛒 خرید ویژه":             "danger",
		t(langFA, "menuTrial"):    "success",
		t(langFA, "menuAccount"):  "",
		t(langFA, "menuSupport"):  "primary",
		shopMenuDisplayLabel(langFA, "invoices", cfg): "primary",
	} {
		if got := styles[label]; got != want {
			test.Errorf("button %q style = %q, want %q", label, got, want)
		}
	}
	if action := shopMenuAction(langFA, "🛒 خرید ویژه", cfg); action != "buy" {
		test.Fatalf("custom button label stopped routing: %s", action)
	}
}

func TestExtendedShopAccentButtonsAreSafeAndStillRoute(test *testing.T) {
	for _, tc := range []struct {
		style     string
		indicator string
	}{
		{"pink", "🩷"},
		{"yellow", "🟡"},
		{"purple", "🟣"},
		{"orange", "🟠"},
		{"cyan", "🩵"},
		{"brown", "🟤"},
		{"black", "⚫"},
	} {
		test.Run(tc.style, func(test *testing.T) {
			cfg := &model.BotConfig{Presentation: model.BotPresentation{
				ButtonStyles: map[string]string{"buy": tc.style},
				Labels:       map[string]string{"fa:buy": "خرید سریع"},
			}}
			button := shopMenuKeyboard(langFA, cfg).Keyboard[0][0]
			expected := tc.indicator + " خرید سریع"
			if button.Text != expected {
				test.Fatalf("button label = %q, want %q", button.Text, expected)
			}
			// Extra colors must not be sent as invalid Bot API button styles.
			if button.Style != "" {
				test.Fatalf("unsupported Telegram background style: %q", button.Style)
			}
			if action := shopMenuAction(langFA, button.Text, cfg); action != "buy" {
				test.Fatalf("click stopped routing, got %q", action)
			}
			if action := shopMenuAction(langFA, "خرید سریع", cfg); action != "buy" {
				test.Fatalf("existing button stopped routing, got %q", action)
			}
			// Old keyboards also work after the configured accent changes.
			cfg.Presentation.ButtonStyles["buy"] = "success"
			if action := shopMenuAction(langFA, button.Text, cfg); action != "buy" {
				test.Fatalf("stale keyboard stopped routing, got %q", action)
			}
			cfg.Presentation.HideTrial = true
			if action := shopMenuAction(langFA, tc.indicator+" "+t(langFA, "menuTrial"), cfg); action != "" {
				test.Fatalf("hidden menu action routed: %q", action)
			}
		})
	}
}

func TestExtendedAccentDoesNotDoublePrefixCustomLabel(test *testing.T) {
	cfg := &model.BotConfig{Presentation: model.BotPresentation{
		ButtonStyles: map[string]string{"buy": "purple"},
		Labels:       map[string]string{"en:buy": "🟣 My offers"},
	}}
	btn := shopMenuKeyboard(langEN, cfg).Keyboard[0][0]
	if btn.Text != "🟣 My offers" || shopMenuAction(langEN, btn.Text, cfg) != "buy" {
		test.Fatalf("custom prefixed button did not preserve label or routing: %+v", btn)
	}
}

func TestShopMenuContainsOnlyNativeTelegramButtons(test *testing.T) {
	for _, lang := range []string{langFA, langEN} {
		for _, cfg := range []*model.BotConfig{nil, &model.BotConfig{}} {
			keyboard := shopMenuKeyboard(lang, cfg)
			if len(keyboard.Keyboard) != 4 {
				test.Fatalf("expected four Telegram menu rows, got %d", len(keyboard.Keyboard))
			}
			for _, row := range keyboard.Keyboard {
				for _, button := range row {
					if button.WebApp != nil {
						test.Fatalf("unexpected Mini App button: %+v", button)
					}
				}
			}
			if shopMenuAction(lang, keyboard.Keyboard[0][0].Text, cfg) != "buy" {
				test.Fatal("purchase button stopped routing")
			}
		}
	}
}

func TestShopDefaultAccentsAndBackwardCompatibleRouting(test *testing.T) {
	for _, tc := range []struct{ action, indicator string }{
		{"trial", "🟡"},
		{"invoices", "🩷"},
		{"subs", "🟣"},
		{"account", "🩵"},
		{"tutorial", "🟠"},
	} {
		label := shopMenuDisplayLabel(langFA, tc.action, nil)
		if !strings.HasPrefix(label, tc.indicator+" ") {
			test.Fatalf("missing %s accent on %s: %q", tc.indicator, tc.action, label)
		}
		if got := shopMenuAction(langFA, label, nil); got != tc.action {
			test.Fatalf("new button routes to %q", got)
		}
		if got := shopMenuAction(langFA, shopMenuLabel(langFA, tc.action, nil), nil); got != tc.action {
			test.Fatalf("legacy button routes to %q", got)
		}
	}
	cfg := &model.BotConfig{Presentation: model.BotPresentation{
		ButtonStyles: map[string]string{"trial": "default"},
	}}
	if got := shopMenuDisplayLabel(langFA, "trial", cfg); got != t(langFA, "menuTrial") {
		test.Fatalf("explicit neutral style left an accent: %q", got)
	}
}

func TestShopWelcomeSafeFallbackAndPersianAnnouncement(test *testing.T) {
	got := shopWelcome(langFA, nil)
	if !strings.Contains(got, "<b>✨ VPN SHOP</b>") || !strings.Contains(got, "👇") {
		test.Fatalf("welcome layout missing: %q", got)
	}
	cfg := &model.BotConfig{Presentation: model.BotPresentation{
		Announcement: "<b>Unsafe</b>",
	}}
	got = shopWelcome(langFA, cfg)
	if strings.Contains(got, "<b>Unsafe</b>") || !strings.Contains(got, "&lt;b&gt;Unsafe&lt;/b&gt;") {
		test.Fatalf("HTML escaping broken: %q", got)
	}
}

// Obsolete graphical-store buttons are accepted as input for migration but
// must never be emitted as part of a newly constructed Telegram menu.
func TestRetiredGraphicalShopButtonCannotReappear(test *testing.T) {
	legacy := []string{
		"ورود به فروشگاه گرافیکی",
		"🛍 ورود به فروشگاه گرافیکی",
		"🛍️ ورود به فروشگاه گرافیکی",
		"ورود به فروشگاه گرافیکی 🛍",
		"🛒 ورود به فروشگاه گرافیکی",
		"Open graphical shop",
	}
	for _, text := range legacy {
		if !isRetiredGraphicalShopButton(text) {
			test.Fatalf("retired graphical button not detected: %q", text)
		}
	}
	for _, text := range []string{"🛍 خرید اشتراک جدید", "ورود به فروشگاه", "", "🛒 خرید اشتراک"} {
		if isRetiredGraphicalShopButton(text) {
			test.Fatalf("valid or unrelated action was mistaken for the retired button: %q", text)
		}
	}
	for _, lang := range []string{langFA, langEN} {
		for _, layout := range []string{"spacious", "compact"} {
			cfg := &model.BotConfig{Presentation: model.BotPresentation{
				MenuLayout: layout,
				Labels: map[string]string{lang + ":buy": "🛍 ورود به فروشگاه گرافیکی"},
			}}
			kb := shopMenuKeyboard(lang, cfg)
			if len(kb.Keyboard) == 0 || kb.Keyboard[0][0].Text != t(lang, "menuBuyNew") {
				test.Fatalf("retired button was persisted as a custom buy label: %+v", kb.Keyboard)
			}
			for _, row := range kb.Keyboard {
				for _, button := range row {
					if isRetiredGraphicalShopButton(button.Text) || button.WebApp != nil {
						test.Fatalf("legacy graphical/miniapp button leaked into %s menu: %+v", layout, button)
					}
				}
			}
		}
	}
}
