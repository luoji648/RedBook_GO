package service

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"redbook-go/internal/model"
)

type ProductSaveRequest struct {
	ID        *int64 `json:"id"`
	Title     string `json:"title"`
	Cover     string `json:"cover"`
	PriceCent int64  `json:"priceCent"`
	Stock     int32  `json:"stock"`
	Status    int32  `json:"status"`
}

type ProductStockRequest struct {
	Stock int32 `json:"stock"`
}

type OrderCreateDirectRequest struct {
	ProductID    int64  `json:"productId"`
	Quantity     int32  `json:"quantity"`
	UserCouponID *int64 `json:"userCouponId"`
}

type OrderCreateFromCartRequest struct {
	UserCouponID *int64 `json:"userCouponId"`
}

type OrderPayRequest struct {
	PayPassword string `json:"payPassword"`
}

type WalletRechargeRequest struct {
	AmountCent int64 `json:"amountCent"`
}

type CouponClaimRequest struct {
	ProductID int64 `json:"productId"`
}

func (s *Service) ProductMy(userID int64, current, size int64) ([]model.Product, int64, error) {
	var products []model.Product
	var total int64
	s.DB.Model(&model.Product{}).Where("seller_id = ?", userID).Count(&total)
	err := s.DB.Where("seller_id = ?", userID).Order("create_time DESC").
		Offset(int((current - 1) * size)).Limit(int(size)).Find(&products).Error
	return products, total, err
}

func (s *Service) ProductList(current, size int64) ([]model.Product, int64, error) {
	var products []model.Product
	var total int64
	s.DB.Model(&model.Product{}).Where("status = 1").Count(&total)
	err := s.DB.Where("status = 1").Order("create_time DESC").
		Offset(int((current - 1) * size)).Limit(int(size)).Find(&products).Error
	return products, total, err
}

func (s *Service) ProductGet(id int64) (*model.Product, error) {
	var product model.Product
	if err := s.DB.First(&product, id).Error; err != nil {
		return nil, errors.New("商品不存在")
	}
	return &product, nil
}

func (s *Service) ProductSave(userID int64, req ProductSaveRequest) (*model.Product, error) {
	var product model.Product
	if req.ID != nil {
		if err := s.DB.Where("id = ? AND seller_id = ?", *req.ID, userID).First(&product).Error; err != nil {
			return nil, errors.New("商品不存在")
		}
	} else {
		product.SellerID = &userID
	}
	product.Title = req.Title
	product.Cover = req.Cover
	product.PriceCent = req.PriceCent
	product.Stock = req.Stock
	product.Status = req.Status
	if product.Status == 0 {
		product.Status = 1
	}
	if err := s.DB.Save(&product).Error; err != nil {
		return nil, err
	}
	return &product, nil
}

func (s *Service) ProductUpdateStock(userID, id int64, stock int32) error {
	res := s.DB.Model(&model.Product{}).Where("id = ? AND seller_id = ? AND status = 1", id, userID).
		Update("stock", stock)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return errors.New("仅已上架商品可调整库存")
	}
	return nil
}

