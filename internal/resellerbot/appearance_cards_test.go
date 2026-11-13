package resellerbot

import (
	"strings"
	"testing"

	"github.com/mhsanaei/3x-ui/v3/internal/database/model"
	"github.com/mhsanaei/3x-ui/v3/internal/web/service/panel"
)

func TestInvoiceCardNativeButtonsAndEscapedDetails(t *testing.T) {
	order := &model.BotOrder{
		Id: 42, PublicToken: "AB<&", Kind: model.BotOrderKindPurchase,
		Status: model.BotOrderPendingPayment, FinalAmountToman: 1500000,
		DiscountAmountToman: 50000, TransactionRef: "tx<&",
	}
	text, kb := invoiceCard(order, "Plus <VIP>", langFA)
	for _, want := range []string{
		"🧾 جزئیات فاکتور", "1,500,000 تومان", "50,000 تومان",
		"AB&lt;&amp;", "Plus &lt;VIP&gt;", "tx&lt;&amp;",
	} {
		if !strings.Contains(text, want) {
			t.Errorf("invoice missing %q: %s", want, text)
		}
	}
	if kb == nil || len(kb.InlineKeyboard) != 3 {
		t.Fatalf("pending purchase should show pay, discount and cancel actions: %+v", kb)
	}
	want := []struct{ callback, style string }{
		{"inv:pay:42", "success"},
		{"inv:disc:42", "primary"},
		{"inv:cancel:42", "danger"},
	}
	for i, expected := range want {
		button := kb.InlineKeyboard[i][0]
		if button.CallbackData != expected.callback || button.Style != expected.style {
			t.Errorf("invoice action %d = %+v, want %q/%q", i, button, expected.callback, expected.style)
		}
	}
}

func TestInvoiceReviewHasRefreshWithoutPaymentActions(t *testing.T) {
	order := &model.BotOrder{
		Id: 71, Kind: model.BotOrderKindPurchase,
		Status: model.BotOrderPendingReview, PublicToken: "test-71",
	}
	text, kb := invoiceCard(order, "Starter", langFA)
	if !strings.Contains(text, "بررسی") || kb == nil || len(kb.InlineKeyboard) != 1 ||
		kb.InlineKeyboard[0][0].CallbackData != "inv:refresh:71" {
		t.Fatalf("pending review status should only offer refresh: %q %+v", text, kb)
	}
}

func TestInvoiceCardKeepsNonPayableStatusesAndTopupSafe(t *testing.T) {
	order := &model.BotOrder{
		Id: 17, PublicToken: "T-17", Kind: model.BotOrderKindTopup,
		Status: model.BotOrderPendingPayment, FinalAmountToman: 75000,
	}
	text, kb := invoiceCard(order, "must not be shown", langEN)
	if strings.Contains(text, "must not be shown") || kb == nil || len(kb.InlineKeyboard) != 2 {
		t.Fatalf("wallet topup must omit plan and discount actions: %s, %+v", text, kb)
	}
	if kb.InlineKeyboard[0][0].CallbackData != "inv:pay:17" ||
		kb.InlineKeyboard[1][0].CallbackData != "inv:cancel:17" {
		t.Fatal("topup payment callbacks changed")
	}
	order.Status = model.BotOrderApproved
	text, kb = invoiceCard(order, "", langEN)
	if kb != nil || !strings.Contains(text, "Approved") {
		t.Fatalf("approved invoice must be read-only: %s, %+v", text, kb)
	}
}

func TestSubscriptionCardButtonsLinkAndTraffic(t *testing.T) {
	sub := &model.BotSubscription{Id: 9, ClientEmail: "test<&", AutoRenewEnabled: true}
	stats := &panel.SubscriptionStats{Enable: true, Up: 12, Down: 20, Total: 10}
	text, kb := subscriptionCardWithStats(sub, stats, "https://example.com/sub?a=1&b=2", langFA)
	for _, want := range []string{
		"💎 جزئیات اشتراک", "test&lt;&amp;", "<code>https://example.com/sub?a=1&amp;b=2</code>",
		"🟢 فعال", "باقی", "زمان", "100%", "🟥", "⚠️",
	} {
		if !strings.Contains(text, want) {
			t.Errorf("subscription missing %q: %s", want, text)
		}
	}
	if kb == nil || len(kb.InlineKeyboard) != 4 || len(kb.InlineKeyboard[0]) != 2 || len(kb.InlineKeyboard[1]) != 2 {
		t.Fatalf("subscription action layout changed: %+v", kb)
	}
	got := []struct{ cb, style string }{
		{kb.InlineKeyboard[0][0].CallbackData, kb.InlineKeyboard[0][0].Style},
		{kb.InlineKeyboard[0][1].CallbackData, kb.InlineKeyboard[0][1].Style},
		{kb.InlineKeyboard[1][0].CallbackData, kb.InlineKeyboard[1][0].Style},
		{kb.InlineKeyboard[1][1].CallbackData, kb.InlineKeyboard[1][1].Style},
		{kb.InlineKeyboard[2][0].CallbackData, kb.InlineKeyboard[2][0].Style},
		{kb.InlineKeyboard[3][0].CallbackData, kb.InlineKeyboard[3][0].Style},
	}
	want := []struct{ cb, style string }{
		{"sub:configs:9", "primary"},
		{"sub:renew:9", "success"},
		{"sub:refresh:9", "primary"},
		{"sub:autorenew:9", "success"},
		{"sub:revoke:9", "danger"},
		{"sub:list:0", ""},
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("subscription button %d = %+v, want %+v", i, got[i], want[i])
		}
	}
	sub.AutoRenewEnabled = false
	_, kb = subscriptionCardWithStats(sub, nil, "", langEN)
	if kb.InlineKeyboard[1][1].Style != "primary" {
		t.Fatal("disabled auto-renew should use primary style")
	}
}

func TestWalletCardLocalizedAndGroupedBalance(t *testing.T) {
	fa := accountCardText(langFA, 12345, 1234567, "1405/08/22")
	if !strings.Contains(fa, "<b>💳 کیف پول</b>") ||
		!strings.Contains(fa, "<b>1,234,567 تومان</b>") {
		t.Fatalf("Persian wallet card is not formatted: %s", fa)
	}
	en := accountCardText(langEN, 12345, 1234567, "1405/08/22")
	if !strings.Contains(en, "Wallet") ||
		!strings.Contains(en, "<b>1,234,567 Toman</b>") {
		t.Fatalf("English wallet card is not formatted: %s", en)
	}
}
