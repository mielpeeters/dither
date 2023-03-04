package main

import (
	"flag"
	"fmt"
	"image"
	"image/gif"
	"log"
	"os"
	"strings"
)

var (
	imagePath      *string
	scaleFactor    *int
	paletteName    *string
	outputType     *string
	gifFrames      *int
	outputPath     *string
	amountOfColors *int
	outputName     *string
	fromImage      bool
	scaleGif       *bool
	amountKnnRuns  *int
	xPixels        *int
	scale          int
	paletteOutput  *bool
)

func init() {
	imagePath = flag.String("p", "", "the path of the input image") //check later for non named thing
	scaleFactor = flag.Int("scale", 10, "factor by which the image is downscaled")
	paletteName = flag.String("colors", "FromImage", "the colorpalette to be used for dithering")
	gifFrames = flag.Int("frames", 15, "amount of different frames in gif format")
	outputPath = flag.String("o", "output.png", "the desired output name and type")
	amountOfColors = flag.Int("k", 10, "the amount of colors to use (no -colors flag specified!)")
	amountKnnRuns = flag.Int("knn", 5, "the amount of KNN runs with random initialization.")
	scaleGif = flag.Bool("scaleGif", false, "yeah")
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

	splitted := strings.Split(*outputPath, ".")
	outputType = &splitted[1]
	outputName = &splitted[0]

	if *outputType == "gif" {
		if *paletteName != "FromImage" {
			log.Fatal("Cannot specify colorpalette with gif format...")
		}

		if *amountOfColors != 10 {
			log.Fatal("Cannot specify amount of colors with gif output format...")
		}

	} else {
		if *gifFrames != 15 {
			log.Fatal("Cannot specify ao frames without gif format...")
		}

		if *scaleGif {
			log.Fatal("Cannot specify scaleGif without gif format...")
		}
	}

	if *amountOfColors != 10 {
		if *paletteName != "FromImage" {
			log.Fatal("Cannot specify amount of colors for already specified colorpalette... (remove -colors)")
		}
	}

	if *paletteName == "FromImage" {
		fromImage = true
	} else {
		fromImage = false
	}

	palettes := getPalettesFromJson("colorpalette.json")

	img, err := openImage(*imagePath)
	if err != nil {
		return
	}

	if *outputType == "gif" || *outputType == "mp4" {
		var images []*image.Paletted
		var scaleVar int

		if *scaleGif {
			scaleVar = *scaleFactor + *gifFrames*3
		} else {
			scaleVar = *scaleFactor
		}
		for i := 1; i <= *gifFrames; i++ {
			pixels := imageToPixels(img)
			X := len(*pixels)
			Y := len((*pixels)[0])
			downscaleNoUpscale(pixels, scaleVar)
			palette := createColorPalette(pixels, i, 4, *amountKnnRuns)
			_, paletted := floydSteinbergDithering(pixels, palette, scaleVar, Y, X)
			//upscale(pixels, 20)
			images = append(images, paletted)

			if *scaleGif {
				scaleVar -= 3
			}
		}

		for i := 1; i < *gifFrames; i++ {
			images = append(images, images[*gifFrames-i-1])
		}

		if *outputType == "gif" {
			anim := gif.GIF{LoopCount: *gifFrames * 2}

			for _, img := range images {
				anim.Image = append(anim.Image, img)
				anim.Delay = append(anim.Delay, 0)
			}

			output := fmt.Sprintf(*outputPath)
			file, err := os.Create(output)
			if err != nil {
				fmt.Println("Error create file")
			}
			defer file.Close()
			gif.EncodeAll(file, &anim)
		} else {
			for index, img := range images {
				savePNG(img, "temp/"+*outputName+fmt.Sprint(index))
			}
		}

		fmt.Println("\nFile saved.")
	} else {

		pixels := imageToPixels(img)
		X := len(*pixels)
		Y := len((*pixels)[0])

		if *xPixels != 0 {
			scale = X / *xPixels
		} else {
			scale = *scaleFactor
		}

		downscaleNoUpscale(pixels, scale)

		var palette ColorPalette

		if *paletteName != "FromImage" {
			palette = getPaletteWithName(*paletteName, palettes)
		} else {
			sampleFactor := getSampleFactor(scale)
			palette = createColorPalette(pixels, *amountOfColors, sampleFactor, *amountKnnRuns)
		}

		if *xPixels != 0 || *paletteOutput {
			paletteToJsonFile(palette, "selectedPalette.json")
			paletteToImage(palette, "selectedPalette")
		}

		pixels, _ = floydSteinbergDithering(pixels, palette, scale, Y, X)

		upscale(pixels, scale)

		nImg := pixelsToImage(pixels)

		savePNG(nImg, *outputName)
	}
}
