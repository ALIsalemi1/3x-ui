package resellerbot

import (
	"bytes"
	"context"
	_ "embed"

	"github.com/mymmrac/telego"
	tu "github.com/mymmrac/telego/telegoutil"
)

//go:embed assets/vpn-shop-banner.png
var vpnShopBanner []byte

func (inst *botInstance) sendShopBanner(chatID int64, lang string) {
	caption := "🛡️ <b>VPN SHOP</b>\n⚡ Fast • Secure • Private"
	if lang == langFA {
		caption = "🛡️ <b>VPN SHOP</b>\n⚡ سریع • امن • پایدار"
	}
	_, _ = inst.bot.SendPhoto(context.Background(), &telego.SendPhotoParams{
		ChatID:    tu.ID(chatID),
		Photo:     tu.FileFromReader(bytes.NewReader(vpnShopBanner), "vpn-shop-banner.png"),
		Caption:   caption,
		ParseMode: "HTML",
	})
}
