package service

import (
	"context"
	"time"

	"gorm.io/gorm"

	"redbook-go/internal/model"
)

func (s *Service) StartSchedulers(ctx context.Context, interval time.Duration) {
	if interval <= 0 {
		interval = time.Minute
	}
	ticker := time.NewTicker(interval)
	go func() {
		defer ticker.Stop()
		s.closeExpiredOrders(ctx)
		s.settlePaidOrders(ctx)
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				s.closeExpiredOrders(ctx)
				s.settlePaidOrders(ctx)
			}
		}
	}()
}

func (s *Service) closeExpiredOrders(ctx context.Context) {
	cutoff := time.Now().Add(-15 * time.Minute)
	s.DB.WithContext(ctx).Model(&model.ShopOrder{}).
		Where("status = 0 AND create_time < ?", cutoff).
		Update("status", 2)
}

func (s *Service) settlePaidOrders(ctx context.Context) {
	cutoff := time.Now().Add(-time.Hour)
	var orders []model.ShopOrder
	s.DB.WithContext(ctx).
		Where("status = 1 AND seller_settled = 0 AND pay_time IS NOT NULL AND pay_time < ?", cutoff).
		Limit(100).Find(&orders)
	for _, order := range orders {
		s.settleOrder(ctx, order)
	}
}

func (s *Service) settleOrder(ctx context.Context, order model.ShopOrder) {
	var item model.OrderItem
	if err := s.DB.WithContext(ctx).Where("order_id = ?", order.ID).Order("id ASC").First(&item).Error; err != nil {
		return
	}
	var product model.Product
	if err := s.DB.WithContext(ctx).First(&product, item.ProductID).Error; err != nil || product.SellerID == nil {
		return
	}
	pay := order.PayCent
	if pay == nil || *pay <= 0 {
		return
	}
	err := s.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var wallet model.UserWallet
		if err := tx.First(&wallet, *product.SellerID).Error; err != nil {
			wallet = model.UserWallet{UserID: *product.SellerID}
			if err := tx.Create(&wallet).Error; err != nil {
				return err
			}
		}
		if err := tx.Model(&model.UserWallet{}).Where("user_id = ?", *product.SellerID).
			UpdateColumn("balance_cent", gorm.Expr("balance_cent + ?", *pay)).Error; err != nil {
			return err
		}
		return tx.Model(&model.ShopOrder{}).Where("id = ?", order.ID).Update("seller_settled", 1).Error
	})
	if err != nil {
		_ = err
	}
}
