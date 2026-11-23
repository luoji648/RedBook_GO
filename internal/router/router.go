package router

import (
	"time"

	"github.com/gin-gonic/gin"

	"redbook-go/internal/handler"
	"redbook-go/internal/middleware"
	"redbook-go/internal/service"
)

func New(svc *service.Service) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(
		gin.Recovery(),
		middleware.RequestID(),
		middleware.RequestLogger(),
		middleware.CORS(),
		middleware.ContextTimeout(120*time.Second),
	)

	h := handler.New(svc)
	authMW := middleware.Auth(svc.DB, svc.Tokens)
	optionalAuthMW := middleware.OptionalAuth(svc.DB, svc.Tokens)

	public := r.Group("/", optionalAuthMW)
	{
		public.POST("/user/code", h.SendCode)
		public.POST("/user/login", h.Login)
		public.POST("/user/login/password", h.LoginPassword)
		public.GET("/user/public/:userId", h.PublicProfile)

		public.GET("/note/detail/:id", h.NoteDetail)
		public.GET("/note/recommend", h.NoteRecommend)
		public.GET("/note/related/:id", h.NoteRelated)
		public.GET("/note/comment/tree/:noteId", h.CommentTree)
		public.GET("/note/user/:userId", h.NoteUser)

		public.GET("/group/by-owner/:ownerUserId", h.GroupByOwner)
		public.GET("/like/user/:userId", h.LikeByUser)
		public.GET("/collect/user/:userId", h.CollectByUser)
		public.GET("/follow/public/:userId/followers", h.FollowersOfUser)
		public.GET("/follow/public/:userId/following", h.FollowingOfUser)

		public.GET("/product/list", h.ProductList)
		public.GET("/product/:id", h.ProductGet)
		public.GET("/s/:code", h.ShortLinkRedirect)
	}

	protected := r.Group("/", authMW)
	{
		protected.POST("/user/logout", h.Logout)
		protected.POST("/user/sign", h.Sign)
		protected.GET("/user/sign/count", h.SignCount)
		protected.GET("/user/me", h.Me)
		protected.PUT("/user/profile", h.UpdateProfile)
		protected.GET("/user/info", h.MyInfo)
		protected.PUT("/user/info", h.UpdateInfo)
		protected.PUT("/user/password", h.ChangePassword)

		protected.POST("/note/save", h.SaveNote)
		protected.POST("/note/publish/:id", h.PublishNote)
		protected.DELETE("/note/:id", h.DeleteNote)
		protected.GET("/note/my", h.NoteMy)
		protected.GET("/note/feed", h.NoteFeed)
		protected.POST("/note/comment/:noteId", h.CommentAdd)
		protected.DELETE("/note/comment/:commentId", h.CommentDelete)

		protected.GET("/like/:receiverId", h.LikeView)
		protected.POST("/like/:receiverId", h.Like)
		protected.DELETE("/like/:receiverId", h.Unlike)
		protected.GET("/like/my", h.LikeMy)

		protected.POST("/collect/:noteId", h.Collect)
		protected.DELETE("/collect/:noteId", h.Uncollect)
		protected.GET("/collect/my", h.CollectMy)

		protected.POST("/follow/:userId", h.Follow)
		protected.DELETE("/follow/:userId", h.Unfollow)
		protected.GET("/follow/followers", h.Followers)
		protected.GET("/follow/following", h.Following)

		protected.POST("/share/note/:noteId", h.ShareNote)
		protected.POST("/short-link/generate", h.GenerateShortLink)

		protected.POST("/oss/presign", h.OssPresign)
		protected.POST("/oss/upload", h.OssUpload)

		protected.GET("/product/my", h.ProductMy)
		protected.POST("/product/save", h.ProductSave)
		protected.POST("/product/admin", h.ProductSave)
		protected.PUT("/product/:id/stock", h.ProductUpdateStock)
		protected.POST("/product/footprint/record", h.ProductFootprintRecord)
		protected.GET("/product/footprint/my", h.ProductFootprintMy)

		protected.POST("/cart/add", h.CartAdd)
		protected.GET("/cart/my", h.CartMy)
		protected.PUT("/cart/:cartId", h.CartUpdate)
		protected.DELETE("/cart/:cartId", h.CartRemove)

		protected.POST("/order/create", h.OrderCreate)
		protected.GET("/order/direct-preview", h.OrderDirectPreview)
		protected.POST("/order/create-direct", h.OrderCreateDirect)
		protected.GET("/order/my", h.OrderMy)
		protected.GET("/order/detail/:id", h.OrderDetail)
		protected.POST("/order/pay/:id", h.OrderPay)
		protected.POST("/order/refund/:id", h.OrderRefund)
		protected.DELETE("/order/:id", h.OrderDeleteRecord)
		protected.POST("/order/:id/close", h.OrderClose)
		protected.POST("/order/close/:id", h.OrderClose)

		protected.GET("/wallet/balance", h.WalletBalance)
		protected.POST("/wallet/recharge/alipay", h.WalletRechargeAlipay)

		protected.GET("/coupon/my", h.CouponMy)
		protected.POST("/coupon/claim-follow", h.CouponClaimFollow)

		protected.GET("/chat/unread-total", h.ChatUnreadTotal)
		protected.GET("/chat/threads", h.ChatThreads)
		protected.POST("/chat/:threadId/read", h.ChatMarkThreadRead)
		protected.GET("/chat/peer-thread", h.ChatPeerThread)
		protected.GET("/chat/:threadId/messages", h.ChatMessages)
		protected.POST("/chat/send", h.ChatSend)

		protected.POST("/group/create", h.GroupCreate)
		protected.GET("/group/my", h.GroupMy)
		protected.POST("/group/:groupId/join", h.GroupJoin)
		protected.POST("/group/join-request/:requestId/handle", h.GroupHandleJoinRequest)
		protected.GET("/group/:groupId/meta", h.GroupMeta)
		protected.GET("/group/:groupId/messages", h.GroupMessages)
		protected.POST("/group/:groupId/send", h.GroupSend)
		protected.POST("/group/:groupId/read", h.GroupMarkRead)
		protected.GET("/group/:groupId/detail", h.GroupDetail)
		protected.POST("/group/:groupId/kick", h.GroupKick)
		protected.PUT("/group/:groupId/profile", h.GroupUpdateProfile)

		protected.GET("/notice/interactions", h.NoticeInteractions)
		protected.POST("/notice/read", h.NoticeRead)

		protected.GET("/ai/note/:noteId/stream", h.AiStream)
	}

	r.GET("/swagger/doc.json", handler.SwaggerDoc(r.Routes()))
	r.GET("/doc.html", handler.SwaggerHTML())

	return r
}

