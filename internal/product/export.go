package product

import (
	"encoding/csv"
	"fmt"
	"io"
)

func WriteProductsCSV(w io.Writer, products []Product) error {
	writer := csv.NewWriter(w)
	defer writer.Flush()

	// header
	if err := writer.Write([]string{"SKU", "Name", "Description", "Price", "Quantity"}); err != nil {
		return err
	}
	for _, p := range products {
		row := []string{
			p.SKU,
			p.Name,
			p.Description,
			fmt.Sprintf("%.2f", p.Price),
			fmt.Sprintf("%d", p.Quantity),
		}
		if err := writer.Write(row); err != nil {
			return err
		}
	}
	return nil
}
