package main

import (
	"flag"
	"strings"
)

var (
	imagePath      *string
	scaleFactor    *int
	paletteName    *string
	outputPath     *string
	amountOfColors *int
	outputName     *string
	amountKnnRuns  *int
	xPixels        *int
	scale          int
	paletteOutput  *bool
)

func init() {
	imagePath = flag.String("p", "", "the path of the input image") //check later for non named thing
	scaleFactor = flag.Int("scale", 10, "factor by which the image is downscaled")
	paletteName = flag.String("colors", "FromImage", "the colorpalette to be used for dithering")
	outputPath = flag.String("o", "output.png", "the desired output name and type")
	amountOfColors = flag.Int("k", 10, "the amount of colors to use (no -colors flag specified!)")
	amountKnnRuns = flag.Int("knn", 5, "the amount of KNN runs with random initialization.")
	xPixels = flag.Int("x", 0, "amount of pixels in x direction")
	paletteOutput = flag.Bool("showPalette", false, "output an image \"selectedPalette.png\" with used colorpalette.")
}

func getSampleFactor(scaleFactor int) int {
	output := 12 - 0.2*float64(scaleFactor)
	outputInt := int(output)
	if outputInt < 1 {
		outputInt = 1
	}
	return outputInt
}

func main() {
	flag.Parse()

	// get the output file name and file type
	splitted := strings.Split(*outputPath, ".")
	outputType := &splitted[1]
	outputName = &splitted[0]

	// open the input image
	img, err := openImage(*imagePath)
	if err != nil {
		return
	}

	// convert the input into pixels
	pixels := imageToPixels(img)
	X := len(*pixels)
	Y := len((*pixels)[0])

	if *xPixels != 0 {
		scale = X / *xPixels
	} else {
		scale = *scaleFactor
	}

	// scale the image down
	downscaleNoUpscale(pixels, scale)

	// get the palette in which to create the new image
	var palette ColorPalette
	if *paletteName != "FromImage" {
		palettes := getPalettesFromJson("colorpalette.json")
		palette = getPaletteWithName(*paletteName, palettes)
	} else {
		sampleFactor := getSampleFactor(scale)
		palette = createColorPalette(pixels, *amountOfColors, sampleFactor, *amountKnnRuns)
	}

	// write the used palette to the output if needed
	if *xPixels != 0 || *paletteOutput {
		paletteToJsonFile(palette, "selectedPalette.json")
		paletteToImage(palette, "selectedPalette")
	}

	// apply dithering to the image for a nicer effect
	pixels, paletted := floydSteinbergDithering(pixels, palette, scale, Y, X)

	// save the image, either PNG or GIF
	if *outputType != "gif" {
		upscale(pixels, scale)

		nImg := pixelsToImage(pixels)

		savePNG(nImg, *outputName)
	} else {
		saveGIF(paletted, *outputName)
	}
}