func (s *Service) CartAdd(userID, productID int64, quantity int32) error {
	if quantity <= 0 {
		quantity = 1
	}
	var product model.Product
	if err := s.DB.First(&product, productID).Error; err != nil {
		return errors.New("商品不存在")
	}
	if product.Status != 1 {
		return errors.New("商品已下架")
	}
	var cart model.Cart
	err := s.DB.Where("user_id = ? AND product_id = ?", userID, productID).First(&cart).Error
	if err == nil {
		return s.DB.Model(&cart).Update("quantity", gorm.Expr("quantity + ?", quantity)).Error
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	return s.DB.Create(&model.Cart{UserID: userID, ProductID: productID, Quantity: quantity}).Error
}

func (s *Service) CartMy(userID int64) ([]model.Cart, error) {
	var rows []model.Cart
	if err := s.DB.Where("user_id = ?", userID).Order("create_time DESC").Find(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}

func (s *Service) CartUpdate(userID, cartID, quantity int64) error {
	res := s.DB.Model(&model.Cart{}).Where("id = ? AND user_id = ?", cartID, userID).Update("quantity", quantity)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return errors.New("购物车记录不存在")
	}
	return nil
}

func (s *Service) CartRemove(userID, cartID int64) error {
	res := s.DB.Where("id = ? AND user_id = ?", cartID, userID).Delete(&model.Cart{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return errors.New("购物车记录不存在")
	}
	return nil
}

func (s *Service) ProductFootprintRecord(userID, productID int64) error {
	var product model.Product
	if err := s.DB.First(&product, productID).Error; err != nil {
		return errors.New("商品不存在")
	}
	var fp model.ProductFootprint
	err := s.DB.Where("user_id = ? AND product_id = ?", userID, productID).First(&fp).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return s.DB.Create(&model.ProductFootprint{UserID: userID, ProductID: productID}).Error
	}
	if err != nil {
		return err
	}
	return s.DB.Model(&fp).Update("viewed_at", time.Now()).Error
}

func (s *Service) ProductFootprintMy(userID int64, current, size int64) ([]map[string]any, int64, error) {
	var fps []model.ProductFootprint
	var total int64
	s.DB.Model(&model.ProductFootprint{}).Where("user_id = ?", userID).Count(&total)
	if err := s.DB.Where("user_id = ?", userID).Order("viewed_at DESC").
		Offset(int((current - 1) * size)).Limit(int(size)).Find(&fps).Error; err != nil {
		return nil, 0, err
	}
	list := make([]map[string]any, 0, len(fps))
	for _, fp := range fps {
		var product model.Product
		s.DB.First(&product, fp.ProductID)
		list = append(list, map[string]any{"footprint": fp, "product": product})
	}
	return list, total, nil
}

func (s *Service) WalletBalance(userID int64) (int64, error) {
	var wallet model.UserWallet
	if err := s.DB.First(&wallet, userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			wallet = model.UserWallet{UserID: userID}
			s.DB.Create(&wallet)
			return 0, nil
		}
		return 0, err
	}
	return wallet.BalanceCent, nil
}

func (s *Service) WalletRechargeAlipay(userID, amountCent int64) (map[string]any, error) {
	if amountCent <= 0 {
		return nil, errors.New("充值金额必须大于 0")
	}
	outTradeNo := uuid.NewString()
	now := time.Now()
	recharge := model.WalletRecharge{UserID: userID, OutTradeNo: outTradeNo, AmountCent: amountCent, Status: 1, PayTime: &now}
	err := s.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&recharge).Error; err != nil {
			return err
		}
		var wallet model.UserWallet
		if err := tx.First(&wallet, userID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				wallet = model.UserWallet{UserID: userID}
				if err := tx.Create(&wallet).Error; err != nil {
					return err
				}
			} else {
				return err
			}
		}
		return tx.Model(&model.UserWallet{}).Where("user_id = ?", userID).
			UpdateColumn("balance_cent", gorm.Expr("balance_cent + ?", amountCent)).Error
	})
	if err != nil {
		return nil, err
	}
	return map[string]any{"outTradeNo": outTradeNo, "payFormHtml": nil}, nil
}

func (s *Service) CouponMy(userID int64, current, size int64) ([]map[string]any, int64, error) {
	var coupons []model.UserCoupon
	var total int64
	s.DB.Model(&model.UserCoupon{}).Where("user_id = ? AND status = 0 AND expire_time > ?", userID, time.Now()).Count(&total)
	if err := s.DB.Where("user_id = ? AND status = 0 AND expire_time > ?", userID, time.Now()).
		Order("expire_time ASC").Offset(int((current - 1) * size)).Limit(int(size)).Find(&coupons).Error; err != nil {
		return nil, 0, err
	}
	list := make([]map[string]any, 0, len(coupons))
	for _, c := range coupons {
		m := map[string]any{
			"id":           c.ID,
			"userId":       c.UserID,
			"productId":    c.ProductID,
			"title":        c.Title,
			"discountCent": c.DiscountCent,
			"minOrderCent": c.MinOrderCent,
			"status":       c.Status,
			"expireTime":   c.ExpireTime,
			"createTime":   c.CreateTime,
		}
		if c.ProductID != nil {
			var product model.Product
			s.DB.First(&product, *c.ProductID)
			m["productTitle"] = product.Title
			m["productCover"] = product.Cover
		}
		list = append(list, m)
	}
	return list, total, nil
}

