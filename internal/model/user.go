package model

import "time"

type User struct {
	ID           int64     `gorm:"column:id;primaryKey" json:"id"`
	Phone        string    `gorm:"column:phone;type:varchar(20);not null;uniqueIndex:phone" json:"-"`
	Password     string    `gorm:"column:password;type:varchar(128)" json:"-"`
	NickName     string    `gorm:"column:nick_name;type:varchar(64);not null" json:"nickName"`
	Icon         string    `gorm:"column:icon;type:varchar(512);not null" json:"icon"`
	TokenVersion int32     `gorm:"column:token_version" json:"-"`
	CreateTime   time.Time `gorm:"column:create_time;autoCreateTime" json:"createTime"`
	UpdateTime   time.Time `gorm:"column:update_time;autoUpdateTime" json:"updateTime"`
}

func (User) TableName() string { return "tb_user" }

type UserInfo struct {
	UserID        int64      `gorm:"column:user_id;primaryKey" json:"userId"`
	City          string     `gorm:"column:city;type:varchar(64)" json:"city"`
	Introduce     string     `gorm:"column:introduce;type:varchar(256)" json:"introduce"`
	Fans          int32      `gorm:"column:fans" json:"fans"`
	Followee      int32      `gorm:"column:followee" json:"followee"`
	Gender        bool       `gorm:"column:gender" json:"gender"`
	Birthday      *time.Time `gorm:"column:birthday" json:"birthday"`
	Credits       int32      `gorm:"column:credits" json:"credits"`
	Level         int32      `gorm:"column:level" json:"level"`
	CollectPublic bool       `gorm:"column:collect_public" json:"collectPublic"`
	LikePublic    bool       `gorm:"column:like_public" json:"likePublic"`
	CreateTime    time.Time  `gorm:"column:create_time;autoCreateTime" json:"createTime"`
	UpdateTime    time.Time  `gorm:"column:update_time;autoUpdateTime" json:"updateTime"`
}

func (UserInfo) TableName() string { return "tb_user_info" }
