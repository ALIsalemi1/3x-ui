package database

import (
	"database/sql"
	"path/filepath"
	"testing"

	"github.com/mhsanaei/3x-ui/v3/internal/database/model"
)

func TestLegacyBotConfigMigrationKeepsTokenAndWelcome(t *testing.T) {
	path := filepath.Join(t.TempDir(), "x-ui.db")
	legacy, err := sql.Open("sqlite3", path)
	if err != nil {
		t.Fatal(err)
	}
	_, err = legacy.Exec(`CREATE TABLE bot_configs (id INTEGER PRIMARY KEY,owner_id INTEGER,token TEXT,welcome_message TEXT,enabled NUMERIC); INSERT INTO bot_configs VALUES (1,77,'existing-token','سلام',1)`)
	if err != nil {
		t.Fatal(err)
	}
	if err := legacy.Close(); err != nil {
		t.Fatal(err)
	}
	if err := InitDB(path); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = CloseDB() })
	var cfg model.BotConfig
	if err := GetDB().First(&cfg, 1).Error; err != nil {
		t.Fatal(err)
	}
	if cfg.OwnerId != 77 || cfg.Token != "existing-token" || cfg.WelcomeMessage != "سلام" || !cfg.Enabled {
		t.Fatal("migration changed existing shop configuration")
	}
	var raw string
	if err := GetDB().Table("bot_configs").Select("presentation").Where("id = ?", 1).Scan(&raw).Error; err != nil {
		t.Fatal(err)
	}
	if raw != "{}" {
		t.Fatalf("legacy presentation = %q, want an empty JSON object", raw)
	}
}
