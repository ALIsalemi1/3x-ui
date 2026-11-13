package resellerbot

import (
	"html"
	"strconv"
	"time"

	"github.com/mhsanaei/3x-ui/v3/internal/database/model"
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
		name := ""
		if plan, pErr := botSvc.GetPlan(inst.ownerId, sub.PlanId); pErr == nil && plan != nil {
			name = plan.Name
		}
		stats, _ := botSvc.GetSubscriptionStats(sub.ClientEmail)
		rows = append(rows, tu.InlineKeyboardRow(subscriptionListButton(sub, name, stats)))
	}
	caption := "<b>" + t(lang, "menuSubs") + "</b>\n━━━━━━━━━━━━━━\n" +
		t(lang, "subscriptionListCount", len(subs)) + "\n" +
		t(lang, "subscriptionListHint")
	inst.send(chatID, caption, tu.InlineKeyboard(rows...))
}

// clientStatusLabel collapses enable/expiry into the same three states the
// subscription-page/subscription bot card shows.
func clientStatusLabel(lang string, stats *panel.SubscriptionStats) string {
	switch {
	case stats == nil:
		return t(lang, "clientStatusUnknown")
	case !stats.Enable:
		return t(lang, "clientStatusDisabled")
	case stats.ExpiryTime > 0 && stats.ExpiryTime/1000 < time.Now().Unix():
		return t(lang, "clientStatusExpired")
	default:
		return t(lang, "clientStatusActive")
	}
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
		inst.send(chatID, t(lang, "configDeliveryEmpty"))
		return
	}
	inst.send(chatID,
		"<b>"+t(lang, "configDeliveryTitle")+"</b>\n━━━━━━━━━━━━━━\n"+
			t(lang, "subUsername", "<code>"+html.EscapeString(sub.ClientEmail)+"</code>")+"\n"+
			t(lang, "configDeliveryCount", len(links))+"\n\n"+
			t(lang, "configDeliveryHint"))
	for i, link := range links {
		inst.send(chatID, configDeliveryItem(lang, i+1, link))
	}
	id := strconv.Itoa(sub.Id)
	inst.send(chatID, t(lang, "configDeliveryFooter"), tu.InlineKeyboard(
		tu.InlineKeyboardRow(
			shopCatalogButton(t(lang, "btnSubDetails"), "sub:view:"+id, "primary"),
			shopCatalogButton(t(lang, "btnBackToSubs"), "sub:list:0", ""),
		),
	))
}

func (inst *botInstance) handleRevokePrompt(chatID int64, subIDStr, lang string) {
	subID, err := strconv.Atoi(subIDStr)
	if err != nil {
		return
	}
	if _, err := botSvc.GetSubscription(inst.ownerId, subID); err != nil {
		inst.send(chatID, t(lang, "error"))
		return
	}
	inst.send(chatID, "<b>"+t(lang, "revokeTitle")+"</b>\n\n"+t(lang, "revokeConfirm"),
		tu.InlineKeyboard(
			tu.InlineKeyboardRow(shopCatalogButton(t(lang, "btnConfirmRevoke"), "sub:revokeconfirm:"+subIDStr, "danger")),
			tu.InlineKeyboardRow(shopCatalogButton(t(lang, "btnCancelRevoke"), "sub:view:"+subIDStr, "primary")),
		))
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
	inst.sendSubscriptionDetail(chatID, subID, lang)
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
