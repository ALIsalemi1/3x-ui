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
	if kb.Keyboard[0][0].Text != "Browse plans" || kb.Keyboard[0][1].Text != t("en", "menuSubs") {
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
	cfg := &model.BotConfig{WelcomeMessage: "سلام", Presentation: model.BotPresentation{BrandName: "A & B <shop>", WelcomeEN: "Hello <friend>", Announcement: "Summer & winter"}}
	got := shopWelcome("en", cfg)
	if got != "<b>A &amp; B &lt;shop&gt;</b>\n\nHello &lt;friend&gt;\n\nSummer &amp; winter" {
		test.Fatalf("invalid HTML or wrong welcome: %q", got)
	}
	if strings.Contains(shopWelcome("fa", cfg), "Hello") {
		test.Fatal("English greeting leaked into Persian")
	}
}

func TestShopButtonsHaveDistinctTelegramColors(test *testing.T) {
	kb := shopMenuKeyboard(langFA, nil)
	expected := map[string]string{
		t(langFA, "menuBuyNew"):   "success",
		t(langFA, "menuTrial"):    "primary",
		t(langFA, "menuInvoices"): "primary",
		t(langFA, "menuSubs"):     "success",
		t(langFA, "menuAccount"):  "primary",
		t(langFA, "menuSupport"):  "danger",
		t(langFA, "menuTutorial"): "primary",
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
		t(langFA, "menuInvoices"): "primary",
	} {
		if got := styles[label]; got != want {
			test.Errorf("button %q style = %q, want %q", label, got, want)
		}
	}
	if action := shopMenuAction(langFA, "🛒 خرید ویژه", cfg); action != "buy" {
		test.Fatalf("custom button label stopped routing: %s", action)
	}
}
