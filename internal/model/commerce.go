package model

import "time"

type Product struct {
	ID         int64     `gorm:"column:id;primaryKey" json:"id"`
	Title      string    `gorm:"column:title;type:varchar(256);not null" json:"title"`
	Cover      string    `gorm:"column:cover;type:varchar(512)" json:"cover"`
	PriceCent  int64     `gorm:"column:price_cent" json:"priceCent"`
	Stock      int32     `gorm:"column:stock" json:"stock"`
	Status     int32     `gorm:"column:status" json:"status"`
	SellerID   *int64    `gorm:"column:seller_id" json:"sellerId"`
	CreateTime time.Time `gorm:"column:create_time;autoCreateTime" json:"createTime"`
	UpdateTime time.Time `gorm:"column:update_time;autoUpdateTime" json:"updateTime"`
}

func (Product) TableName() string { return "tb_product" }

type Cart struct {
	ID         int64     `gorm:"column:id;primaryKey" json:"id"`
	UserID     int64     `gorm:"column:user_id;uniqueIndex:uk_cart_user_prod" json:"userId"`
	ProductID  int64     `gorm:"column:product_id;uniqueIndex:uk_cart_user_prod" json:"productId"`
	Quantity   int32     `gorm:"column:quantity" json:"quantity"`
	CreateTime time.Time `gorm:"column:create_time;autoCreateTime" json:"createTime"`
	UpdateTime time.Time `gorm:"column:update_time;autoUpdateTime" json:"updateTime"`
}

func (Cart) TableName() string { return "tb_cart" }

type ShopOrder struct {
	ID            int64      `gorm:"column:id;primaryKey" json:"id"`
	UserID        int64      `gorm:"column:user_id" json:"userId"`
	TotalCent     int64      `gorm:"column:total_cent" json:"totalCent"`
	PayCent       *int64     `gorm:"column:pay_cent" json:"payCent"`
	DiscountCent  int64      `gorm:"column:discount_cent" json:"discountCent"`
	UserCouponID  *int64     `gorm:"column:user_coupon_id" json:"userCouponId"`
	Status        int32      `gorm:"column:status" json:"status"`
	PayTime       *time.Time `gorm:"column:pay_time" json:"payTime"`
	SellerSettled int32      `gorm:"column:seller_settled" json:"sellerSettled"`
	FromCart      int32      `gorm:"column:from_cart" json:"fromCart"`
	CreateTime    time.Time  `gorm:"column:create_time;autoCreateTime" json:"createTime"`
	UpdateTime    time.Time  `gorm:"column:update_time;autoUpdateTime" json:"updateTime"`
}

func (ShopOrder) TableName() string { return "tb_order" }

type OrderItem struct {
	ID        int64 `gorm:"column:id;primaryKey" json:"id"`
	OrderID   int64 `gorm:"column:order_id" json:"orderId"`
	ProductID int64 `gorm:"column:product_id" json:"productId"`
	Quantity  int32 `gorm:"column:quantity" json:"quantity"`
	PriceCent int64 `gorm:"column:price_cent" json:"priceCent"`
}

func (OrderItem) TableName() string { return "tb_order_item" }

type UserCoupon struct {
	ID           int64     `gorm:"column:id;primaryKey" json:"id"`
	UserID       int64     `gorm:"column:user_id" json:"userId"`
	ProductID    *int64    `gorm:"column:product_id" json:"productId"`
	Title        string    `gorm:"column:title;type:varchar(128);not null" json:"title"`
	DiscountCent int64     `gorm:"column:discount_cent" json:"discountCent"`
	MinOrderCent int64     `gorm:"column:min_order_cent" json:"minOrderCent"`
	Status       int32     `gorm:"column:status" json:"status"`
	ExpireTime   time.Time `gorm:"column:expire_time" json:"expireTime"`
	CreateTime   time.Time `gorm:"column:create_time;autoCreateTime" json:"createTime"`
}

func (UserCoupon) TableName() string { return "tb_user_coupon" }

type ProductFootprint struct {
	ID        int64     `gorm:"column:id;primaryKey" json:"id"`
	UserID    int64     `gorm:"column:user_id;uniqueIndex:uk_fp_user_product" json:"userId"`
	ProductID int64     `gorm:"column:product_id;uniqueIndex:uk_fp_user_product" json:"productId"`
	ViewedAt  time.Time `gorm:"column:viewed_at;autoCreateTime" json:"viewedAt"`
}

func (ProductFootprint) TableName() string { return "tb_product_footprint" }

type UserWallet struct {
	UserID      int64     `gorm:"column:user_id;primaryKey" json:"userId"`
	BalanceCent int64     `gorm:"column:balance_cent" json:"balanceCent"`
	CreateTime  time.Time `gorm:"column:create_time;autoCreateTime" json:"createTime"`
	UpdateTime  time.Time `gorm:"column:update_time;autoUpdateTime" json:"updateTime"`
}

func (UserWallet) TableName() string { return "tb_user_wallet" }

type WalletRecharge struct {
	ID            int64      `gorm:"column:id;primaryKey" json:"id"`
	UserID        int64      `gorm:"column:user_id" json:"userId"`
	OutTradeNo    string     `gorm:"column:out_trade_no;type:varchar(64);not null;uniqueIndex:uk_wr_out_trade" json:"outTradeNo"`
	AmountCent    int64      `gorm:"column:amount_cent" json:"amountCent"`
	AlipayTradeNo *string    `gorm:"column:alipay_trade_no;type:varchar(64)" json:"alipayTradeNo"`
	Status        int32      `gorm:"column:status" json:"status"`
	CreateTime    time.Time  `gorm:"column:create_time;autoCreateTime" json:"createTime"`
	PayTime       *time.Time `gorm:"column:pay_time" json:"payTime"`
}

func (WalletRecharge) TableName() string { return "tb_wallet_recharge" }
