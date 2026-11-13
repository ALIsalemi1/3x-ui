package controller

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"
)

func signMiniAppForTest(token string, values url.Values) string {
	var keys []string
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	var lines []string
	for _, key := range keys {
		lines = append(lines, key+"="+values.Get(key))
	}
	secret := hmac.New(sha256.New, []byte("WebAppData"))
	_, _ = secret.Write([]byte(token))
	mac := hmac.New(sha256.New, secret.Sum(nil))
	_, _ = mac.Write([]byte(strings.Join(lines, "\n")))
	values.Set("hash", hex.EncodeToString(mac.Sum(nil)))
	return values.Encode()
}

func TestTelegramMiniAppInitDataRequiresValidSignatureAndFreshness(t *testing.T) {
	now := time.Unix(1800000000, 0)
	data := url.Values{
		"auth_date": {strconv.FormatInt(now.Unix(), 10)},
		"query_id": {"AAF-42"},
		"user": {`{"id":12345,"first_name":"Buyer","username":"vpn_user"}`},
	}
	const token = "123456:BOT_TOKEN_SECRET"
	signed := signMiniAppForTest(token, data)
	buyer, err := verifyMiniAppInitData(signed, token, now)
	if err != nil || buyer.ID != 12345 || buyer.Username != "vpn_user" {
		t.Fatalf("valid signed init data rejected: buyer=%+v err=%v", buyer, err)
	}
	cases := []struct {
		name string
		data string
		key  string
		time time.Time
	}{
		{"wrong bot token", signed, "123456:DIFFERENT", now},
		{"tampered user", strings.Replace(signed, "vpn_user", "hacker", 1), token, now},
		{"stale", signed, token, now.Add(11 * time.Minute)},
		{"too early", signed, token, now.Add(-2 * time.Minute)},
		{"missing hash", strings.Replace(signed, "hash=", "none=", 1), token, now},
		{"duplicate user", signed + "&user=another", token, now},
		{"missing", "", token, now},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := verifyMiniAppInitData(tc.data, tc.key, tc.time); err == nil {
				t.Fatal("untrusted Telegram data was accepted")
			}
		})
	}
}

func TestTelegramMiniAppInitDataRejectsMissingUserAndBadDate(t *testing.T) {
	const token = "123456:BOT_TOKEN_SECRET"
	now := time.Unix(1800000000, 0)
	for _, data := range []url.Values{
		{"auth_date": {strconv.FormatInt(now.Unix(), 10)}},
		{"auth_date": {"not-a-number"}, "user": {`{"id":123}`}},
		{"auth_date": {strconv.FormatInt(now.Unix(), 10)}, "user": {`{"id":0}`}},
		{"auth_date": {strconv.FormatInt(now.Unix(), 10)}, "user": {"invalid"}},
	} {
		if _, err := verifyMiniAppInitData(signMiniAppForTest(token, data), token, now); err == nil {
			t.Fatalf("should reject invalid init data: %+v", data)
		}
	}
}
