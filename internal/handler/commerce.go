package handler

import (
	"github.com/gin-gonic/gin"

	"redbook-go/internal/service"
)

func (h *Handler) ProductMy(c *gin.Context) {
	current, size := pagination(c)
	list, total, err := h.Service.ProductMy(h.userID(c), current, size)
	if err != nil {
		fail(c, 200, err.Error())
		return
	}
	okList(c, list, total)
}

func (h *Handler) ProductList(c *gin.Context) {
	current, size := pagination(c)
	list, total, err := h.Service.ProductList(current, size)
	if err != nil {
		fail(c, 200, err.Error())
		return
	}
	okList(c, list, total)
}

func (h *Handler) ProductGet(c *gin.Context) {
	id, _ := parseInt64(c.Param("id"))
	data, err := h.Service.ProductGet(id)
	if err != nil {
		fail(c, 200, err.Error())
		return
	}
	ok(c, data)
}

func (h *Handler) ProductSave(c *gin.Context) {
	var req service.ProductSaveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, 400, "参数错误")
		return
	}
	data, err := h.Service.ProductSave(h.userID(c), req)
	if err != nil {
		fail(c, 200, err.Error())
		return
	}
	ok(c, data)
}

func (h *Handler) ProductUpdateStock(c *gin.Context) {
	id, _ := parseInt64(c.Param("id"))
	var req service.ProductStockRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, 400, "参数错误")
		return
	}
	if err := h.Service.ProductUpdateStock(h.userID(c), id, req.Stock); err != nil {
		fail(c, 200, err.Error())
		return
	}
	ok(c, nil)
}

func (h *Handler) CartAdd(c *gin.Context) {
	productID, _ := parseInt64(c.Query("productId"))
	quantity, _ := parseInt64(c.DefaultQuery("quantity", "1"))
	if err := h.Service.CartAdd(h.userID(c), productID, int32(quantity)); err != nil {
		fail(c, 200, err.Error())
		return
	}
	ok(c, nil)
}

func (h *Handler) CartMy(c *gin.Context) {
	data, err := h.Service.CartMy(h.userID(c))
	if err != nil {
		fail(c, 200, err.Error())
		return
	}
	ok(c, data)
}

func (h *Handler) CartUpdate(c *gin.Context) {
	id, _ := parseInt64(c.Param("cartId"))
	quantity, _ := parseInt64(c.Query("quantity"))
	if err := h.Service.CartUpdate(h.userID(c), id, quantity); err != nil {
		fail(c, 200, err.Error())
		return
	}
	ok(c, nil)
}

func (h *Handler) CartRemove(c *gin.Context) {
	id, _ := parseInt64(c.Param("cartId"))
	if err := h.Service.CartRemove(h.userID(c), id); err != nil {
		fail(c, 200, err.Error())
		return
	}
	ok(c, nil)
}

func (h *Handler) ProductFootprintRecord(c *gin.Context) {
	productID, _ := parseInt64(c.Query("productId"))
	if err := h.Service.ProductFootprintRecord(h.userID(c), productID); err != nil {
		fail(c, 200, err.Error())
		return
	}
	ok(c, nil)
}

func (h *Handler) ProductFootprintMy(c *gin.Context) {
	current, size := pagination(c)
	list, total, err := h.Service.ProductFootprintMy(h.userID(c), current, size)
	if err != nil {
		fail(c, 200, err.Error())
		return
	}
	okList(c, list, total)
}

func (h *Handler) WalletBalance(c *gin.Context) {
	balance, err := h.Service.WalletBalance(h.userID(c))
	if err != nil {
		fail(c, 200, err.Error())
		return
	}
	ok(c, gin.H{"balanceCent": balance})
}

func (h *Handler) WalletRechargeAlipay(c *gin.Context) {
	var req service.WalletRechargeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, 400, "参数错误")
		return
	}
	data, err := h.Service.WalletRechargeAlipay(h.userID(c), req.AmountCent)
	if err != nil {
		fail(c, 200, err.Error())
		return
	}
	ok(c, data)
}

func (h *Handler) CouponMy(c *gin.Context) {
	current, size := pagination(c)
	list, total, err := h.Service.CouponMy(h.userID(c), current, size)
	if err != nil {
		fail(c, 200, err.Error())
		return
	}
	okList(c, list, total)
}

func (h *Handler) CouponClaimFollow(c *gin.Context) {
	var req service.CouponClaimRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, 400, "参数错误")
		return
	}
	data, err := h.Service.CouponClaimFollow(h.userID(c), req.ProductID)
	if err != nil {
		fail(c, 200, err.Error())
		return
	}
	ok(c, data)
}

func (h *Handler) OrderCreate(c *gin.Context) {
	var req service.OrderCreateFromCartRequest
	_ = c.ShouldBindJSON(&req)
	id, err := h.Service.OrderCreateFromCart(h.userID(c), req)
	if err != nil {
		fail(c, 200, err.Error())
		return
	}
	ok(c, id)
}

func (h *Handler) OrderDirectPreview(c *gin.Context) {
	productID, _ := parseInt64(c.Query("productId"))
	quantity, _ := parseInt64(c.DefaultQuery("quantity", "1"))
	var fallback *int64
	if v := c.Query("fallbackSellerUserId"); v != "" {
		if n, err := parseInt64(v); err == nil {
			fallback = &n
		}
	}
	data, err := h.Service.OrderDirectPreview(h.userID(c), productID, int32(quantity), fallback)
	if err != nil {
		fail(c, 200, err.Error())
		return
	}
	ok(c, data)
}

func (h *Handler) OrderCreateDirect(c *gin.Context) {
	var req service.OrderCreateDirectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, 400, "参数错误")
		return
	}
	id, err := h.Service.OrderCreateDirect(h.userID(c), req)
	if err != nil {
		fail(c, 200, err.Error())
		return
	}
	ok(c, id)
}

func (h *Handler) OrderMy(c *gin.Context) {
	current, size := pagination(c)
	list, total, err := h.Service.OrderMy(h.userID(c), current, size)
	if err != nil {
		fail(c, 200, err.Error())
		return
	}
	okList(c, list, total)
}

func (h *Handler) OrderDetail(c *gin.Context) {
	id, _ := parseInt64(c.Param("id"))
	data, err := h.Service.OrderDetail(h.userID(c), id)
	if err != nil {
		fail(c, 200, err.Error())
		return
	}
	ok(c, data)
}

func (h *Handler) OrderPay(c *gin.Context) {
	id, _ := parseInt64(c.Param("id"))
	var req service.OrderPayRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, 400, "参数错误")
		return
	}
	if err := h.Service.OrderPay(h.userID(c), id, req.PayPassword); err != nil {
		fail(c, 200, err.Error())
		return
	}
	ok(c, nil)
}

func (h *Handler) OrderRefund(c *gin.Context) {
	id, _ := parseInt64(c.Param("id"))
	if err := h.Service.OrderRefund(h.userID(c), id); err != nil {
		fail(c, 200, err.Error())
		return
	}
	ok(c, nil)
}

func (h *Handler) OrderClose(c *gin.Context) {
	id, _ := parseInt64(c.Param("id"))
	if err := h.Service.OrderClose(h.userID(c), id); err != nil {
		fail(c, 200, err.Error())
		return
	}
	ok(c, nil)
}

func (h *Handler) OrderDeleteRecord(c *gin.Context) {
	id, _ := parseInt64(c.Param("id"))
	if err := h.Service.OrderDeleteRecord(h.userID(c), id); err != nil {
		fail(c, 200, err.Error())
		return
	}
	ok(c, nil)
}
