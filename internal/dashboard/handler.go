package dashboard

import (
	"go-simple-warehouse/internal/product"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func DashboardMetrics(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		type metrics struct {
			TotalProducts int64 `json:"total_products"`
			TotalStock    int64 `json:"total_stock"`
			LowStockCount int64 `json:"low_stock_count"`
		}
		var m metrics
		db.Model(&product.Product{}).Count(&m.TotalProducts)
		db.Model(&product.Product{}).Select("SUM(quantity)").Scan(&m.TotalStock)
		threshold := 5
		db.Model(&product.Product{}).Where("quantity <= ?", threshold).Count(&m.LowStockCount)
		c.JSON(http.StatusOK, m)
	}
}
