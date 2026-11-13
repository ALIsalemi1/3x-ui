package resellerbot

import (
	"fmt"
	"html"
	"strconv"
	"strings"

	"github.com/mhsanaei/3x-ui/v3/internal/database/model"
	"github.com/mhsanaei/3x-ui/v3/internal/web/service/panel"

	"github.com/mymmrac/telego"
	tu "github.com/mymmrac/telego/telegoutil"
)

// adminMenuKeyboard is the owner's own management menu - shown instead of
// the buyer shop menu in the owner's linked chat, so managing the shop
// doesn't require falling back to /orders and the web panel for everything.
func adminMenuButton(label, style string) telego.KeyboardButton {
	button := tu.KeyboardButton(label)
	button.Style = style
	return button
}

func adminMenuKeyboard() *telego.ReplyKeyboardMarkup {
	kb := tu.Keyboard(
		tu.KeyboardRow(adminMenuButton(t(langFA, "adminMenuOrders"), "danger")),
		tu.KeyboardRow(adminMenuButton(t(langFA, "adminMenuCustomers"), "primary"), adminMenuButton(t(langFA, "adminMenuStats"), "success")),
	)
	kb.ResizeKeyboard = true
	kb.IsPersistent = true
	return kb
}

func (inst *botInstance) sendAdminMenu(chatID int64) {
	inst.send(chatID, t(langFA, "adminWelcome"), adminMenuKeyboard())
}

func adminMenuAction(text string) string {
	switch text {
	case t(langFA, "adminMenuOrders"):
		return "orders"
	case t(langFA, "adminMenuCustomers"):
		return "customers"
	case t(langFA, "adminMenuStats"):
		return "stats"
	}
	return ""
}

func (inst *botInstance) handleAdminMenuText(message telego.Message) {
	switch adminMenuAction(message.Text) {
	case "orders":
		inst.sendPendingOrders(message.Chat.ID)
	case "customers":
		inst.sendCustomersSummary(message.Chat.ID)
	case "stats":
		inst.sendOwnerStats(message.Chat.ID)
	default:
		inst.sendAdminMenu(message.Chat.ID)
	}
}

func (inst *botInstance) sendPendingOrders(chatID int64) {
	orders, err := botSvc.ListOrders(inst.ownerId, model.BotOrderPendingReview)
	if err != nil || len(orders) == 0 {
		inst.send(chatID, t(langFA, "noPendingOrders"))
		return
	}
	for _, o := range orders {
		inst.sendOrderReviewCard(chatID, o)
	}
}

func (inst *botInstance) sendCustomersSummary(chatID int64) {
	customers, err := botSvc.ListCustomers(inst.ownerId)
	if err != nil {
		inst.send(chatID, t(langFA, "error"))
		return
	}
	if len(customers) == 0 {
		inst.send(chatID, t(langFA, "noCustomers"))
		return
	}
	text := t(langFA, "customersCount", len(customers))
	limit := min(len(customers), 20)
	for _, c := range customers[:limit] {
		label := c.BuyerUsername
		if label == "" {
			label = strconv.FormatInt(c.BuyerTgId, 10)
		}
		text += fmt.Sprintf("\n👤 @%s - %s", html.EscapeString(label), t(langFA, "money", c.WalletBalanceToman))
	}
	inst.send(chatID, text)
}

func ownerStatsCard(stats *panel.OwnerStats) string {
	return strings.Join([]string{
		"<b>" + t(langFA, "adminStatsTitle") + "</b>",
		"━━━━━━━━━━━━━━",
		t(langFA, "statsRevenueToday", groupedToman(stats.RevenueTodayToman)),
		t(langFA, "statsRevenueMonth", groupedToman(stats.RevenueMonthToman)),
		t(langFA, "statsRevenue", stats.RevenueToman),
		"━━━━━━━━━━━━━━",
		t(langFA, "statsActiveSubs", stats.ActiveSubscriptions),
		t(langFA, "statsCustomers", stats.CustomerCount),
		t(langFA, "statsNewCustomersToday", stats.NewCustomersToday),
		t(langFA, "statsApprovedOrders", stats.ApprovedOrders),
		t(langFA, "statsPendingReview", stats.PendingReview),
	}, "\n")
}

func (inst *botInstance) sendOwnerStats(chatID int64) {
	stats, err := botSvc.GetOwnerStats(inst.ownerId)
	if err != nil {
		inst.send(chatID, t(langFA, "error"))
		return
	}
	inst.send(chatID, ownerStatsCard(stats), tu.InlineKeyboard(
		tu.InlineKeyboardRow(
			shopCatalogButton(t(langFA, "adminRefreshStats"), "admin:stats", "success"),
			shopCatalogButton(t(langFA, "adminMenuOrders"), "admin:orders", "danger"),
		),
	))
}
