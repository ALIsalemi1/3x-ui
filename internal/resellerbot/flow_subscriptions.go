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

// timeLeftLabel renders remaining time as "Xd Xh" ("∞" unlimited, "0" past
// expiry) - a short local variant of internal/sub's richer formatter, not
// worth sharing across packages for this much logic.
func timeLeftLabel(expiryMs int64, lang string) string {
	if expiryMs == 0 {
		return "∞"
	}
	exp := expiryMs / 1000
	var secs int64
	if exp > 0 {
		secs = exp - time.Now().Unix()
	} else {
		secs = -exp
	}
	if secs <= 0 {
		return "0"
	}
	days := secs / 86400
	hours := (secs % 86400) / 3600
	if days > 0 {
		return t(lang, "daysHours", days, hours)
	}
	return t(lang, "hours", hours)
}

func (inst *botInstance) sendSubscriptionList(chatID, buyerTgID int64, buyerUsername, lang string) {
	customer, err := botSvc.GetOrCreateCustomer(inst.ownerId, buyerTgID, buyerUsername)
	if err != nil {
		inst.send(chatID, t(lang, "error"))
		return
	}
	subs, err := botSvc.ListSubscriptions(inst.ownerId, customer.Id)
	if err != nil || len(subs) == 0 {
		inst.send(chatID, t(lang, "subListEmpty"))
		return
	}
	var rows [][]telego.InlineKeyboardButton
	for _, sub := range subs {
		label := sub.ClientEmail
		if plan, pErr := botSvc.GetPlan(inst.ownerId, sub.PlanId); pErr == nil {
			label = plan.Name
		}
		rows = append(rows, tu.InlineKeyboardRow(shopCatalogButton("💎 "+label, "sub:view:"+strconv.Itoa(sub.Id), "primary")))
	}
	caption := "<b>" + t(lang, "menuSubs") + "</b>\n━━━━━━━━━━━━━━\n" + t(lang, "subscriptionListHint")
	inst.send(chatID, caption, tu.InlineKeyboard(rows...))
}

// clientStatusLabel collapses enable/expiry into the same three states the
// subscription-page/subscription bot card shows.
func clientStatusLabel(lang string, stats *panel.SubscriptionStats) string {
	switch {
	case stats == nil || !stats.Enable:
		return t(lang, "clientStatusDisabled")
	case stats.ExpiryTime > 0 && stats.ExpiryTime/1000 < time.Now().Unix():
		return t(lang, "clientStatusExpired")
	default:
		return t(lang, "clientStatusActive")
	}
}

// subscriptionCardWithStats assembles an HTML-safe detail card with the
// existing callback IDs. A separate renderer keeps live DB access out of tests.
func subscriptionCardWithStats(sub *model.BotSubscription, stats *panel.SubscriptionStats, subURL, lang string) (string, *telego.InlineKeyboardMarkup) {
	lines := []string{
		"<b>" + t(lang, "subscriptionCardTitle") + "</b>",
		"━━━━━━━━━━━━━━",
		t(lang, "subStatus", clientStatusLabel(lang, stats)),
		t(lang, "subUsername", html.EscapeString(sub.ClientEmail)),
	}
	if stats != nil {
		used := common.FormatTraffic(stats.Up + stats.Down)
		remaining := t(lang, "unlimited")
		if stats.Total > 0 {
			left := stats.Total - stats.Up - stats.Down
			if left < 0 {
				left = 0
			}
			remaining = common.FormatTraffic(left)
		}
		lines = append(lines,
			"━━━━━━━━━━━━━━",
			t(lang, "subUsed", used),
			t(lang, "subRemainingQuota", remaining),
			t(lang, "subRemaining", timeLeftLabel(stats.ExpiryTime, lang)))
	}
	if subURL != "" {
		lines = append(lines, "━━━━━━━━━━━━━━", t(lang, "subLink", "<code>"+html.EscapeString(subURL)+"</code>"))
	}
	autoLabel := "autoRenewOff"
	autoStyle := "primary"
	if sub.AutoRenewEnabled {
		autoLabel = "autoRenewOn"
		autoStyle = "success"
	}
	id := strconv.Itoa(sub.Id)
	rows := [][]telego.InlineKeyboardButton{
		tu.InlineKeyboardRow(
			shopCatalogButton(t(lang, "btnGetConfigs"), "sub:configs:"+id, "primary"),
			shopCatalogButton(t(lang, "btnRenew"), "sub:renew:"+id, "success"),
		),
		tu.InlineKeyboardRow(shopCatalogButton(t(lang, autoLabel), "sub:autorenew:"+id, autoStyle)),
		tu.InlineKeyboardRow(shopCatalogButton(t(lang, "btnRevoke"), "sub:revoke:"+id, "danger")),
	}
	return strings.Join(lines, "\n"), tu.InlineKeyboard(rows...)
}

