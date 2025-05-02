package utils

import (
	"fmt"
	"image/png"
	"os"

	"github.com/boombuler/barcode"
	"github.com/boombuler/barcode/code128"
)

func GenerateBarcodePNG(sku string) (string, error) {
	bar, err := code128.Encode(sku)
	if err != nil {
		return "", err
	}
	scaled, err := barcode.Scale(bar, 300, 100)
	if err != nil {
		return "", err
	}
	path := fmt.Sprintf("barcodes/%s.png", sku)
	f, err := os.Create(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	if err := png.Encode(f, scaled); err != nil {
		return "", err
	}
	return "/" + path, nil
}
