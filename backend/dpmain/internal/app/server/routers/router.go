package routers

import (
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	_ "oip/dpmain/docs"
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

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	r.GET("/health", HealthCheck)

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

// HealthCheck godoc
// @Summary      健康检查
// @Description  检查服务运行状态
// @Tags         system
// @Produce      json
// @Success      200 {object} map[string]string
// @Router       /health [get]
func HealthCheck(c *gin.Context) {
	c.JSON(200, gin.H{
		"status":  "ok",
		"service": "dpmain",
		"message": "Service is running",
	})
}
