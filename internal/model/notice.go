package model

import "time"

type NoticeRead struct {
	ID           int64     `gorm:"column:id;primaryKey" json:"id"`
	UserID       int64     `gorm:"column:user_id;uniqueIndex:uk_notice_read_user_cat" json:"userId"`
	Category     string    `gorm:"column:category;type:varchar(32);not null;uniqueIndex:uk_notice_read_user_cat" json:"category"`
	LastReadTime time.Time `gorm:"column:last_read_time" json:"lastReadTime"`
}

func (NoticeRead) TableName() string { return "tb_notice_read" }

type ShortLink struct {
	ID         int64      `gorm:"column:id;primaryKey" json:"id"`
	ShortURI   *string    `gorm:"column:short_uri;type:varchar(8) CHARACTER SET utf8mb3 COLLATE utf8mb3_bin;comment:短链接;uniqueIndex:uk_short_uri" json:"shortUri"`
	LongURL    *string    `gorm:"column:long_url;type:varchar(1024)" json:"longUrl"`
	ValidDate  *time.Time `gorm:"column:valid_date" json:"validDate"`
	CreateTime time.Time  `gorm:"column:create_time;autoCreateTime" json:"createTime"`
	DelFlag    int32      `gorm:"column:del_flag" json:"delFlag"`
}

func (ShortLink) TableName() string { return "tb_short_link" }
