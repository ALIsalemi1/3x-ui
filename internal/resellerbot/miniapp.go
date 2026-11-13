package resellerbot

import (
	"encoding/json"
	"strconv"

	"github.com/mymmrac/telego"
)

type miniAppAction struct {
	Type   string `json:"type"`
	PlanID int    `json:"planId"`
}

// WebAppData is a Telegram service message; its JSON payload is NOT trusted.
// Always use the sender Telegram attached to the message for identity.
func (inst *botInstance) handleMiniAppMessage(message telego.Message, lang string) {
	if message.From == nil || message.Chat.Type != "private" || message.From.ID != message.Chat.ID || len(message.WebAppData.Data) > 1024 {
		return
	}
	var action miniAppAction
	if err := json.Unmarshal([]byte(message.WebAppData.Data), &action); err != nil {
		return
	}
	if isOwnerChat(inst.ownerId, message.Chat.ID) {
		return
	}
	inst.states.clear(message.Chat.ID)
	switch action.Type {
	case "buy":
		if action.PlanID <= 0 {
			return
		}
		plan, err := botSvc.GetPlan(inst.ownerId, action.PlanID)
		if err != nil || plan == nil || !plan.Enabled || plan.IsTrial {
			inst.send(message.Chat.ID, t(lang, "noPlans"))
			return
		}
		inst.handleBuyPlan(message.Chat.ID, strconv.Itoa(action.PlanID), lang)
	case "account", "wallet":
		inst.sendAccountCard(message.Chat.ID, message.From.ID, message.From.Username, lang)
	case "subscriptions":
		inst.sendSubscriptionList(message.Chat.ID, message.From.ID, message.From.Username, lang)
	case "invoices":
		inst.sendInvoiceList(message.Chat.ID, message.From.ID, message.From.Username, lang)
	}
}
