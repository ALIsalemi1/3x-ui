package resellerbot

import (
	"strings"
	"testing"
	"time"

	"github.com/mhsanaei/3x-ui/v3/internal/database/model"
	"github.com/mhsanaei/3x-ui/v3/internal/web/service/panel"
)

func TestQuotaUsageBarColorsAndBounds(t *testing.T) {
	for _, tc := range []struct {
		used, total int64
		wantPercent int
		wantSegment string
	}{
		{0, 100, 0, "⬜"},
		{25, 100, 25, "🟩"},
		{78, 100, 78, "🟨"},
		{95, 100, 95, "🟥"},
		{200, 100, 100, "🟥"},
	} {
		bar, percent := quotaUsageBar(tc.used, tc.total)
		if percent != tc.wantPercent || !strings.Contains(bar, tc.wantSegment) {
			t.Errorf("quotaUsageBar(%d,%d) = %q,%d", tc.used, tc.total, bar, percent)
		}
	}
	if bar, percent := quotaUsageBar(10, 0); bar != "" || percent != 0 {
		t.Fatalf("unlimited data should not have a percent bar: %q,%d", bar, percent)
	}
}

func TestSubscriptionListButtonsUseRealStatus(t *testing.T) {
	sub := model.BotSubscription{Id: 41, ClientEmail: "alice"}
	now := time.Now().Add(24 * time.Hour).UnixMilli()
	for _, tc := range []struct {
		name   string
		stats  *panel.SubscriptionStats
		prefix string
		style  string
	}{
		{"unknown", nil, "⚪", "primary"},
		{"disabled", &panel.SubscriptionStats{Enable: false}, "🔴", "danger"},
		{"expired", &panel.SubscriptionStats{Enable: true, ExpiryTime: time.Now().Add(-time.Hour).UnixMilli()}, "🟠", "danger"},
		{"active", &panel.SubscriptionStats{Enable: true, ExpiryTime: now}, "🟢", "success"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			button := subscriptionListButton(sub, "Premium", tc.stats)
			if !strings.HasPrefix(button.Text, tc.prefix) ||
				button.Style != tc.style || button.CallbackData != "sub:view:41" {
				t.Fatalf("wrong %s button: %+v", tc.name, button)
			}
		})
	}
	if got := clientStatusLabel(langFA, nil); !strings.Contains(got, "نامشخص") {
		t.Fatalf("unknown usage incorrectly says disabled: %q", got)
	}
}

func TestSubscriptionCardPlanAndExpiryWarning(t *testing.T) {
	sub := &model.BotSubscription{Id: 4, ClientEmail: "a<b>"}
	stats := &panel.SubscriptionStats{
		Enable:     true,
		ExpiryTime: time.Now().Add(48 * time.Hour).UnixMilli(),
		Up:         10, Down: 10, Total: 100,
	}
	text, _ := subscriptionCardWithDetails(sub, stats, "", "Pro <Plus>", langEN)
	for _, want := range []string{"Pro &lt;Plus&gt;", "a&lt;b&gt;", "20%", "Less than 3 days"} {
		if !strings.Contains(text, want) {
			t.Errorf("missing %q: %s", want, text)
		}
	}
	// No false warning on unlimited traffic and no exposed raw HTML.
	stats.Total = 0
	text, _ = subscriptionCardWithDetails(sub, stats, "", "", langFA)
	if !strings.Contains(text, "حجم نامحدود") || strings.Contains(text, "20%") {
		t.Fatalf("unlimited traffic rendered incorrectly: %s", text)
	}
}

func TestConfigDeliveryUsesKnownProtocolAndEscapesLink(t *testing.T) {
	for _, tc := range []struct{ link, want string }{
		{"vless://example", "VLESS"},
		{"vmess://example", "VMess"},
		{"ss://example", "Shadowsocks"},
		{"trojan://example", "Trojan"},
		{"hy2://example", "Hysteria2"},
		{"https://example", "Link"},
	} {
		got := configProtocolName(tc.link)
		if got != tc.want {
			t.Errorf("%q protocol = %q, want %q", tc.link, got, tc.want)
		}
	}
	text := configDeliveryItem(langFA, 2, "vless://abc?x=1&z=<tag>")
	if !strings.Contains(text, "کانفیگ 2 · VLESS") ||
		!strings.Contains(text, "<code>vless://abc?x=1&amp;z=&lt;tag&gt;</code>") {
		t.Fatalf("config delivery formatting or HTML escaping failed: %s", text)
	}
}
