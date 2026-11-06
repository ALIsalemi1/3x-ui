package resellerbot

import (
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
