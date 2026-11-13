package controller

import (
	"crypto/hmac"
	"crypto/sha256"
	"embed"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io/fs"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/mhsanaei/3x-ui/v3/internal/database"
	"github.com/mhsanaei/3x-ui/v3/internal/database/model"
	"github.com/mhsanaei/3x-ui/v3/internal/web/service/panel"
)

// miniAppFiles ships a standalone, responsive Telegram store without an
// external frontend host or JavaScript framework. No generated imagery is used.
//
//go:embed miniapp/*
var miniAppFiles embed.FS

const miniAppAuthMaxAge = 10 * time.Minute

type miniAppBuyer struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
	FirstName string `json:"first_name"`
}

func verifyMiniAppInitData(initData, botToken string, now time.Time) (miniAppBuyer, error) {
	var buyer miniAppBuyer
	if len(initData) == 0 || len(initData) > 8192 || botToken == "" {
		return buyer, errors.New("missing init data")
	}
	values, err := url.ParseQuery(initData)
	if err != nil || len(values.Get("hash")) != 64 {
		return buyer, errors.New("invalid init data")
	}
	for _, vals := range values {
		if len(vals) != 1 {
			return buyer, errors.New("duplicate init data field")
		}
	}
	hash, err := hex.DecodeString(values.Get("hash"))
	if err != nil || len(hash) != sha256.Size {
		return buyer, errors.New("invalid init data hash")
	}
	delete(values, "hash")
	// Telegram's signature is an HMAC over the sorted raw query values.
	keys := make([]string, 0, len(values))
	for k := range values {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	check := make([]string, 0, len(keys))
	for _, k := range keys {
		check = append(check, k+"="+values.Get(k))
	}
	secret := hmac.New(sha256.New, []byte("WebAppData"))
	_, _ = secret.Write([]byte(botToken))
	mac := hmac.New(sha256.New, secret.Sum(nil))
	_, _ = mac.Write([]byte(strings.Join(check, "\n")))
	if !hmac.Equal(hash, mac.Sum(nil)) {
		return buyer, errors.New("invalid Telegram signature")
	}

	authDate, err := strconv.ParseInt(values.Get("auth_date"), 10, 64)
	if err != nil || authDate <= 0 {
		return buyer, errors.New("invalid authorization date")
	}
	issued := time.Unix(authDate, 0)
	if issued.After(now.Add(time.Minute)) || now.Sub(issued) > miniAppAuthMaxAge {
		return buyer, errors.New("expired Telegram login")
	}
	if err := json.Unmarshal([]byte(values.Get("user")), &buyer); err != nil || buyer.ID <= 0 {
		return miniAppBuyer{}, errors.New("invalid Telegram user")
	}
	return buyer, nil
}

type MiniAppController struct {
	service panel.BotService
}

func NewMiniAppController(g *gin.RouterGroup) *MiniAppController {
	a := &MiniAppController{}
	g.GET("/shop/:ownerId/", a.page)
	g.GET("/shop/:ownerId/app.js", a.javascript)
	g.GET("/shop/:ownerId/app.css", a.stylesheet)
	g.POST("/shop/:ownerId/state", a.state)
	return a
}

func (a *MiniAppController) shopConfig(c *gin.Context) (*model.BotConfig, bool) {
	ownerId, err := strconv.Atoi(c.Param("ownerId"))
	if err != nil || ownerId <= 0 {
		c.AbortWithStatus(http.StatusNotFound)
		return nil, false
	}
	cfg := &model.BotConfig{}
	if err := database.GetDB().Where("owner_id = ? AND enabled = ? AND token != ''", ownerId, true).First(cfg).Error; err != nil || cfg.Presentation.MiniAppURL == "" {
		c.AbortWithStatus(http.StatusNotFound)
		return nil, false
	}
	return cfg, true
}

func (a *MiniAppController) asset(c *gin.Context, name, contentType string) {
	if _, ok := a.shopConfig(c); !ok {
		return
	}
	data, err := fs.ReadFile(miniAppFiles, "miniapp/"+name)
	if err != nil {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}
	c.Header("Cache-Control", "no-store")
	c.Header("X-Content-Type-Options", "nosniff")
	c.Data(http.StatusOK, contentType, data)
}

func (a *MiniAppController) page(c *gin.Context) {
	// The global CSP excludes third-party scripts; this narrow public page
	// loads only the official Telegram WebApp SDK, plus same-origin assets.
	c.Header("Content-Security-Policy",
		"default-src 'none'; script-src 'self' https://telegram.org; style-src 'self'; connect-src 'self'; img-src 'self' data:; font-src 'self'; base-uri 'none'; form-action 'none'; frame-ancestors 'none'")
	c.Header("Referrer-Policy", "no-referrer")
	a.asset(c, "index.html", "text/html; charset=utf-8")
}

func (a *MiniAppController) javascript(c *gin.Context) {
	a.asset(c, "app.js", "application/javascript; charset=utf-8")
}

func (a *MiniAppController) stylesheet(c *gin.Context) {
	a.asset(c, "app.css", "text/css; charset=utf-8")
}

func (a *MiniAppController) state(c *gin.Context) {
	c.Header("Cache-Control", "no-store")
	cfg, ok := a.shopConfig(c)
	if !ok {
		return
	}
	buyer, err := verifyMiniAppInitData(c.GetHeader("X-Telegram-Init-Data"), cfg.Token, time.Now())
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Open the shop from the Telegram bot to sign in."})
		return
	}
	customer, err := a.service.GetOrCreateCustomer(cfg.OwnerId, buyer.ID, buyer.Username)
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	plans, err := a.service.ListPlans(cfg.OwnerId)
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	type offer struct {
		ID       int    `json:"id"`
		Name     string `json:"name"`
		Traffic  int64  `json:"trafficGB"`
		Duration int    `json:"durationDays"`
		Price    int64  `json:"priceToman"`
	}
	offers := make([]offer, 0)
	for _, p := range plans {
		if p.Enabled && !p.IsTrial {
			offers = append(offers, offer{p.Id, p.Name, p.TrafficGB, p.DurationDays, p.PriceToman})
		}
	}
	subs, err := a.service.ListSubscriptions(cfg.OwnerId, customer.Id)
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	type subscription struct {
		Name      string `json:"name"`
		Email     string `json:"email"`
		Used      int64  `json:"usedBytes"`
		Total     int64  `json:"totalBytes"`
		Expiry    int64  `json:"expiryTime"`
		Enabled   bool   `json:"enabled"`
	}
	summaries := make([]subscription, 0, min(30, len(subs)))
	for _, sub := range subs {
		if len(summaries) >= 30 {
			break
		}
		name := sub.ClientEmail
		for _, plan := range plans {
			if plan.Id == sub.PlanId {
				name = plan.Name
				break
			}
		}
		item := subscription{Name: name, Email: sub.ClientEmail}
		if stat, statErr := a.service.GetSubscriptionStats(sub.ClientEmail); statErr == nil && stat != nil {
			item.Used = stat.Up + stat.Down
			item.Total = stat.Total
			item.Expiry = stat.ExpiryTime
			item.Enabled = stat.Enable
		}
		summaries = append(summaries, item)
	}
	orders, err := a.service.ListOrdersForBuyer(cfg.OwnerId, buyer.ID)
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	type invoice struct {
		ID     int    `json:"id"`
		Amount int64  `json:"amountToman"`
		Status string `json:"status"`
		Date   int64  `json:"createdAt"`
	}
	invoices := make([]invoice, 0, min(20, len(orders)))
	for i := 0; i < len(orders) && i < 20; i++ {
		o := orders[i]
		invoices = append(invoices, invoice{o.Id, o.FinalAmountToman, o.Status, o.CreatedAt})
	}
	lang := customer.Language
	if lang != "en" {
		lang = "fa"
	}
	c.JSON(http.StatusOK, gin.H{
		"shopName": cfg.Presentation.BrandName,
		"botUsername": cfg.BotUsername,
		"walletEnabled": cfg.WalletEnabled,
		"walletBalance": customer.WalletBalanceToman,
		"language": lang,
		"customerName": buyer.FirstName,
		"plans": offers,
		"subscriptions": summaries,
		"invoices": invoices,
		"supportURL": cfg.Presentation.SupportURL,
		"tutorialURL": cfg.Presentation.TutorialURL,
	})
}