func (s *Service) CouponClaimFollow(userID, productID int64) (map[string]any, error) {
	var product model.Product
	if err := s.DB.First(&product, productID).Error; err != nil {
		return nil, errors.New("商品不存在")
	}
	if product.SellerID == nil {
		return nil, errors.New("该商品暂无卖家")
	}
	var follow model.UserFollow
	if err := s.DB.Where("follower_id = ? AND followee_id = ?", userID, *product.SellerID).First(&follow).Error; err != nil {
		return nil, errors.New("请先关注卖家")
	}
	var existing model.UserCoupon
	if err := s.DB.Where("user_id = ? AND product_id = ?", userID, productID).First(&existing).Error; err == nil {
		return nil, errors.New("该商品优惠券已领取")
	}
	discount := product.PriceCent / 10
	if discount <= 0 {
		discount = 1
	}
	coupon := model.UserCoupon{
		UserID:       userID,
		ProductID:    &productID,
		Title:        fmt.Sprintf("%s 关注券", product.Title),
		DiscountCent: discount,
		MinOrderCent: product.PriceCent,
		Status:       0,
		ExpireTime:   time.Now().Add(30 * 24 * time.Hour),
	}
	if err := s.DB.Create(&coupon).Error; err != nil {
		return nil, err
	}
	return map[string]any{
		"id":           coupon.ID,
		"userId":       coupon.UserID,
		"productId":    coupon.ProductID,
		"title":        coupon.Title,
		"discountCent": coupon.DiscountCent,
		"minOrderCent": coupon.MinOrderCent,
		"status":       coupon.Status,
		"expireTime":   coupon.ExpireTime,
		"createTime":   coupon.CreateTime,
		"productTitle": product.Title,
		"productCover": product.Cover,
	}, nil
}

func (s *Service) OrderCreateFromCart(userID int64, req OrderCreateFromCartRequest) (int64, error) {
	var carts []model.Cart
	if err := s.DB.Where("user_id = ?", userID).Find(&carts).Error; err != nil {
		return 0, err
	}
	if len(carts) == 0 {
		return 0, errors.New("购物车为空")
	}
	orderID, err := s.createOrder(userID, carts, req.UserCouponID, 1)
	if err != nil {
		return 0, err
	}
	return orderID, nil
}

func (s *Service) OrderCreateDirect(userID int64, req OrderCreateDirectRequest) (int64, error) {
	if req.ProductID == 0 {
		return 0, errors.New("商品不能为空")
	}
	if req.Quantity <= 0 {
		req.Quantity = 1
	}
	var product model.Product
	if err := s.DB.First(&product, req.ProductID).Error; err != nil {
		return 0, errors.New("商品不存在")
	}
	if product.Status != 1 {
		return 0, errors.New("商品已下架")
	}
	carts := []model.Cart{{UserID: userID, ProductID: product.ID, Quantity: req.Quantity}}
	return s.createOrder(userID, carts, req.UserCouponID, 0)
}

func (s *Service) createOrder(userID int64, carts []model.Cart, couponID *int64, fromCart int32) (int64, error) {
	var orderID int64
	err := s.DB.Transaction(func(tx *gorm.DB) error {
		var total int64
		type line struct {
			product model.Product
			qty     int32
		}
		lines := make([]line, 0, len(carts))
		for _, c := range carts {
			var product model.Product
			if err := tx.First(&product, c.ProductID).Error; err != nil {
				return err
			}
			if product.Status != 1 {
				return fmt.Errorf("%s 已下架", product.Title)
			}
			if product.Stock < c.Quantity {
				return fmt.Errorf("%s 库存不足", product.Title)
			}
			lines = append(lines, line{product: product, qty: c.Quantity})
			total += product.PriceCent * int64(c.Quantity)
		}

		discount := int64(0)
		if couponID != nil {
			var coupon model.UserCoupon
			if err := tx.Where("id = ? AND user_id = ? AND status = 0", *couponID, userID).First(&coupon).Error; err != nil {
				return errors.New("优惠券不可用")
			}
			if coupon.MinOrderCent > total {
				return errors.New("未达到优惠券使用门槛")
			}
			if coupon.ProductID == nil {
				discount = coupon.DiscountCent
				if discount > total {
					discount = total
				}
			} else {
				var matchedQty int64
				var matchedSub int64
				for _, l := range lines {
					if l.product.ID == *coupon.ProductID {
						matchedQty += int64(l.qty)
						matchedSub += l.product.PriceCent * int64(l.qty)
					}
				}
				discount = coupon.DiscountCent * matchedQty
				if discount > matchedSub {
					discount = matchedSub
				}
				if discount == 0 {
					return errors.New("优惠券不适用于当前商品")
				}
			}
		}

		order := model.ShopOrder{
			UserID:       userID,
			TotalCent:    total,
			DiscountCent: discount,
			UserCouponID: couponID,
			Status:       0,
			FromCart:     fromCart,
		}
		if err := tx.Create(&order).Error; err != nil {
			return err
		}
		orderID = order.ID
		for _, l := range lines {
			item := model.OrderItem{OrderID: order.ID, ProductID: l.product.ID, Quantity: l.qty, PriceCent: l.product.PriceCent}
			if err := tx.Create(&item).Error; err != nil {
				return err
			}
		}
		return nil
	})
	return orderID, err
}

