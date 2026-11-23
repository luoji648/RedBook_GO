package model

import "time"

type UserFollow struct {
	ID         int64     `gorm:"column:id;primaryKey" json:"id"`
	FollowerID int64     `gorm:"column:follower_id;uniqueIndex:uk_follow" json:"followerId"`
	FolloweeID int64     `gorm:"column:followee_id;uniqueIndex:uk_follow" json:"followeeId"`
	CreateTime time.Time `gorm:"column:create_time;autoCreateTime" json:"createTime"`
}

func (UserFollow) TableName() string { return "tb_user_follow" }

type ClickRecords struct {
	ID         int64     `gorm:"column:id;primaryKey" json:"id"`
	ReceiverID int64     `gorm:"column:receiver_id" json:"receiverId"`
	ClickerID  int64     `gorm:"column:clicker_id" json:"clickerId"`
	ClickType  int32     `gorm:"column:click_type" json:"clickType"`
	Created    time.Time `gorm:"column:created;autoCreateTime" json:"created"`
}

func (ClickRecords) TableName() string { return "click_records" }

type ClickSummary struct {
	ID         int64     `gorm:"column:id;primaryKey" json:"id"`
	ReceiverID int64     `gorm:"column:receiver_id;uniqueIndex:uniq_receiver_id" json:"receiverId"`
	LikeCount  int64     `gorm:"column:like_count" json:"likeCount"`
	Created    time.Time `gorm:"column:created;autoCreateTime" json:"created"`
	Modified   time.Time `gorm:"column:modified;autoUpdateTime" json:"modified"`
}

func (ClickSummary) TableName() string { return "click_summary" }
