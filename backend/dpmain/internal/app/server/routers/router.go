package routers

import (
	"github.com/gin-gonic/gin"
	"oip/dpmain/internal/app/server/handlers/account"
	"oip/dpmain/internal/app/server/handlers/order"
	"oip/dpmain/internal/app/server/middlewares"
)

// SetupRoutes 配置所有路由，使用 Route Group 分类
func SetupRoutes(
	orderHandler *order.OrderHandler,
	accountHandler *account.AccountHandler,
) *gin.Engine {
	r := gin.Default()

	r.Use(middlewares.CORS())
	r.Use(middlewares.Logger())
	r.Use(middlewares.ErrorHandler())

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"service": "dpmain",
			"message": "Service is running",
		})
	})

	v1 := r.Group("/api/v1")
	{
		accounts := v1.Group("/accounts")
		{
			accounts.POST("", accountHandler.Create)
			accounts.GET("/:id", accountHandler.Get)
		}

		orders := v1.Group("/orders")
		{
			orders.POST("", orderHandler.Create)
			orders.GET("/:id", orderHandler.Get)
		}
	}

	return r
}