func (s *Service) OrderMy(userID int64, current, size int64) ([]map[string]any, int64, error) {
	var orders []model.ShopOrder
	var total int64
	s.DB.Model(&model.ShopOrder{}).Where("user_id = ?", userID).Count(&total)
	if err := s.DB.Where("user_id = ?", userID).Order("create_time DESC").
		Offset(int((current - 1) * size)).Limit(int(size)).Find(&orders).Error; err != nil {
		return nil, 0, err
	}
	list := make([]map[string]any, 0, len(orders))
	for _, o := range orders {
		vo := s.orderVO(o)
		list = append(list, vo)
	}
	return list, total, nil
}

func (s *Service) OrderDetail(userID, orderID int64) (map[string]any, error) {
	var order model.ShopOrder
	if err := s.DB.Where("id = ? AND user_id = ?", orderID, userID).First(&order).Error; err != nil {
		return nil, errors.New("订单不存在")
	}
	vo := s.orderVO(order)
	vo["items"] = s.orderItems(order.ID)
	return vo, nil
}

func (s *Service) orderVO(order model.ShopOrder) map[string]any {
	items := s.orderItems(order.ID)
	products := make([]map[string]any, 0, len(items))
	titleParts := make([]string, 0, len(items))
	for _, item := range items {
		var product model.Product
		s.DB.First(&product, item.ProductID)
		products = append(products, map[string]any{
			"productId": product.ID,
			"title":     product.Title,
			"cover":     product.Cover,
			"priceCent": item.PriceCent,
			"quantity":  item.Quantity,
		})
		titleParts = append(titleParts, product.Title)
	}
	title := "订单"
	if len(titleParts) > 0 {
		title = titleParts[0]
	}
	return map[string]any{
		"id":           order.ID,
		"userId":       order.UserID,
		"totalCent":    order.TotalCent,
		"payCent":      order.PayCent,
		"discountCent": order.DiscountCent,
		"userCouponId": order.UserCouponID,
		"status":       order.Status,
		"payTime":      order.PayTime,
		"sellerSettled": order.SellerSettled,
		"fromCart":     order.FromCart,
		"createTime":   order.CreateTime,
		"updateTime":   order.UpdateTime,
		"listTitle":    title,
		"products":     products,
	}
}

func (s *Service) orderItems(orderID int64) []model.OrderItem {
	var items []model.OrderItem
	s.DB.Where("order_id = ?", orderID).Find(&items)
	return items
}

