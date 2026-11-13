package resellerbot

import (
	"bytes"
	"context"
	_ "embed"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/mhsanaei/3x-ui/v3/internal/database/model"

	"github.com/mymmrac/telego"
	tu "github.com/mymmrac/telego/telegoutil"
)

const shopCatalogPageSize = 9

//go:embed assets/vpn-shop-locations.png
var shopLocationsBanner []byte

func visibleShopPlans(plans []model.BotPlan) []model.BotPlan {
	visible := make([]model.BotPlan, 0, len(plans))
	for _, plan := range plans {
		if plan.Enabled && !plan.IsTrial {
			visible = append(visible, plan)
		}
	}
	sort.SliceStable(visible, func(i, j int) bool {
		a, b := visible[i], visible[j]
		if a.TrafficGB != b.TrafficGB {
			return a.TrafficGB < b.TrafficGB
		}
		if a.DurationDays != b.DurationDays {
			return a.DurationDays < b.DurationDays
		}
		if a.PriceToman != b.PriceToman {
			return a.PriceToman < b.PriceToman
		}
		return a.Id < b.Id
	})
	return visible
}

func shopCatalogButton(text, callback, style string) telego.InlineKeyboardButton {
	button := tu.InlineKeyboardButton(text).WithCallbackData(callback)
	button.Style = style
	return button
}

func groupedToman(amount int64) string {
	digits := strconv.FormatInt(amount, 10)
	start := 0
	if strings.HasPrefix(digits, "-") {
		start = 1
	}
	for i := len(digits) - 3; i > start; i -= 3 {
		digits = digits[:i] + "," + digits[i:]
	}
	return digits
}

func shopPlanStyle(index int) string {
	switch {
	case index >= 7:
		return "danger"
	case index == 6:
		return "success"
	case index >= 4:
		return "primary"
	default:
		return ""
	}
}

func shopCatalogKeyboard(lang string, plans []model.BotPlan, page int) (*telego.InlineKeyboardMarkup, int, int) {
	if len(plans) == 0 {
		return nil, 0, 0
	}
	pageCount := (len(plans) + shopCatalogPageSize - 1) / shopCatalogPageSize
	if page < 0 {
		page = 0
	}
	if page >= pageCount {
		page = pageCount - 1
	}

	rows := [][]telego.InlineKeyboardButton{
		tu.InlineKeyboardRow(
			shopCatalogButton(t(lang, "catalogTraffic"), "shop:noop", "primary"),
			shopCatalogButton(t(lang, "catalogDuration"), "shop:noop", "primary"),
			shopCatalogButton(t(lang, "catalogPrice"), "shop:noop", "primary"),
		),
	}
	start := page * shopCatalogPageSize
	end := min(start+shopCatalogPageSize, len(plans))
	for index, plan := range plans[start:end] {
		callback := "buy:" + strconv.Itoa(plan.Id)
		style := shopPlanStyle(start + index)
		traffic := fmt.Sprintf("%dGB", plan.TrafficGB)
		if plan.TrafficGB <= 0 {
			traffic = t(lang, "unlimited")
		}
		rows = append(rows, tu.InlineKeyboardRow(
			shopCatalogButton(traffic, callback, style),
			shopCatalogButton(t(lang, "catalogDays", plan.DurationDays), callback, style),
			shopCatalogButton(t(lang, "catalogToman", groupedToman(plan.PriceToman)), callback, style),
		))
	}
	if pageCount > 1 {
		navigation := make([]telego.InlineKeyboardButton, 0, 2)
		if page > 0 {
			navigation = append(navigation, shopCatalogButton(t(lang, "catalogPrevious"), fmt.Sprintf("shop:page:%d", page-1), "primary"))
		}
		if page+1 < pageCount {
			navigation = append(navigation, shopCatalogButton(t(lang, "catalogNext"), fmt.Sprintf("shop:page:%d", page+1), "primary"))
		}
		rows = append(rows, navigation)
	}
	rows = append(rows, tu.InlineKeyboardRow(shopCatalogButton(t(lang, "btnBack"), "shop:back", "")))
	return tu.InlineKeyboard(rows...), page, pageCount
}

