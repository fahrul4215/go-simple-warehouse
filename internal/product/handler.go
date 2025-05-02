package product

import (
	"errors"
	"go-simple-warehouse/pkg/utils"
	"net/http"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"
)

var validate = validator.New()

// CreateProduct
func CreateProduct(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var in Product
		if err := c.ShouldBindJSON(&in); err != nil {
			c.Error(err)
			return
		}
		if err := validate.Struct(in); err != nil {
			c.Error(err)
			return
		}

		// Generate barcode
		url, err := utils.GenerateBarcodePNG(in.SKU)
		if err != nil {
			c.Error(err)
			return
		}
		in.BarcodeURL = url

		// Save product to database
		if err := db.Create(&in).Error; err != nil {
			c.Error(err)
			return
		}
		c.JSON(http.StatusCreated, in)
	}
}

// UpdateProduct
func UpdateProduct(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		idParam := c.Param("id")
		id, err := strconv.ParseUint(idParam, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid product ID"})
			return
		}

		var existing Product
		if err := db.First(&existing, id).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				c.JSON(http.StatusNotFound, gin.H{"error": "product not found"})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "db error"})
			}
			return
		}

		var input Product
		if err := c.ShouldBindJSON(&input); err != nil {
			c.Error(err)
			return
		}
		if err := validate.Struct(input); err != nil {
			c.Error(err)
			return
		}

		if input.SKU != existing.SKU {
			if existing.BarcodeURL != "" {
				oldPath := "." + existing.BarcodeURL
				_ = os.Remove(oldPath) // ignore error if missing
			}
			newURL, err := utils.GenerateBarcodePNG(input.SKU)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "barcode generation failed"})
				return
			}
			existing.BarcodeURL = newURL
			existing.SKU = input.SKU
		}

		existing.SKU = input.SKU
		existing.Name = input.Name
		existing.Description = input.Description
		existing.Price = input.Price
		existing.Quantity = input.Quantity

		if err := db.Save(&existing).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "could not update product"})
			return
		}

		c.JSON(http.StatusOK, existing)
	}
}

// DeleteProduct
func DeleteProduct(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.ParseUint(c.Param("id"), 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid product ID"})
			return
		}

		var p Product
		if err := db.First(&p, id).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				c.JSON(http.StatusNotFound, gin.H{"error": "product not found"})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "db error"})
			}
			return
		}

		if p.BarcodeURL != "" {
			path := "." + p.BarcodeURL
			if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete barcode file"})
				return
			}
		}

		if err := db.Delete(&Product{}, id).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "could not delete product"})
			return
		}

		c.Status(http.StatusNoContent)
	}
}

// GetProduct
func GetProduct(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var product Product
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid product ID"})
			return
		}
		if err := db.First(&product, id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "product not found"})
			return
		}
		c.JSON(http.StatusOK, product)
	}
}

// ListProducts
func ListProducts(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var (
			products  []Product
			totalRows int64
		)
		q := db.Model(&Product{})

		// Apply filters
		if search := c.Query("search"); search != "" {
			q = q.Where("sku LIKE ? OR name LIKE ?", "%"+search+"%", "%"+search+"%")
		}
		if minQty := c.Query("min_qty"); minQty != "" {
			if qty, err := strconv.Atoi(minQty); err == nil {
				q = q.Where("quantity >= ?", qty)
			} else {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid min_qty value"})
				return
			}
		}
		if maxQty := c.Query("max_qty"); maxQty != "" {
			if qty, err := strconv.Atoi(maxQty); err == nil {
				q = q.Where("quantity <= ?", qty)
			} else {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid max_qty value"})
				return
			}
		}
		if status := c.Query("status"); status != "" {
			if status != "active" && status != "inactive" {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid status value"})
				return
			}
			q = q.Where("status = ?", status)
		}
		if location := c.Query("location"); location != "" {
			q = q.Where("location = ?", location)
		}

		// count total rows
		if err := q.Count(&totalRows).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
			return
		}

		// Apply sorting
		sortBy := c.DefaultQuery("sort_by", "id")
		order := c.DefaultQuery("order", "asc")
		if order != "asc" && order != "desc" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid order value"})
			return
		}
		q = q.Order(sortBy + " " + order)

		// Apply pagination
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		size, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
		if page < 1 {
			page = 1
		}
		if size < 1 {
			size = 20
		}
		offset := (page - 1) * size
		q = q.Offset(offset).Limit(size)

		// Fetch data
		if err := q.Find(&products).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
			return
		}

		// total pages
		totalPages := int((totalRows + int64(size) - 1) / int64(size))
		if totalPages == 0 {
			totalPages = 1
		}

		c.JSON(http.StatusOK, gin.H{
			"data": products,
			"meta": gin.H{
				"total":       totalRows,
				"page":        page,
				"page_size":   size,
				"total_pages": totalPages,
			},
		})
	}
}

// ExportCSV handles exporting products to a CSV file.
func ExportCSV(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var products []Product
		q := db.Model(&Product{})

		// Apply filters
		if search := c.Query("search"); search != "" {
			q = q.Where("sku LIKE ? OR name LIKE ?", "%"+search+"%", "%"+search+"%")
		}
		if minQty := c.Query("min_qty"); minQty != "" {
			if qty, err := strconv.Atoi(minQty); err == nil {
				q = q.Where("quantity >= ?", qty)
			} else {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid min_qty value"})
				return
			}
		}

		// Apply sorting
		sortBy := c.DefaultQuery("sort_by", "id")
		order := c.DefaultQuery("order", "asc")
		if order != "asc" && order != "desc" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid order value"})
			return
		}
		q = q.Order(sortBy + " " + order)

		// Fetch data
		if err := q.Find(&products).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
			return
		}

		// Write CSV response
		c.Header("Content-Type", "text/csv")
		c.Header("Content-Disposition", "attachment;filename=products.csv")
		if err := WriteProductsCSV(c.Writer, products); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to write CSV"})
		}
	}
}
