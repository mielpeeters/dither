package main

import (
	"fmt"
	// "image/color"
	"os"
	"strconv"
)

func main() {
	//os.Args[1]: image path
	//os.Args[2]: downscale factor
	//os.Args[3]: colorpalette name, set to FromImage to generate
	//os.Args[4]: only with FromImage: the amount of colors to be extracted

	palettes := getPalettesFromJson("colorpalette.json")
	if len(os.Args) == 1 {
		fmt.Println(palettes)
		return
	}

	fmt.Println("Start processing that image!")

	img, err := openImage(os.Args[1])
	if err != nil {
		return
	}

	pixels := imageToPixels(img)
	integerScale, err := strconv.Atoi(os.Args[2])
	if err != nil {
		fmt.Println("Couln't convert that input into a integer")
	}

	//first, downscale the image to reduce the payload
	if integerScale != 1 {
		downscaleNoUpscale(pixels, integerScale)
	}

	var palette ColorPalette
	var kd string
	if os.Args[3] != "FromImage" {
		palette = getPaletteWithName(os.Args[3], palettes)
		kd = os.Args[4]
	} else {
		integer, err := strconv.Atoi(os.Args[4])
		if err != nil {
			fmt.Println("Couln't convert that input into a integer")
		}
		palette = createColorPalette(pixels, integer, 4)
		kd = os.Args[5]
	}

	floydSteinbergDithering(pixels, palette, kd)

	upscale(pixels, integerScale)

	nImg := pixelsToImage(pixels)

	savePNG(nImg, "output")
}
