package resellerbot

import (
	"strings"
	"testing"
	"time"

	"github.com/mhsanaei/3x-ui/v3/internal/database/model"
	"github.com/mhsanaei/3x-ui/v3/internal/web/service/panel"
)

func TestCatalogSortingAndFilterNavigation(t *testing.T) {
	plans := []model.BotPlan{
		{Id: 1, PriceToman: 300000, TrafficGB: 20, DurationDays: 90},
		{Id: 2, PriceToman: 100000, TrafficGB: 10, DurationDays: 30},
		{Id: 3, PriceToman: 200000, TrafficGB: 0, DurationDays: 60},
	}
	for _, tc := range []struct {
		mode string
		firstID int
	}{
		{"price", 2},
		{"traffic", 3},
		{"duration", 2},
	} {
		sorted := shopSortPlans(plans, tc.mode)
		if sorted[0].Id != tc.firstID {
			t.Errorf("%s sort first id = %d, want %d", tc.mode, sorted[0].Id, tc.firstID)
		}
	}
	keyboard, page, total := shopCatalogKeyboardMode(langFA, plans, 0, "price")
	if keyboard == nil || page != 0 || total != 1 ||
		keyboard.InlineKeyboard[3][0].CallbackData != "buy:2" ||
		keyboard.InlineKeyboard[0][1].CallbackData != "shop:filter:price" ||
		keyboard.InlineKeyboard[0][1].Style != "success" {
		t.Fatalf("catalog filter or buying navigation incorrect: %+v", keyboard)
	}
	many := make([]model.BotPlan, 10)
	for i := range many {
		many[i] = model.BotPlan{Id: i+1, TrafficGB: int64(i+1), PriceToman: int64(i+1) * 1000}
	}
	keyboard, _, _ = shopCatalogKeyboardMode(langEN, many, 0, "traffic")
	foundNext := false
	for _, row := range keyboard.InlineKeyboard {
		for _, button := range row {
			if button.CallbackData == "shop:page:traffic:1" {
				foundNext = true
			}
		}
	}
	if !foundNext {
		t.Fatal("sort mode was lost when paging")
	}
}

func TestExpiryAlertStagesAndRestartDeduplication(t *testing.T) {
	now := time.Now().UnixMilli()
	sub := model.BotSubscription{}
	for _, tc := range []struct {
		remaining time.Duration
		wantStage int
	}{
		{96*time.Hour, 0},
		{71*time.Hour, 1},
		{20*time.Hour, 2},
		{5*time.Hour, 3},
		{-time.Hour, 0},
	} {
		expiry := now+tc.remaining.Milliseconds()
		if got := expiryAlertStage(expiry, now); got != tc.wantStage {
			t.Errorf("%s alert stage = %d, want %d", tc.remaining, got, tc.wantStage)
		}
	}
	expiry := now+(5*time.Hour).Milliseconds()
	if stage, send := expiryAlertDue(sub, expiry, now); !send || stage != 3 {
		t.Fatal("due reminder was not scheduled")
	}
	sub.LastExpiryAlertFor = expiry
	sub.LastExpiryAlertStage = 3
	if _, send := expiryAlertDue(sub, expiry, now); send {
		t.Fatal("repeat reminder sent after restart")
	}
	if _, send := expiryAlertDue(sub, expiry+(24*time.Hour).Milliseconds(), now); !send {
		t.Fatal("renewed expiry did not re-arm reminder")
	}
}

func TestLowQuotaAlertCooldownAndOwnerPreferences(t *testing.T) {
	now := time.Now().UnixMilli()
	stats := &panel.SubscriptionStats{Enable: true, Up: 90, Down: 0, Total: 100}
	sub := model.BotSubscription{}
	if !quotaAlertDue(sub, stats, now) {
		t.Fatal("90% usage should trigger low-data alert")
	}
	sub.LastLowQuotaAlertAt = now
	if quotaAlertDue(sub, stats, now) {
		t.Fatal("quota alert repeated immediately")
	}
	if !quotaAlertDue(sub, stats, now+(25*time.Hour).Milliseconds()) {
		t.Fatal("quota alert did not re-arm after 24h")
	}
	stats.Enable = false
	if quotaAlertDue(sub, stats, now+(25*time.Hour).Milliseconds()) {
		t.Fatal("disabled service must not generate quota alerts")
	}
}

func TestAdminStatsAndReferralCardEscaping(t *testing.T) {
	stats := &panel.OwnerStats{
		RevenueTodayToman: 120000,
		RevenueMonthToman: 1200000,
		ActiveSubscriptions: 12,
		NewCustomersToday: 3,
	}
	text := ownerStatsCard(stats)
	for _, wanted := range []string{"120,000", "1,200,000", "12", "3", "VPN SHOP"} {
		if !strings.Contains(text, wanted) {
			t.Errorf("missing %s in owner dashboard: %s", wanted, text)
		}
	}
	info := referralCardText(langFA, "https://t.me/bot?start=ref_a&b", 2, 45000, 50000,
		[]string{"<bad>", "ali&reza"})
	if !strings.Contains(info, "&amp;b") || !strings.Contains(info, "&lt;bad&gt;") ||
		!strings.Contains(info, "ali&amp;reza") {
		t.Fatalf("unsafe referral HTML: %s", info)
	}
	adminKB := adminMenuKeyboard()
	if adminKB.Keyboard[0][0].Style != "danger" || adminKB.Keyboard[1][1].Style != "success" {
		t.Fatal("admin buttons lost their native Telegram colors")
	}
}
