package panel

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/mhsanaei/3x-ui/v3/internal/database/model"
)

func TestBotPresentationPersistsWithoutReplacingTokenOrOtherOwners(t *testing.T) {
	setupBotTestDB(t)
	svc := BotService{}
	p := &model.BotPresentation{BrandName: "Salemi", WelcomeEN: "Welcome aboard", Announcement: "New plans", MenuLayout: "compact", DefaultLanguage: "en", SupportURL: "https://t.me/support", Labels: map[string]string{"en:buy": "Browse plans"}, ButtonStyles: map[string]string{"buy": "success", "trial": "danger"}}
	if _, err := svc.SaveConfig(1, BotConfigInput{Token: "saved-token", Presentation: p}); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.SaveConfig(2, BotConfigInput{Presentation: &model.BotPresentation{BrandName: "Other shop"}}); err != nil {
		t.Fatal(err)
	}
	// Older callers omit presentation; updating their connection must retain branding.
	if _, err := svc.SaveConfig(1, BotConfigInput{Enabled: true}); err != nil {
		t.Fatal(err)
	}
	got, err := svc.GetConfig(1)
	if err != nil {
		t.Fatal(err)
	}
	if got.Presentation.BrandName != "Salemi" || got.Presentation.Labels["en:buy"] != "Browse plans" || got.Presentation.ButtonStyles["trial"] != "danger" || got.Token != "saved-token" {
		t.Fatalf("lost saved configuration: %+v", got.Presentation)
	}
	other, err := svc.GetConfig(2)
	if err != nil {
		t.Fatal(err)
	}
	if other.Presentation.BrandName != "Other shop" {
		t.Fatal("modified another owner's shop")
	}
	if other.Presentation.ButtonStyles == nil {
		t.Fatal("legacy shop should return a non-nil buttonStyles object")
	}
}

func TestBotPresentationAcceptsExtendedButtonAccentStyles(t *testing.T) {
	setupBotTestDB(t)
	svc := BotService{}
	for _, style := range []string{"pink", "yellow", "purple", "orange", "cyan", "brown", "black"} {
		cfg, err := svc.SaveConfig(1, BotConfigInput{Presentation: &model.BotPresentation{
			ButtonStyles: map[string]string{"buy": style},
		}})
		if err != nil || cfg.Presentation.ButtonStyles["buy"] != style {
			t.Fatalf("style %q not saved: config=%+v, error=%v", style, cfg, err)
		}
	}
}

func TestBotPresentationRejectsInvalidInputBeforeSaving(t *testing.T) {
	setupBotTestDB(t)
	svc := BotService{}
	for _, tc := range []struct {
		p       model.BotPresentation
		message string
	}{
		{model.BotPresentation{SupportURL: "javascript:alert(1)"}, "support URL must use https"},
		{model.BotPresentation{MenuLayout: "invalid"}, "menu layout must be spacious or compact"},
		{model.BotPresentation{ButtonStyles: map[string]string{"fake": "success"}}, "invalid menu button style key"},
		{model.BotPresentation{ButtonStyles: map[string]string{"buy": "rainbow"}}, "invalid menu button style"},
		{model.BotPresentation{DefaultLanguage: "zz"}, "default language must be auto, fa or en"},
		{model.BotPresentation{BrandName: strings.Repeat("a", 81)}, "brand name must be at most 80 characters"},
		{model.BotPresentation{Labels: map[string]string{"en:buy": "Same", "en:subs": "Same"}}, "menu labels must be unique within each language"},
	} {
		_, err := svc.SaveConfig(1, BotConfigInput{Presentation: &tc.p})
		if err == nil || err.Error() != tc.message {
			t.Fatalf("got %v, want %s", err, tc.message)
		}
	}
	cfg, err := svc.GetConfig(1)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Id != 0 {
		t.Fatal("invalid presentation created a config row")
	}
}

func TestOldMiniAppConfigurationIsDiscarded(t *testing.T) {
	setupBotTestDB(t)
	var p model.BotPresentation
	if err := json.Unmarshal([]byte(`{"brandName":"VPN SHOP","miniAppURL":"https://panel.example/shop/1/"}`), &p); err != nil {
		t.Fatal(err)
	}
	cfg, err := (&BotService{}).SaveConfig(1, BotConfigInput{Presentation: &p})
	if err != nil {
		t.Fatal(err)
	}
	data, err := json.Marshal(cfg.Presentation)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(data), "miniAppURL") {
		t.Fatalf("removed Mini App field was persisted: %s", data)
	}
}
