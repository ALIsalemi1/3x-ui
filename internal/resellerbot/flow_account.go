package resellerbot

import (
	"fmt"
	"strings"
	"time"

	"github.com/mhsanaei/3x-ui/v3/internal/util/jalali"

	"github.com/mymmrac/telego"
	tu "github.com/mymmrac/telego/telegoutil"
)

// accountCardText displays the wallet prominently without exposing any
// unescaped user-provided content or changing account state.
func accountCardText(lang string, buyerTgID, balance int64, joined string) string {
	return strings.Join([]string{
		"<b>" + t(lang, "menuAccount") + "</b>",
		"━━━━━━━━━━━━━━",
		t(lang, "accountId", buyerTgID),
		"",
		"<b>" + t(lang, "walletCardTitle") + "</b>",
		t(lang, "walletBalancePretty", groupedToman(balance)),
		"━━━━━━━━━━━━━━",
		t(lang, "accountJoined", joined),
	}, "\n")
}

func (inst *botInstance) sendAccountCard(chatID, buyerTgID int64, buyerUsername, lang string) {
	customer, err := botSvc.GetOrCreateCustomer(inst.ownerId, buyerTgID, buyerUsername)
	if err != nil {
		inst.send(chatID, t(lang, "error"))
		return
	}
	joined := "-"
	if customer.CreatedAt > 0 {
		tm := time.UnixMilli(customer.CreatedAt).In(time.Local)
		y, m, d := jalali.ToJalali(tm.Year(), int(tm.Month()), tm.Day())
		joined = fmt.Sprintf("%d/%02d/%02d", y, m, d)
	}

	text := accountCardText(lang, customer.BuyerTgId, customer.WalletBalanceToman, joined)

	notifLabel := "notificationsOff"
	if customer.NotificationsEnabled {
		notifLabel = "notificationsOn"
	}

	var rows [][]telego.InlineKeyboardButton
	cfg, _ := botSvc.GetConfig(inst.ownerId)
	if cfg != nil && cfg.WalletEnabled {
		rows = append(rows, tu.InlineKeyboardRow(
			shopCatalogButton(t(lang, "btnTopup"), "acct:topup", "success"),
			shopCatalogButton(t(lang, "btnRedeem"), "acct:redeem", "primary"),
		))
	}
	notifStyle := "primary"
	if customer.NotificationsEnabled {
		notifStyle = "success"
	}
	rows = append(rows, tu.InlineKeyboardRow(shopCatalogButton(t(lang, notifLabel), "acct:notif", notifStyle)))
	if cfg != nil && cfg.ReferralEnabled {
		rows = append(rows, tu.InlineKeyboardRow(shopCatalogButton(t(lang, "btnReferral"), "acct:referral", "primary")))
	}
	inst.send(chatID, text, tu.InlineKeyboard(rows...))
}

func (inst *botInstance) handleToggleNotifications(chatID, buyerTgID int64, buyerUsername, lang string) {
	customer, err := botSvc.GetOrCreateCustomer(inst.ownerId, buyerTgID, buyerUsername)
	if err != nil {
		inst.send(chatID, t(lang, "error"))
		return
	}
	if err := botSvc.SetNotificationsEnabled(customer.Id, !customer.NotificationsEnabled); err != nil {
		inst.send(chatID, t(lang, "error"))
		return
	}
	inst.sendAccountCard(chatID, buyerTgID, buyerUsername, lang)
}

func (inst *botInstance) sendReferralInfo(chatID, buyerTgID int64, buyerUsername, lang string) {
	customer, err := botSvc.GetOrCreateCustomer(inst.ownerId, buyerTgID, buyerUsername)
	if err != nil {
		inst.send(chatID, t(lang, "error"))
		return
	}
	stats, err := botSvc.ReferralStats(customer.Id)
	if err != nil {
		inst.send(chatID, t(lang, "error"))
		return
	}
	cfg, _ := botSvc.GetConfig(inst.ownerId)
	link := "https://t.me/" + cfg.BotUsername + "?start=ref_" + stats.ReferralCode
	inst.send(chatID, t(lang, "referralInfo", link, stats.ReferredCount, stats.TotalBonusToman))
}
