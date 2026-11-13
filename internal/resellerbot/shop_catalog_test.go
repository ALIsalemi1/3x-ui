package resellerbot

import (
	"bytes"
	"image/png"
	"strings"
	"testing"

	"github.com/mhsanaei/3x-ui/v3/internal/database/model"
)

func TestShopCatalogFiltersAndSortsRealPlans(t *testing.T) {
	plans := []model.BotPlan{
		{Id: 8, Enabled: true, TrafficGB: 70, DurationDays: 60, PriceToman: 420000},
		{Id: 6, Enabled: false, TrafficGB: 1, DurationDays: 30, PriceToman: 10000},
		{Id: 10, Enabled: true, IsTrial: true, TrafficGB: 2, DurationDays: 1},
		{Id: 5, Enabled: true, TrafficGB: 5, DurationDays: 30, PriceToman: 30000},
	}
	visible := visibleShopPlans(plans)
	if len(visible) != 2 || visible[0].Id != 5 || visible[1].Id != 8 {
		t.Fatalf("expected enabled, non-trial plans sorted by quota; got %+v", visible)
	}
	keyboard, page, total := shopCatalogKeyboard(langFA, visible, 0)
	if page != 0 || total != 1 || len(keyboard.InlineKeyboard) != 4 {
		t.Fatalf("unexpected catalog shape: page=%d total=%d rows=%d", page, total, len(keyboard.InlineKeyboard))
	}
	for _, button := range keyboard.InlineKeyboard[1] {
		if button.CallbackData != "buy:5" {
			t.Fatalf("all three plan cells must buy plan 5, got %q", button.CallbackData)
		}
	}
	if got := keyboard.InlineKeyboard[1][2].Text; got != "30,000 تومان" {
		t.Fatalf("price must reflect the actual configured amount, got %q", got)
	}
	if keyboard.InlineKeyboard[3][0].CallbackData != "shop:back" {
		t.Fatal("back button is missing")
	}
}

func TestShopCatalogPagingKeepsCallbackIDsAndNavigation(t *testing.T) {
	plans := make([]model.BotPlan, 0, 11)
	for i := 1; i <= 11; i++ {
		plans = append(plans, model.BotPlan{
			Id: i * 17, Enabled: true, TrafficGB: int64(i * 5),
			DurationDays: 30, PriceToman: int64(i * 10000),
		})
	}
	plans = visibleShopPlans(plans)
	first, page, total := shopCatalogKeyboard(langEN, plans, 0)
	if page != 0 || total != 2 || len(first.InlineKeyboard) != 12 {
		t.Fatalf("first page expected 9 plans, next and back: page=%d total=%d rows=%d", page, total, len(first.InlineKeyboard))
	}
	if first.InlineKeyboard[10][0].CallbackData != "shop:page:1" {
		t.Fatalf("missing next-page callback: %+v", first.InlineKeyboard[10])
	}
	if first.InlineKeyboard[7][0].Style != "success" || first.InlineKeyboard[8][0].Style != "danger" {
		t.Fatal("featured and premium plan rows need distinct native Telegram colors")
	}
	second, page, total := shopCatalogKeyboard(langEN, plans, 99)
	if page != 1 || total != 2 || len(second.InlineKeyboard) != 5 {
		t.Fatalf("second page expected 2 plans, previous and back: page=%d total=%d rows=%d", page, total, len(second.InlineKeyboard))
	}
	for _, button := range second.InlineKeyboard[1] {
		if button.CallbackData != "buy:170" {
			t.Fatalf("second page must use actual plan ID, got %q", button.CallbackData)
		}
	}
	if second.InlineKeyboard[3][0].CallbackData != "shop:page:0" {
		t.Fatal("previous-page callback is missing")
	}
}

func TestShopCatalogPhotoAndLocalizedCaption(t *testing.T) {
	config, err := png.DecodeConfig(bytes.NewReader(shopLocationsBanner))
	if err != nil || config.Width != 880 || config.Height != 340 {
		t.Fatalf("shop banner must decode at 880x340: %v, %+v", err, config)
	}
	persian := shopCatalogCaption(langFA, 0, 2)
	english := shopCatalogCaption(langEN, 1, 2)
	for _, label := range []string{"VPN SHOP", "آلمان", "هلند", "آلبانی", "صفحهٔ 1 از 2"} {
		if !strings.Contains(persian, label) {
			t.Fatalf("Persian catalog missing %q: %q", label, persian)
		}
	}
	if !strings.Contains(english, "Germany") || !strings.Contains(english, "Page 2 of 2") {
		t.Fatalf("English catalog language or page count incorrect: %q", english)
	}
}