// subscriptionCard is also used after wallet payment and order approval.
func subscriptionCard(sub *model.BotSubscription, lang string) (string, *telego.InlineKeyboardMarkup) {
	stats, _ := botSvc.GetSubscriptionStats(sub.ClientEmail)
	subURL, _ := botSvc.GetSubscriptionConfigs(sub.ClientEmail)
	return subscriptionCardWithStats(sub, stats, subURL, lang)
}

func (inst *botInstance) sendSubscriptionDetail(chatID int64, subID int, lang string) {
	sub, err := botSvc.GetSubscription(inst.ownerId, subID)
	if err != nil {
		inst.send(chatID, t(lang, "error"))
		return
	}
	text, kb := subscriptionCard(sub, lang)
	inst.send(chatID, text, kb)
}

func (inst *botInstance) handleGetConfigs(chatID int64, subIDStr, lang string) {
	subID, err := strconv.Atoi(subIDStr)
	if err != nil {
		return
	}
	sub, err := botSvc.GetSubscription(inst.ownerId, subID)
	if err != nil {
		inst.send(chatID, t(lang, "error"))
		return
	}
	links, err := botSvc.GetClientConfigLinks(sub.ClientEmail)
	if err != nil || len(links) == 0 {
		inst.send(chatID, t(lang, "error"))
		return
	}
	for _, link := range links {
		inst.send(chatID, "<code>"+html.EscapeString(link)+"</code>")
	}
}

func (inst *botInstance) handleRevokePrompt(chatID int64, subIDStr, lang string) {
	inst.send(chatID, t(lang, "revokeConfirm"), tu.InlineKeyboard(tu.InlineKeyboardRow(
		tu.InlineKeyboardButton(t(lang, "btnRevoke")).WithCallbackData("sub:revokeconfirm:"+subIDStr),
	)))
}

func (inst *botInstance) handleRevokeConfirm(chatID int64, subIDStr, lang string) {
	subID, err := strconv.Atoi(subIDStr)
	if err != nil {
		return
	}
	if _, err := botSvc.RevokeLink(inst.ownerId, subID); err != nil {
		inst.send(chatID, t(lang, "error"))
		return
	}
	inst.send(chatID, t(lang, "revokeDone"))
}

func (inst *botInstance) handleToggleAutoRenew(chatID int64, subIDStr, lang string) {
	subID, err := strconv.Atoi(subIDStr)
	if err != nil {
		return
	}
	sub, err := botSvc.GetSubscription(inst.ownerId, subID)
	if err != nil {
		inst.send(chatID, t(lang, "error"))
		return
	}
	if err := botSvc.SetAutoRenew(inst.ownerId, subID, !sub.AutoRenewEnabled); err != nil {
		inst.send(chatID, t(lang, "error"))
		return
	}
	inst.send(chatID, t(lang, "autoRenewToggled"))
	inst.sendSubscriptionDetail(chatID, subID, lang)
}