func shopCatalogCaption(lang string, page, total int) string {
	caption := "<b>🛡️ VPN SHOP | " + t(lang, "catalogTitle") + "</b>\n"
	caption += "🇩🇪 " + t(lang, "catalogGermany") + "  •  🇳🇱 " + t(lang, "catalogNetherlands") + "  •  🇦🇱 " + t(lang, "catalogAlbania")
	caption += "\n\n" + t(lang, "catalogInstructions")
	if total > 1 {
		caption += "\n" + t(lang, "catalogPage", page+1, total)
	}
	return caption
}

// shopSortPlans offers useful real-plan filters without inventing location
// metadata: a plan may contain inbounds from multiple countries.
func shopSortPlans(plans []model.BotPlan, mode string) []model.BotPlan {
	sorted := append([]model.BotPlan(nil), plans...)
	switch mode {
	case "price":
		sort.SliceStable(sorted, func(i, j int) bool {
			return sorted[i].PriceToman < sorted[j].PriceToman
		})
	case "traffic":
		sort.SliceStable(sorted, func(i, j int) bool {
			a, b := sorted[i].TrafficGB, sorted[j].TrafficGB
			if a == 0 || b == 0 {
				return a == 0 && b != 0
			}
			return a > b
		})
	case "duration":
		sort.SliceStable(sorted, func(i, j int) bool {
			return sorted[i].DurationDays < sorted[j].DurationDays
		})
	}
	return sorted
}

func shopCatalogKeyboardMode(lang string, plans []model.BotPlan, page int, mode string) (*telego.InlineKeyboardMarkup, int, int) {
	plans = shopSortPlans(plans, mode)
	keyboard, currentPage, pageCount := shopCatalogKeyboard(lang, plans, page)
	if keyboard == nil {
		return nil, currentPage, pageCount
	}
	filters := []struct{ name, label string }{
		{"all", "catalogFilterAll"},
		{"price", "catalogFilterPrice"},
		{"traffic", "catalogFilterTraffic"},
		{"duration", "catalogFilterDuration"},
	}
	controls := make([]telego.InlineKeyboardButton, 0, 4)
	for _, f := range filters {
		style := "primary"
		if f.name == mode || (mode == "" && f.name == "all") {
			style = "success"
		}
		controls = append(controls, shopCatalogButton(t(lang, f.label), "shop:filter:"+f.name, style))
	}
	keyboard.InlineKeyboard = append([][]telego.InlineKeyboardButton{
		tu.InlineKeyboardRow(controls[:2]...),
		tu.InlineKeyboardRow(controls[2:]...),
	}, keyboard.InlineKeyboard...)
	if mode != "" && mode != "all" {
		for row := range keyboard.InlineKeyboard {
			for col := range keyboard.InlineKeyboard[row] {
				button := &keyboard.InlineKeyboard[row][col]
				if strings.HasPrefix(button.CallbackData, "shop:page:") {
					button.CallbackData = "shop:page:"+mode+":"+strings.TrimPrefix(button.CallbackData, "shop:page:")
				}
			}
		}
	}
	return keyboard, currentPage, pageCount
}

func (inst *botInstance) sendPlanPage(chatID int64, lang string, page int) {
	inst.sendPlanPageMode(chatID, lang, page, "all")
}

func (inst *botInstance) sendPlanPageMode(chatID int64, lang string, page int, mode string) {
	plans, err := botSvc.ListPlans(inst.ownerId)
	if err != nil {
		inst.send(chatID, t(lang, "error"))
		return
	}
	visible := visibleShopPlans(plans)
	if len(visible) == 0 {
		inst.send(chatID, t(lang, "noPlans"))
		return
	}
	keyboard, currentPage, pages := shopCatalogKeyboardMode(lang, visible, page, mode)
	caption := shopCatalogCaption(lang, currentPage, pages)
	_, err = inst.bot.SendPhoto(context.Background(), &telego.SendPhotoParams{
		ChatID:      tu.ID(chatID),
		Photo:       tu.FileFromReader(bytes.NewReader(shopLocationsBanner), "vpn-shop-locations.png"),
		Caption:     caption,
		ParseMode:   "HTML",
		ReplyMarkup: keyboard,
	})
	if err != nil {
		inst.send(chatID, caption, keyboard)
	}
}
