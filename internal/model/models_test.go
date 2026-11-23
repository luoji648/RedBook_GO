package model

import (
	"sync"
	"testing"

	"gorm.io/gorm/schema"
)

func TestTimeColumnsAutoFill(t *testing.T) {
	models := []any{
		&User{}, &UserInfo{}, &Note{}, &NoteComment{}, &NoteCollect{}, &UserFollow{},
		&Product{}, &Cart{}, &ShopOrder{}, &UserCoupon{}, &ProductFootprint{},
		&UserWallet{}, &WalletRecharge{}, &ChatMessage{}, &ChatGroup{},
		&ChatGroupMember{}, &ChatGroupJoinRequest{}, &GroupMessage{}, &ShortLink{},
		&ClickRecords{}, &ClickSummary{},
	}

	createColumns := map[string]bool{
		"create_time": true,
		"created":     true,
		"viewed_at":   true,
	}
	updateColumns := map[string]bool{
		"update_time": true,
		"modified":    true,
	}

	cache := &sync.Map{}
	for _, value := range models {
		parsed, err := schema.Parse(value, cache, schema.NamingStrategy{})
		if err != nil {
			t.Fatalf("parse %T: %v", value, err)
		}
		for _, field := range parsed.Fields {
			if createColumns[field.DBName] && field.AutoCreateTime == 0 {
				t.Errorf("%s.%s must use autoCreateTime", parsed.Table, field.DBName)
			}
			if updateColumns[field.DBName] && field.AutoUpdateTime == 0 {
				t.Errorf("%s.%s must use autoUpdateTime", parsed.Table, field.DBName)
			}
		}
	}
}
