package resellerbot

import (
	"context"
	"html"
	"strconv"
	"time"

	"github.com/mhsanaei/3x-ui/v3/internal/database"
	"github.com/mhsanaei/3x-ui/v3/internal/database/model"
	"github.com/mhsanaei/3x-ui/v3/internal/logger"
	"github.com/mhsanaei/3x-ui/v3/internal/web/service/panel"

	"github.com/mymmrac/telego"
	tu "github.com/mymmrac/telego/telegoutil"
)

const (
	subscriptionAlertInterval = 15 * time.Minute
	quotaAlertCooldown        = 24 * time.Hour
)

// expiryAlertStage picks just one alert per expiry threshold, even if the
// worker was offline when the earlier (72h or 24h) threshold was crossed.
func expiryAlertStage(expiryMs, nowMs int64) int {
	if expiryMs <= nowMs || expiryMs <= 0 {
		return 0
	}
	remaining := time.Duration(expiryMs-nowMs) * time.Millisecond
	switch {
	case remaining <= 6*time.Hour:
		return 3
	case remaining <= 24*time.Hour:
		return 2
	case remaining <= 72*time.Hour:
		return 1
	default:
		return 0
	}
}

func expiryAlertDue(sub model.BotSubscription, expiryMs, nowMs int64) (int, bool) {
	stage := expiryAlertStage(expiryMs, nowMs)
	if stage == 0 {
		return 0, false
	}
	return stage, sub.LastExpiryAlertFor != expiryMs || sub.LastExpiryAlertStage < stage
}

func quotaAlertDue(sub model.BotSubscription, stats *panel.SubscriptionStats, nowMs int64) bool {
	if stats == nil || !stats.Enable || stats.Total <= 0 || stats.Up+stats.Down < 0 {
		return false
	}
	used := stats.Up + stats.Down
	// Avoid overflow on giant traffic counters.
	percentage := float64(used) / float64(stats.Total)
	return percentage >= 0.9 && nowMs-sub.LastLowQuotaAlertAt >= quotaAlertCooldown.Milliseconds()
}

func (inst *botInstance) subscriptionAlertLoop(ctx context.Context) {
	defer inst.wg.Done()
	ticker := time.NewTicker(subscriptionAlertInterval)
	defer ticker.Stop()

	// Existing customers also need reminders after a bot/server restart.
	inst.deliverSubscriptionAlerts(ctx)
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			inst.deliverSubscriptionAlerts(ctx)
		}
	}
}

func (inst *botInstance) deliverSubscriptionAlerts(ctx context.Context) {
	var subs []model.BotSubscription
	err := database.GetDB().Table("bot_subscriptions").
		Select("bot_subscriptions.*").
		Joins("JOIN bot_customers ON bot_customers.id = bot_subscriptions.customer_id").
		Where("bot_subscriptions.owner_id = ? AND bot_customers.notifications_enabled = ?", inst.ownerId, true).
		Order("bot_subscriptions.id DESC").
		Limit(1000).Find(&subs).Error
	if err != nil {
		logger.Warningf("resellerbot: alert candidate query failed for owner %d: %v", inst.ownerId, err)
		return
	}

	for _, sub := range subs {
		if ctx.Err() != nil {
			return
		}
		var customer model.BotCustomer
		err := database.GetDB().
			Where("id = ? AND owner_id = ?", sub.CustomerId, inst.ownerId).
			First(&customer).Error
		if err != nil || customer.BuyerTgId == 0 {
			continue
		}
		stats, err := botSvc.GetSubscriptionStats(sub.ClientEmail)
		if err != nil || stats == nil || !stats.Enable {
			continue
		}
		lang := customer.Language
		if lang != langEN {
			lang = langFA
		}
		nowMs := time.Now().UnixMilli()
		id := strconv.Itoa(sub.Id)
		keyboard := tu.InlineKeyboard(tu.InlineKeyboardRow(
			shopCatalogButton(t(lang, "btnRenew"), "sub:renew:"+id, "success"),
			shopCatalogButton(t(lang, "btnSubDetails"), "sub:view:"+id, "primary"),
		))
		escapedEmail := html.EscapeString(sub.ClientEmail)
		if stage, due := expiryAlertDue(sub, stats.ExpiryTime, nowMs); due {
			text := "<b>" + t(lang, "expiryAlertHeading") + "</b>\n" +
				t(lang, "subUsername", escapedEmail) + "\n" +
				t(lang, "subRemaining", timeLeftLabel(stats.ExpiryTime, lang)) + "\n" +
				t(lang, "expiryAlertCallToAction")
			if inst.sendSubscriptionAlert(ctx, customer.BuyerTgId, text, keyboard) {
				err := database.GetDB().Model(&model.BotSubscription{}).
					Where("id = ? AND owner_id = ?", sub.Id, inst.ownerId).
					Updates(map[string]any{
						"last_expiry_alert_for":   stats.ExpiryTime,
						"last_expiry_alert_stage": stage,
					}).Error
				if err != nil {
					logger.Warningf("resellerbot: could not save expiry reminder checkpoint: %v", err)
				}
			}
		}
		if quotaAlertDue(sub, stats, nowMs) {
			bar, percentage := quotaUsageBar(stats.Up+stats.Down, stats.Total)
			text := "<b>" + t(lang, "quotaAlertHeading") + "</b>\n" +
				t(lang, "subUsername", escapedEmail) + "\n" +
				bar + " " + strconv.Itoa(percentage) + "%\n" +
				t(lang, "quotaAlertCallToAction")
			if inst.sendSubscriptionAlert(ctx, customer.BuyerTgId, text, keyboard) {
				if err := database.GetDB().Model(&model.BotSubscription{}).
					Where("id = ? AND owner_id = ?", sub.Id, inst.ownerId).
					Update("last_low_quota_alert_at", nowMs).Error; err != nil {
					logger.Warningf("resellerbot: could not save quota reminder checkpoint: %v", err)
				}
			}
		}
	}
}

func (inst *botInstance) sendSubscriptionAlert(
	parent context.Context, chatID int64, text string, keyboard *telego.InlineKeyboardMarkup,
) bool {
	ctx, cancel := context.WithTimeout(parent, 12*time.Second)
	defer cancel()
	_, err := inst.bot.SendMessage(ctx, &telego.SendMessageParams{
		ChatID: tu.ID(chatID), Text: text, ParseMode: "HTML", ReplyMarkup: keyboard,
	})
	if err != nil {
		logger.Warningf("resellerbot: failed to deliver a subscription reminder: %v", err)
		return false
	}
	return true
}
