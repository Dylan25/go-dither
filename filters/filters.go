// imagefry applies pseudo random filter to an image
package filters

import (
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"image/png"
	"math/rand"
	"time"
	"math"
)

func init() {
	image.RegisterFormat("png", "png", png.Decode, png.DecodeConfig)
	image.RegisterFormat("jpeg", "jpeg", jpeg.Decode, jpeg.DecodeConfig)
}

func RandFilter(imageData image.Image, imgCfg image.Config, timesFry int) image.Image {
	// copy old image to a new template

	alteredImage := image.NewRGBA(imageData.Bounds())
	draw.Draw(alteredImage, imageData.Bounds(), imageData, image.Point{}, draw.Over)

	width := imgCfg.Width
	height := imgCfg.Height

	// apply random changes to the image
	for i := 0; i < timesFry; i++ {
		for y := 0; y < height; y++ {
			rand.Seed(time.Now().UTC().UnixNano())
			for x := 0; x < width; x++ {
				r, g, b, a := alteredImage.At(x, y).RGBA()
				newColor := color.RGBA{randColor(uint8(r)), randColor(uint8(g)), randColor(uint8(b)), uint8(a)}
				alteredImage.Set(x, y, newColor)
			}
		}
	}
	return alteredImage
}

func DitherFilter(imageData image.Image , imgCfg image.Config, timesFry int) image.Image {
	// copy old image to a new template

	alteredImage := image.NewRGBA(imageData.Bounds())
	draw.Draw(alteredImage, imageData.Bounds(), imageData, image.Point{}, draw.Over)

	width := imgCfg.Width
	height := imgCfg.Height

	// dither image
	for i := 0; i < timesFry; i++ {
		for y := 0; y < height; y++ {
			for x := 0; x < width; x++ {

				dither(alteredImage, x+1, y,   (float64(7) / float64(16)))
				dither(alteredImage, x-1, y+1, (float64(3) / float64(16)))
				dither(alteredImage, x,   y+1, (float64(5) / float64(16)))
				dither(alteredImage, x+1, y+1, (float64(1) / float64(16)))
			}
		}
	}

	return alteredImage
}

func XorFilter(imageData image.Image, imgCfg image.Config, timesFry int) image.Image {
	// copy old image to a new template

	alteredImage := image.NewRGBA(imageData.Bounds())
	draw.Draw(alteredImage, imageData.Bounds(), imageData, image.Point{}, draw.Over)

	width := imgCfg.Width
	height := imgCfg.Height

	// apply random changes to the image
	for i := 0; i < timesFry; i++ {
		for y := 0; y < height; y++ {
			for x := 0; x < width; x++ {
				xor(alteredImage, x, y)
			}
		}
	}
	return alteredImage
}

func DitherFilterColor(imageData image.Image , imgCfg image.Config, timesFry int) image.Image {
	// copy old image to a new template

	alteredImage := image.NewRGBA(imageData.Bounds())
	draw.Draw(alteredImage, imageData.Bounds(), imageData, image.Point{}, draw.Over)

	width := imgCfg.Width
	height := imgCfg.Height

	// dither image
	for i := 0; i < timesFry; i++ {
		for y := 0; y < height; y++ {
			for x := 0; x < width; x++ {

				dithercolor(alteredImage, x+1, y,   (float64(7) / float64(16)))
				dithercolor(alteredImage, x-1, y+1, (float64(3) / float64(16)))
				dithercolor(alteredImage, x,   y+1, (float64(5) / float64(16)))
				dithercolor(alteredImage, x+1, y+1, (float64(1) / float64(16)))
			}
		}
	}

	return alteredImage
}

func xor(image *image.RGBA, x int, y int) {
	r, g, b, a := image.At(x, y).RGBA()
	newr := r ^ g
	newg := g ^ r
	newb := b ^ newr
	newRGBAColor := color.RGBA{uint8(newr), uint8(newg), uint8(newb), uint8(a)}
	image.Set(x, y, newRGBAColor)
}

func dither(image *image.RGBA, x int, y int, ratio float64) {
	r, g, b, a := image.At(x, y).RGBA()
	origColor := int((r + g + b) / 3)
	newColor := findClosestColor(origColor)
	quant_error := origColor - newColor
	new_val := origColor + int(math.RoundToEven(float64(quant_error) * ratio))
	newRGBAColor := color.RGBA{uint8(new_val), uint8(new_val), uint8(new_val), uint8(a)}
	image.Set(x, y, newRGBAColor)
}

func dithercolor(image *image.RGBA, x int, y int, ratio float64) {
	r, g, b, a := image.At(x, y).RGBA()
	newr := findClosestColor(int(r))
	newg := findClosestColor(int(g))
	newb := findClosestColor(int(b))
	quant_errorr := int(r) - newr
	quant_errorg := int(g) - newg
	quant_errorb := int(b) - newb
	new_valr := int(r) + int(math.RoundToEven(float64(quant_errorr) * ratio))
	new_valg := int(g) + int(math.RoundToEven(float64(quant_errorg) * ratio))
	new_valb := int(b) + int(math.RoundToEven(float64(quant_errorb) * ratio))
	newRGBAColor := color.RGBA{uint8(new_valr), uint8(new_valg), uint8(new_valb), uint8(a)}
	image.Set(x, y, newRGBAColor)
}

func findClosestColor(origColor int) int {
	tmp := math.RoundToEven(float64(origColor) / float64(255))
	if tmp == 1 {
		return 255
	}
	return 0
}

func findClosestColor2(origColor int) int {
	tmp := math.RoundToEven(float64(origColor) / float64(64))
	if tmp == 1 {
		return 64
	}
	return 0
}

func randColor(origColor uint8) uint8 {
	key := rand.Intn(1)
	if key == 0 {
		return origColor + uint8(rand.Intn(10))
	} else {
		return origColor - uint8(rand.Intn(10))
	}
}