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
}

func getSampleFactor(scaleFactor int) int {
	output := 12 - 0.8*float64(scaleFactor)
	outputInt := int(output)
	if outputInt < 0 {
		outputInt = 0
	}
	return outputInt
}

func main() {
	//os.Args[1]: image path
	//os.Args[2]: downscale factor
	//os.Args[3]: colorpalette name, set to FromImage to generate
	//os.Args[4]: only with FromImage: the amount of colors to be extracted

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

	img.Bounds()

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
			paletted := floydSteinbergDithering(pixels, palette, scaleVar, Y, X)
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
			defer file.Close()
			if err != nil {
				fmt.Println("Error create file")
			}
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
		downscaleNoUpscale(pixels, *scaleFactor)

		var palette ColorPalette

		if *paletteName != "FromImage" {
			palette = getPaletteWithName(*paletteName, palettes)
		} else {
			sampleFactor := getSampleFactor(*scaleFactor)
			palette = createColorPalette(pixels, *amountOfColors, sampleFactor, *amountKnnRuns)
		}

		floydSteinbergDithering(pixels, palette, *scaleFactor, Y, X)

		upscale(pixels, *scaleFactor)

		nImg := pixelsToImage(pixels)

		savePNG(nImg, *outputName)
	}
}