func (s *Service) OrderPay(userID, orderID int64, password string) error {
	if password != "123456" {
		return errors.New("支付密码错误")
	}
	return s.DB.Transaction(func(tx *gorm.DB) error {
		var order model.ShopOrder
		if err := tx.Where("id = ? AND user_id = ?", orderID, userID).First(&order).Error; err != nil {
			return errors.New("订单不存在")
		}
		if order.Status != 0 {
			return errors.New("订单状态不允许支付")
		}
		if time.Since(order.CreateTime) > 15*time.Minute {
			return errors.New("订单已超时")
		}
		var items []model.OrderItem
		if err := tx.Where("order_id = ?", order.ID).Find(&items).Error; err != nil {
			return err
		}
		payCent := order.TotalCent - order.DiscountCent
		if payCent < 0 {
			payCent = 0
		}
		var wallet model.UserWallet
		if err := tx.First(&wallet, userID).Error; err != nil {
			return errors.New("钱包不存在")
		}
		if wallet.BalanceCent < payCent {
			return errors.New("钱包余额不足")
		}
		for _, item := range items {
			var product model.Product
			if err := tx.First(&product, item.ProductID).Error; err != nil {
				return err
			}
			if product.Stock < item.Quantity {
				return errors.New("商品库存不足")
			}
			if err := tx.Model(&model.Product{}).Where("id = ? AND stock >= ?", product.ID, item.Quantity).
				UpdateColumn("stock", gorm.Expr("stock - ?", item.Quantity)).Error; err != nil {
				return err
			}
		}
		if err := tx.Model(&model.UserWallet{}).Where("user_id = ? AND balance_cent >= ?", userID, payCent).
			UpdateColumn("balance_cent", gorm.Expr("balance_cent - ?", payCent)).Error; err != nil {
			return err
		}
		now := time.Now()
		updates := map[string]any{"status": 1, "pay_cent": payCent, "pay_time": now}
		if order.UserCouponID != nil {
			if err := tx.Model(&model.UserCoupon{}).Where("id = ?", *order.UserCouponID).Update("status", 1).Error; err != nil {
				return err
			}
		}
		if err := tx.Model(&model.ShopOrder{}).Where("id = ?", order.ID).Updates(updates).Error; err != nil {
			return err
		}
		if order.FromCart == 1 {
			for _, item := range items {
				tx.Where("user_id = ? AND product_id = ?", userID, item.ProductID).Delete(&model.Cart{})
			}
		}
		return nil
	})
}

func (s *Service) OrderRefund(userID, orderID int64) error {
	return s.DB.Transaction(func(tx *gorm.DB) error {
		var order model.ShopOrder
		if err := tx.Where("id = ? AND user_id = ?", orderID, userID).First(&order).Error; err != nil {
			return errors.New("订单不存在")
		}
		if order.Status != 1 || order.PayCent == nil {
			return errors.New("订单不可退款")
		}
		if order.SellerSettled == 1 {
			return errors.New("订单已结算，不可退款")
		}
		if order.PayTime != nil && time.Since(*order.PayTime) > time.Hour {
			return errors.New("已超过退款时间")
		}
		tx.Model(&model.ShopOrder{}).Where("id = ?", order.ID).Update("status", 3)
		tx.Model(&model.UserWallet{}).Where("user_id = ?", userID).
			UpdateColumn("balance_cent", gorm.Expr("balance_cent + ?", *order.PayCent))
		if order.UserCouponID != nil {
			tx.Model(&model.UserCoupon{}).Where("id = ?", *order.UserCouponID).Update("status", 0)
		}
		var items []model.OrderItem
		tx.Where("order_id = ?", order.ID).Find(&items)
		for _, item := range items {
			tx.Model(&model.Product{}).Where("id = ?", item.ProductID).
				UpdateColumn("stock", gorm.Expr("stock + ?", item.Quantity))
		}
		return nil
	})
}

func (s *Service) OrderClose(userID, orderID int64) error {
	res := s.DB.Model(&model.ShopOrder{}).Where("id = ? AND user_id = ? AND status = 0", orderID, userID).Update("status", 2)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return errors.New("仅待支付订单可关闭")
	}
	return nil
}

func (s *Service) OrderDeleteRecord(userID, orderID int64) error {
	res := s.DB.Where("id = ? AND user_id = ? AND status <> 0", orderID, userID).Delete(&model.ShopOrder{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return errors.New("待支付订单不可删除")
	}
	return nil
}

func (s *Service) OrderDirectPreview(userID, productID int64, quantity int32, fallbackSellerID *int64) (map[string]any, error) {
	var product model.Product
	if err := s.DB.First(&product, productID).Error; err != nil {
		return nil, errors.New("商品不存在")
	}
	if quantity <= 0 {
		quantity = 1
	}
	sellerID := product.SellerID
	if sellerID == nil {
		sellerID = fallbackSellerID
	}
	wallet, _ := s.WalletBalance(userID)
	result := map[string]any{
		"product":         product,
		"quantity":        quantity,
		"totalCent":       product.PriceCent * int64(quantity),
		"payCent":         product.PriceCent * int64(quantity),
		"walletBalanceCent": wallet,
	}
	if sellerID != nil {
		var follow model.UserFollow
		followed := s.DB.Where("follower_id = ? AND followee_id = ?", userID, *sellerID).First(&follow).Error == nil
		result["followedSeller"] = followed
		result["sellerUserId"] = *sellerID
		result["sellerId"] = *sellerID
	}
	return result, nil
}
