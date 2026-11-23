package model

import "time"

type Note struct {
	ID           int64     `gorm:"column:id;primaryKey" json:"id"`
	UserID       int64     `gorm:"column:user_id" json:"userId"`
	Title        string    `gorm:"column:title;type:varchar(128)" json:"title"`
	Content      string    `gorm:"column:content;type:text" json:"content"`
	Type         int32     `gorm:"column:type" json:"type"`
	Status       int32     `gorm:"column:status" json:"status"`
	Visibility   int32     `gorm:"column:visibility" json:"visibility"`
	LikeCount    int32     `gorm:"column:like_count" json:"likeCount"`
	CollectCount int32     `gorm:"column:collect_count" json:"collectCount"`
	CommentCount int32     `gorm:"column:comment_count" json:"commentCount"`
	CreateTime   time.Time `gorm:"column:create_time;autoCreateTime" json:"createTime"`
	UpdateTime   time.Time `gorm:"column:update_time;autoUpdateTime" json:"updateTime"`
}

func (Note) TableName() string { return "tb_note" }

type NoteMedia struct {
	ID        int64  `gorm:"column:id;primaryKey" json:"id"`
	NoteID    int64  `gorm:"column:note_id" json:"noteId"`
	URL       string `gorm:"column:url;type:varchar(1024);not null" json:"url"`
	SortOrder int32  `gorm:"column:sort_order" json:"sortOrder"`
	MediaType int32  `gorm:"column:media_type" json:"mediaType"`
}

func (NoteMedia) TableName() string { return "tb_note_media" }

type NoteProduct struct {
	ID        int64 `gorm:"column:id;primaryKey" json:"id"`
	NoteID    int64 `gorm:"column:note_id;uniqueIndex:uk_note_product" json:"noteId"`
	ProductID int64 `gorm:"column:product_id;uniqueIndex:uk_note_product" json:"productId"`
	SortOrder int32 `gorm:"column:sort_order" json:"sortOrder"`
}

func (NoteProduct) TableName() string { return "tb_note_product" }

type NoteComment struct {
	ID            int64     `gorm:"column:id;primaryKey" json:"id"`
	NoteID        int64     `gorm:"column:note_id" json:"noteId"`
	UserID        int64     `gorm:"column:user_id" json:"userId"`
	ParentID      *int64    `gorm:"column:parent_id" json:"parentId"`
	ReplyToUserID *int64    `gorm:"column:reply_to_user_id" json:"replyToUserId"`
	Content       string    `gorm:"column:content;type:varchar(1024);not null" json:"content"`
	Status        int32     `gorm:"column:status" json:"status"`
	CreateTime    time.Time `gorm:"column:create_time;autoCreateTime" json:"createTime"`
}

func (NoteComment) TableName() string { return "tb_note_comment" }

type NoteCollect struct {
	ID         int64     `gorm:"column:id;primaryKey" json:"id"`
	UserID     int64     `gorm:"column:user_id;uniqueIndex:uk_collect_user_note" json:"userId"`
	NoteID     int64     `gorm:"column:note_id;uniqueIndex:uk_collect_user_note" json:"noteId"`
	CreateTime time.Time `gorm:"column:create_time;autoCreateTime" json:"createTime"`
}

func (NoteCollect) TableName() string { return "tb_note_collect" }
