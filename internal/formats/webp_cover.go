package formats

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"os"
)

// OptimizeCover redimensiona proporcionalmente a imagem de capa (altura máx 1600px) e otimiza a compressão.
func OptimizeCover(srcPath, dstPath string, maxHeight int) error {
	f, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer f.Close()

	img, format, err := image.Decode(f)
	if err != nil {
		return fmt.Errorf("formato de imagem não suportado: %w", err)
	}

	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	if height > maxHeight && maxHeight > 0 {
		newHeight := maxHeight
		newWidth := int(float64(width) * (float64(newHeight) / float64(height)))
		img = resizeBilinear(img, newWidth, newHeight)
	}

	out, err := os.Create(dstPath)
	if err != nil {
		return err
	}
	defer out.Close()

	if format == "png" {
		return png.Encode(out, img)
	}

	// Default para JPEG com qualidade 85
	return jpeg.Encode(out, img, &jpeg.Options{Quality: 85})
}

// OptimizeCoverBytes processa e otimiza bytes de imagem em memória.
func OptimizeCoverBytes(data []byte, maxHeight int) ([]byte, error) {
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return data, err
	}

	bounds := img.Bounds()
	height := bounds.Dy()
	width := bounds.Dx()

	if height > maxHeight && maxHeight > 0 {
		newHeight := maxHeight
		newWidth := int(float64(width) * (float64(newHeight) / float64(height)))
		img = resizeBilinear(img, newWidth, newHeight)
	}

	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: 85}); err != nil {
		return data, err
	}

	return buf.Bytes(), nil
}

// resizeBilinear redimensiona imagens usando interpolação bilinear pura em Go (sem dependências CGO).
func resizeBilinear(src image.Image, targetWidth, targetHeight int) image.Image {
	dst := image.NewRGBA(image.Rect(0, 0, targetWidth, targetHeight))
	bounds := src.Bounds()
	srcW := bounds.Dx()
	srcH := bounds.Dy()

	xRatio := float64(srcW-1) / float64(targetWidth)
	yRatio := float64(srcH-1) / float64(targetHeight)

	for y := 0; y < targetHeight; y++ {
		srcY := int(yRatio * float64(y))
		yDiff := (yRatio * float64(y)) - float64(srcY)

		for x := 0; x < targetWidth; x++ {
			srcX := int(xRatio * float64(x))
			xDiff := (xRatio * float64(x)) - float64(srcX)

			c1 := src.At(bounds.Min.X+srcX, bounds.Min.Y+srcY)
			c2 := src.At(bounds.Min.X+srcX+1, bounds.Min.Y+srcY)
			c3 := src.At(bounds.Min.X+srcX, bounds.Min.Y+srcY+1)
			c4 := src.At(bounds.Min.X+srcX+1, bounds.Min.Y+srcY+1)

			r1, g1, b1, a1 := c1.RGBA()
			r2, g2, b2, a2 := c2.RGBA()
			r3, g3, b3, a3 := c3.RGBA()
			r4, g4, b4, a4 := c4.RGBA()

			interp := func(v1, v2, v3, v4 uint32) uint8 {
				top := float64(v1)*(1.0-xDiff) + float64(v2)*xDiff
				bottom := float64(v3)*(1.0-xDiff) + float64(v4)*xDiff
				res := top*(1.0-yDiff) + bottom*yDiff
				return uint8(res / 257)
			}

			dst.Set(x, y, color.RGBA{
				R: interp(r1, r2, r3, r4),
				G: interp(g1, g2, g3, g4),
				B: interp(b1, b2, b3, b4),
				A: interp(a1, a2, a3, a4),
			})
		}
	}

	return dst
}
