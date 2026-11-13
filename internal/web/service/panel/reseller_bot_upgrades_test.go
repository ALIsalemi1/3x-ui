package panel

import (
	"testing"
	"time"

	"github.com/mhsanaei/3x-ui/v3/internal/database"
	"github.com/mhsanaei/3x-ui/v3/internal/database/model"
)

func TestBuyerRepeatClicksReuseUnpaidInvoices(t *testing.T) {
	setupBotTestDB(t)
	db := database.GetDB()
	plan := &model.BotPlan{
		OwnerId:      710,
		Name:         "Repeated click",
		PriceToman:   100000,
		DurationDays: 30,
		TrafficGB:    20,
		Enabled:      true,
	}
	if err := db.Create(plan).Error; err != nil {
		t.Fatalf("create plan: %v", err)
	}
	svc := &BotService{}
	first, err := svc.CreatePurchaseOrder(710, plan.Id, 5101, "buyer", 0, "", "")
	if err != nil {
		t.Fatal(err)
	}
	second, err := svc.CreatePurchaseOrder(710, plan.Id, 5101, "buyer", 0, "", "")
	if err != nil || first.Id != second.Id {
		t.Fatalf("duplicate purchase generated a second invoice: %+v / %+v (%v)", first, second, err)
	}
	differentBuyer, err := svc.CreatePurchaseOrder(710, plan.Id, 5102, "buyer2", 0, "", "")
	if err != nil || differentBuyer.Id == first.Id {
		t.Fatal("different buyers must never reuse an invoice")
	}
	if err := svc.CancelOrder(710, first.Id); err != nil {
		t.Fatal(err)
	}
	next, err := svc.CreatePurchaseOrder(710, plan.Id, 5101, "buyer", 0, "", "")
	if err != nil || next.Id == first.Id {
		t.Fatal("cancelled invoices must not be reused")
	}

	topup, err := svc.CreateTopupOrder(710, 5101, "buyer", 50000, 0)
	if err != nil {
		t.Fatal(err)
	}
	topupAgain, err := svc.CreateTopupOrder(710, 5101, "buyer", 50000, 0)
	if err != nil || topup.Id != topupAgain.Id {
		t.Fatalf("duplicate wallet top-up was not reused: %v", err)
	}
}

func TestSalesDashboardPeriodRevenueAndReferralHistory(t *testing.T) {
	setupBotTestDB(t)
	db := database.GetDB()
	parent := &model.BotCustomer{OwnerId: 800, BuyerTgId: 77001, ReferralCode: "parent001", CreatedAt: time.Now().UnixMilli()}
	if err := db.Create(parent).Error; err != nil {
		t.Fatal(err)
	}
	for i := range 3 {
		buyer := &model.BotCustomer{
			OwnerId: 800, BuyerTgId: 77002 + int64(i),
			ReferralCode:         "child00" + string(rune('a'+i)),
			ReferredByCustomerId: parent.Id,
		}
		if err := db.Create(buyer).Error; err != nil {
			t.Fatal(err)
		}
	}
	orders := []model.BotOrder{
		{
			OwnerId: 800, BuyerTgId: 77001, PublicToken: "sales-purchase-1",
			Kind: model.BotOrderKindPurchase, Status: model.BotOrderApproved,
			FinalAmountToman: 120000, PaidAt: time.Now().UnixMilli(),
		},
		{
			OwnerId: 800, BuyerTgId: 77001, PublicToken: "sales-topup-1",
			Kind: model.BotOrderKindTopup, Status: model.BotOrderApproved,
			FinalAmountToman: 50000, PaidAt: time.Now().UnixMilli(),
		},
	}
	if err := db.Create(&orders).Error; err != nil {
		t.Fatal(err)
	}
	svc := &BotService{}
	stats, err := svc.GetOwnerStats(800)
	if err != nil {
		t.Fatal(err)
	}
	if stats.RevenueTodayToman != 120000 || stats.RevenueMonthToman != 120000 ||
		stats.RevenueToman != 120000 || stats.NewCustomersToday < 1 {
		t.Fatalf("shop revenue periods double counted a top-up: %+v", stats)
	}
	history, err := svc.RecentReferrals(parent.Id)
	if err != nil || len(history) != 3 {
		t.Fatalf("referral history missing: %v, %+v", err, history)
	}
}
