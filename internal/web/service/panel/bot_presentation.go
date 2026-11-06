package panel

import (
	"errors"
	"net/url"
	"strings"
	"unicode/utf8"

	"github.com/mhsanaei/3x-ui/v3/internal/database/model"
)

func validateBotPresentation(p *model.BotPresentation) error {
	if utf8.RuneCountInString(p.BrandName) > 80 {
		return errors.New("brand name must be at most 80 characters")
	}
	if utf8.RuneCountInString(p.WelcomeEN) > 2000 || utf8.RuneCountInString(p.Announcement) > 600 {
		return errors.New("welcome must be at most 2000 characters and announcement at most 600")
	}
	if p.DefaultLanguage != "" && p.DefaultLanguage != "auto" && p.DefaultLanguage != "fa" && p.DefaultLanguage != "en" {
		return errors.New("default language must be auto, fa or en")
	}
	if p.MenuLayout != "" && p.MenuLayout != "spacious" && p.MenuLayout != "compact" {
		return errors.New("menu layout must be spacious or compact")
	}
	for _, link := range []struct{ name, value string }{{"support", p.SupportURL}, {"tutorial", p.TutorialURL}} {
		if link.value == "" {
			continue
		}
		u, err := url.Parse(link.value)
		if err != nil || u.Scheme != "https" || u.Hostname() == "" || u.User != nil || len(link.value) > 512 {
			return errors.New(link.name + " URL must use https")
		}
	}
	valid := map[string]bool{"buy": true, "trial": true, "invoices": true, "subs": true, "account": true, "support": true, "tutorial": true}
	seen := map[string]bool{}
	for key, label := range p.Labels {
		lang, action, ok := strings.Cut(key, ":")
		if !ok || (lang != "fa" && lang != "en") || !valid[action] {
			return errors.New("invalid menu label key")
		}
		if strings.ContainsAny(label, "\r\n") || utf8.RuneCountInString(label) > 40 {
			return errors.New("menu labels must be one line and at most 40 characters")
		}
		label = strings.TrimSpace(label)
		if label == "" {
			continue
		}
		id := lang + ":" + label
		if seen[id] {
			return errors.New("menu labels must be unique within each language")
		}
		seen[id] = true
	}
	return nil
}
