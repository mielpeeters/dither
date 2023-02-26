package main

import (
	"log"
	"testing"
	"time"
)

type testArgs struct {
	image  string
	scale  int
	colors int
}

func printArgs(args *testArgs, elapsed time.Duration) {
	log.Println(Bold, "\nRunning test:\n  - image:", (*args).image, "\n  - scale:", (*args).scale, "\n  - colors:", (*args).colors, Reset)
	log.Println(Bold, Blink, Green, "\nTest took", elapsed, Reset)
}

func runTest(args *testArgs) time.Duration {

	img, err := openImage((*args).image)
	if err != nil {
		log.Println((*args))
		log.Fatal("Couldn't get image opened")
	}

	pixels := imageToPixels(img)

	X := len(*pixels)
	Y := len((*pixels)[0])

	downscaleNoUpscale(pixels, (*args).scale)

	palettes := getPalettesFromJson("colorpalette.json")
	palette := getPaletteWithName("GameBoy", palettes)

	start := time.Now()

	floydSteinbergDithering(pixels, palette, scale, Y, X)

	return time.Since(start)
}

func TestSpeed(t *testing.T) {
	log.Println(Bold, "\nStarting Program Speed Test.\n", Reset)

	args := testArgs{
		"src/ripZeep.jpg",
		1,
		10,
	}
	printArgs(&args, runTest(&args))

}
