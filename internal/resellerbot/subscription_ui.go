package resellerbot

import (
	"html"
	"strconv"
	"strings"
	"time"

	"github.com/mhsanaei/3x-ui/v3/internal/database/model"
	"github.com/mhsanaei/3x-ui/v3/internal/util/common"
	"github.com/mhsanaei/3x-ui/v3/internal/web/service/panel"

	"github.com/mymmrac/telego"
	tu "github.com/mymmrac/telego/telegoutil"
)

// quotaUsageBar returns a 10-block visual meter and the clamped percent used.
func quotaUsageBar(used, total int64) (string, int) {
	if total <= 0 {
		return "", 0
	}
	if used < 0 {
		used = 0
	}
	percent := int(100*float64(used)/float64(total) + 0.5)
	if percent > 100 {
		percent = 100
	}
	filled := (percent + 5) / 10
	if filled > 10 {
		filled = 10
	}
	color := "🟩"
	if percent >= 90 {
		color = "🟥"
	} else if percent >= 75 {
		color = "🟨"
	}
	return strings.Repeat(color, filled) + strings.Repeat("⬜", 10-filled), percent
}

// subscriptionListButton gives buyers a quick, accurate status signal.
// Missing traffic stats must not be presented as a disabled service.
func subscriptionListButton(sub model.BotSubscription, planName string, stats *panel.SubscriptionStats) telego.InlineKeyboardButton {
	name := strings.TrimSpace(planName)
	if name == "" {
		name = sub.ClientEmail
	}
	runes := []rune(name)
	if len(runes) > 38 {
		name = string(runes[:37]) + "…"
	}
	icon, style := "⚪", "primary"
	switch {
	case stats == nil:
		// State is unknown.
	case !stats.Enable:
		icon, style = "🔴", "danger"
	case stats.ExpiryTime > 0 && stats.ExpiryTime/1000 <= time.Now().Unix():
		icon, style = "🟠", "danger"
	default:
		icon, style = "🟢", "success"
	}
	return shopCatalogButton(icon+" "+name, "sub:view:"+strconv.Itoa(sub.Id), style)
}

// subscriptionCardWithDetails is the unified subscription dashboard for
// purchases, renewals, wallet settlement and the "My subscriptions" menu.
func subscriptionCardWithDetails(sub *model.BotSubscription, stats *panel.SubscriptionStats, subURL, planName, lang string) (string, *telego.InlineKeyboardMarkup) {
	lines := []string{
		"<b>" + t(lang, "subscriptionCardTitle") + "</b>",
		"━━━━━━━━━━━━━━",
	}
	if planName != "" {
		lines = append(lines, t(lang, "subPlan", html.EscapeString(planName)))
	}
	lines = append(lines,
		"<b>"+t(lang, "subStatus", clientStatusLabel(lang, stats))+"</b>",
		t(lang, "subUsername", "<code>"+html.EscapeString(sub.ClientEmail)+"</code>"),
	)
	if stats != nil {
		used := stats.Up + stats.Down
		if used < 0 {
			used = 0
		}
		remaining := t(lang, "unlimited")
		lines = append(lines, "", "<b>"+t(lang, "subscriptionUsageTitle")+"</b>")
		if stats.Total > 0 {
			left := stats.Total - used
			if left < 0 {
				left = 0
			}
			remaining = common.FormatTraffic(left)
			bar, percent := quotaUsageBar(used, stats.Total)
			lines = append(lines, bar+"  "+strconv.Itoa(percent)+"%")
			if percent >= 90 && stats.Enable {
				lines = append(lines, t(lang, "subLowTrafficWarning"))
			}
		} else {
			lines = append(lines, t(lang, "subUnlimitedQuota"))
		}
		lines = append(lines,
			t(lang, "subUsed", common.FormatTraffic(used)),
			t(lang, "subRemainingQuota", remaining),
			t(lang, "subRemaining", timeLeftLabel(stats.ExpiryTime, lang)),
		)
		if stats.Enable && stats.ExpiryTime > 0 {
			secondsLeft := stats.ExpiryTime/1000 - time.Now().Unix()
			if secondsLeft > 0 && secondsLeft <= 72*3600 {
				lines = append(lines, t(lang, "subExpiryWarning"))
			}
		}
	}
	if subURL != "" {
		lines = append(lines,
			"",
			"<b>"+t(lang, "subscriptionLinkTitle")+"</b>",
			"<code>"+html.EscapeString(subURL)+"</code>",
			t(lang, "subscriptionLinkHint"),
		)
	}
	autoLabel, autoStyle := "autoRenewOff", "primary"
	if sub.AutoRenewEnabled {
		autoLabel, autoStyle = "autoRenewOn", "success"
	}
	id := strconv.Itoa(sub.Id)
	rows := [][]telego.InlineKeyboardButton{
		tu.InlineKeyboardRow(
			shopCatalogButton(t(lang, "btnGetConfigs"), "sub:configs:"+id, "primary"),
			shopCatalogButton(t(lang, "btnRenew"), "sub:renew:"+id, "success"),
		),
		tu.InlineKeyboardRow(
			shopCatalogButton(t(lang, "btnRefreshSub"), "sub:refresh:"+id, "primary"),
			shopCatalogButton(t(lang, autoLabel), "sub:autorenew:"+id, autoStyle),
		),
		tu.InlineKeyboardRow(shopCatalogButton(t(lang, "btnRevoke"), "sub:revoke:"+id, "danger")),
		tu.InlineKeyboardRow(shopCatalogButton(t(lang, "btnBackToSubs"), "sub:list:0", "")),
	}
	return strings.Join(lines, "\n"), tu.InlineKeyboard(rows...)
}

func subscriptionCardWithStats(sub *model.BotSubscription, stats *panel.SubscriptionStats, subURL, lang string) (string, *telego.InlineKeyboardMarkup) {
	return subscriptionCardWithDetails(sub, stats, subURL, "", lang)
}

func subscriptionCard(sub *model.BotSubscription, lang string) (string, *telego.InlineKeyboardMarkup) {
	stats, _ := botSvc.GetSubscriptionStats(sub.ClientEmail)
	subURL, _ := botSvc.GetSubscriptionConfigs(sub.ClientEmail)
	planName := ""
	if plan, err := botSvc.GetPlan(sub.OwnerId, sub.PlanId); err == nil && plan != nil {
		planName = plan.Name
	}
	return subscriptionCardWithDetails(sub, stats, subURL, planName, lang)
}

// Protocol names are allowlisted; arbitrary link strings never become markup.
func configProtocolName(link string) string {
	prefix, _, ok := strings.Cut(strings.ToLower(strings.TrimSpace(link)), ":")
	if !ok {
		return "Link"
	}
	switch prefix {
	case "vless":
		return "VLESS"
	case "vmess":
		return "VMess"
	case "trojan":
		return "Trojan"
	case "ss":
		return "Shadowsocks"
	case "hysteria2", "hy2":
		return "Hysteria2"
	case "tuic":
		return "TUIC"
	case "wireguard":
		return "WireGuard"
	default:
		return "Link"
	}
}

func configDeliveryItem(lang string, number int, link string) string {
	return "<b>" + t(lang, "configDeliveryEntry", number, configProtocolName(link)) + "</b>\n" +
		"<code>" + html.EscapeString(link) + "</code>"
}
