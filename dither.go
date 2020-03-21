// imagefry applies pseudo random filter to an image
package main

import (
	"fmt"
	"image"
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

	"./filters"
)

func init() {
	image.RegisterFormat("png", "png", png.Decode, png.DecodeConfig)
	image.RegisterFormat("jpeg", "jpeg", jpeg.Decode, jpeg.DecodeConfig)
}

func pickFilter(input string) string {
	fmt.Println(input)
	filter := "rand"
	if input == "dither" {
		fmt.Println("dithering")
		filter = "dither"
	} else if input == "ditherc" {
		fmt.Println("dithering with color")
		filter = "ditherc"
	} else if input == "xor" {
		fmt.Println("applying xor")
		filter = "xor"
	} else {
		return filter
	}

	return filter
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

	filter := pickFilter(os.Args[3])
	fmt.Println(filter)

	var newImg image.Image
	if filter == "rand" {
		newImg = filters.RandFilter(imageData, imgCfg, timesFry)
	} else if filter == "dither" {
		newImg = filters.DitherFilter(imageData, imgCfg, timesFry)
	} else if filter == "ditherc" {
		newImg = filters.DitherFilterColor(imageData, imgCfg, timesFry)
	} else if filter == "xor" {
		newImg = filters.XorFilter(imageData, imgCfg, timesFry)
	} else {
		fmt.Println("error, bad filter arg")
	}

	return newImg, imageType
}

func main() {
	rand.Seed(time.Now().UTC().UnixNano())
	ImageFile, timesFry := inputParseAndOpen()
	defer ImageFile.Close()

	if strings.HasSuffix(os.Args[1], ".gif") {
		_, newGif := SplitAnimatedGIF(ImageFile, timesFry)
		outputFile, err := os.Create(os.Args[1])
		if err != nil {
			fmt.Fprintf(os.Stderr, "perlin output error: %s\n", err)
		}
		gif.EncodeAll(outputFile, newGif)
		outputFile.Close()
	} else {
		newImg, imageType := openDecodeFilterStatic(ImageFile, timesFry)

		outputFile, err := os.Create(os.Args[1])
		if err != nil {
			fmt.Fprintf(os.Stderr, "perlin output error: %s\n", err)
		}

		fmt.Println(newImg)

		if imageType == "png" {
			png.Encode(outputFile, newImg)
		} else if imageType == "jpeg" {
			jpeg.Encode(outputFile, newImg, nil)
		} else {
			fmt.Println("ERROR: unrecognized file format")
		}
		outputFile.Close()
	}

	fmt.Printf("output written to %s\n", os.Args[1])
}

func inputParseAndOpen() (*os.File, int) {
	if len(os.Args) <= 1 || os.Args[1] == "help" {
		fmt.Fprint(os.Stderr, "ERROR: please provide a filename\n")
		fmt.Println("USAGE: 'imagefry.exe <image.jpg/png/gif> <#times_fryd> <filter_type>'")
		fmt.Println("<filter_type> options are 'rand' or 'dither'")
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
		} else if len(os.Args) > 4 {
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

	filter := pickFilter(os.Args[3])
	fmt.Println(filter)

	for _, srcImg := range inGif.Image {
		var imgCfg image.Config
		imgCfg.Width, imgCfg.Height = srcImg.Rect.Dx(), srcImg.Rect.Dy()
		var alteredImage image.Image
		if filter == "rand" {
			alteredImage = filters.RandFilter(srcImg, imgCfg, timesFry)
		} else if filter == "dither" {
			alteredImage = filters.DitherFilter(srcImg, imgCfg, timesFry)
		} else if filter == "ditherc" {
			alteredImage = filters.DitherFilterColor(srcImg, imgCfg, timesFry)
		} else if filter == "xor" {
			alteredImage = filters.XorFilter(srcImg, imgCfg, timesFry)
		} else {
			fmt.Println("error, bad filter arg")
		}

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