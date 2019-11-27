// imagefry applies pseudo random filter to an image
package main

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"
	"math"
)

func init() {
	image.RegisterFormat("png", "png", png.Decode, png.DecodeConfig)
	image.RegisterFormat("jpeg", "jpeg", jpeg.Decode, jpeg.DecodeConfig)
}

func openDecodeFilterStatic(ImageFile *os.File, timesFry int) (image.Image, string) {
	imageData, imageType, err := image.Decode(ImageFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "perlin: %v\n", err)
	}

	ImageFile.Seek(0, 0)

	imgCfg, _, err := image.DecodeConfig(ImageFile)

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	ImageFile.Seek(0, 0)

	newImg := randFilter(imageData, imgCfg, timesFry)
	return newImg, imageType
}

func main() {
	rand.Seed(time.Now().UTC().UnixNano())
	ImageFile, timesFry := inputParseAndOpen()
	defer ImageFile.Close()

	if strings.HasSuffix(os.Args[1], ".gif") {
		_, newGif := SplitAnimatedGIF(ImageFile, timesFry)
		outputFile, err := os.Create("fryd" + os.Args[1])
		if err != nil {
			fmt.Fprintf(os.Stderr, "perlin output error: %s\n", err)
		}
		gif.EncodeAll(outputFile, newGif)
		outputFile.Close()
	} else {
		newImg, imageType := openDecodeFilterStatic(ImageFile, timesFry)

		outputFile, err := os.Create("fryd" + os.Args[1])
		if err != nil {
			fmt.Fprintf(os.Stderr, "perlin output error: %s\n", err)
		}

		if imageType == "png" {
			png.Encode(outputFile, newImg)
		} else if imageType == "jpeg" {
			jpeg.Encode(outputFile, newImg, nil)
		} else {
			fmt.Println("ERROR: unrecognized file format")
		}
		outputFile.Close()
	}

	fmt.Printf("output written to %s\n", "fryd"+os.Args[1])
}

func inputParseAndOpen() (*os.File, int) {
	if len(os.Args) <= 1 {
		fmt.Fprint(os.Stderr, "ERROR: please provide a filename\n")
		fmt.Println("USAGE: 'imagefry.exe image.jpg/png #times_fryd'")
		os.Exit(1)
	}
	if strings.HasSuffix(os.Args[1], ".png") || strings.HasSuffix(os.Args[1], ".jpg") || strings.HasSuffix(os.Args[1], ".gif") {
		imageFile, err := os.Open(os.Args[1])
		if err != nil {
			fmt.Fprintf(os.Stderr, "could not open file, %v\n", err)
			os.Exit(1)
		}

		if len(os.Args) == 3 {
			numfry := os.Args[2]
			intnumfry, err := strconv.Atoi(numfry)
			if err != nil {
				fmt.Fprintf(os.Stderr, "ERROR: please enter a number of times to fry, %s\n", err)
				os.Exit(1)
			}
			return imageFile, intnumfry
		} else if len(os.Args) > 3 {
			fmt.Fprint(os.Stderr, "ERROR: too many arguments")
			os.Exit(1)
		} else {
			return imageFile, 1
		}

	} else {
		fmt.Fprint(os.Stderr, "ERROR: please provide a filename\n")
		fmt.Println("USAGE: 'imagefry.exe image.jpg/png #times_fryd'")
		os.Exit(1)
	}
	fmt.Fprintf(os.Stderr, "Argument parse error, try again")
	os.Exit(1)
	return nil, 0
}

func randFilter(imageData image.Image, imgCfg image.Config, timesFry int) image.Image {
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

func ditherFilter(imageData image.Image , imgCfg image.Config, timesFry int) image.Image {
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

func dither(image *image.RGBA, x int, y int, ratio float64) {
	r, g, b, a := image.At(x, y).RGBA()
	origColor := int((r + g + b) / 3)
	newColor := findClosestColor(origColor)
	quant_error := origColor - newColor
	new_val := origColor + int(math.RoundToEven(float64(quant_error) * ratio))
	newRGBAColor := color.RGBA{uint8(new_val), uint8(new_val), uint8(new_val), uint8(a)}
	image.Set(x, y, newRGBAColor)
}

func findClosestColor(origColor int) int {
	tmp := math.RoundToEven(float64(origColor) / float64(255))
	if tmp == 1 {
		return 255
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

// Decode reads and analyzes the given reader as a GIF image
func SplitAnimatedGIF(reader io.Reader, timesFry int) (err error, newGif *gif.GIF) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("Error while decoding: %s", r)
		}
	}()
	inGif, err := gif.DecodeAll(reader)
	fryGif := gif.GIF{}

	if err != nil {
		return err, nil
	}

	for _, srcImg := range inGif.Image {
		var imgCfg image.Config
		imgCfg.Width, imgCfg.Height = srcImg.Rect.Dx(), srcImg.Rect.Dy()
		alteredImage := ditherFilter(srcImg, imgCfg, timesFry)
		bounds := alteredImage.Bounds()
		alteredPalette := image.NewPaletted(bounds, srcImg.Palette)
		draw.Draw(alteredPalette, alteredPalette.Rect, alteredImage, bounds.Min, draw.Over)

		// save current frame "stack". This will overwrite an existing file with that name
		fryGif.Delay = append(fryGif.Delay, 8)
		fryGif.Image = append(fryGif.Image, alteredPalette)
	}
	//gif.EncodeAll(out, &fryGif) //ignores encoding errors
	return nil, &fryGif
}