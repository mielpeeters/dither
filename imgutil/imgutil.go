// Package imgutil offers some useful general functions for working with images within the
// dither module.
package imgutil

import (
	"fmt"
	"image"
	"image/color"
	"image/gif"
	"image/jpeg"
	"image/png"
	"os"
	"sync"
)

const reset = "\033[0m"
const cyan = "\033[36m"
const green = "\033[32m"
const itallic = "\033[3m"
const bold = "\033[1m"
const red = "\033[31m"
const blink = "\033[5m"

// ImageToPixels converts an image.Image instance
// into a column-major array
//
// More specifically, a pointer to a [][]color.Color array is returned,
// indexed like so: color @ (x, y) -> (*return)[x][y]
// where x indexes the outer slice, selecting one column, and y indexes those slices
func ImageToPixels(img image.Image) *[][]color.Color {
	size := img.Bounds().Size()
	var pixels [][]color.Color
	// put pixels into two dimensional array
	// for every x value, store an array
	for i := 0; i < size.X; i++ {
		var y []color.Color
		for j := 0; j < size.Y; j++ {
			y = append(y, img.At(i, j))
		}
		pixels = append(pixels, y)
	}

	return &pixels
}

// PixelsToImage creates an image.RGBA from the given slice of color.Color slices
func PixelsToImage(pixels *[][]color.Color) *image.RGBA {
	rect := image.Rect(0, 0, len(*pixels), len((*pixels)[0]))
	nImg := image.NewRGBA(rect)

	wg := sync.WaitGroup{}

	for x := 0; x < len(*pixels); x++ {
		wg.Add(1)
		go func(x int) {
			for y := 0; y < len((*pixels)[0]); y++ {
				if (*pixels)[x] == nil {
					continue
				}
				p := (*pixels)[x][y]
				if p == nil {
					continue
				}
				original, ok := color.RGBAModel.Convert(p).(color.RGBA)
				if ok {
					nImg.Set(x, y, original)
				}
			}

			wg.Done()
		}(x)
	}

	wg.Wait()

	return nImg
}

// OpenImage opens an image by providing a path.
func OpenImage(path string) (image.Image, error) {
	f, err := os.Open(path)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	defer f.Close()

	img, _, err := image.Decode(f)
	if err != nil {
		fmt.Println("Decoding error:", err.Error())
		return nil, err
	}
	return img, nil
}

// SavePNG saves img at path as a PNG file
func SavePNG(img image.Image, name string) {
	f, err := os.Create(name)
	if err != nil {
		fmt.Println("couldn't save")
	}
	defer f.Close()

	// Encode to `PNG` with `DefaultCompression` level
	// then save to file
	err = png.Encode(f, img)
	if err != nil {
		fmt.Println("couldn't save")
	}
}

// SaveGIF saves img at path as a GIF file
func SaveGIF(img image.Image, name string) {
	f, err := os.Create(name)
	if err != nil {
		fmt.Println("couldn't save")
	}
	defer f.Close()

	err = gif.Encode(f, img, nil)
	if err != nil {
		fmt.Println("Couldn't save")
	}
}

// SaveJPEG saves img at path as a JPEG file, with specified JPEG quality
func SaveJPEG(img image.Image, name string, quality int) {
	f, err := os.Create(name)
	if err != nil {
		fmt.Println("couldn't save")
	}
	defer f.Close()

	// Encode to `PNG` with `DefaultCompression` level
	// then save to file

	opt := jpeg.Options{
		Quality: quality,
	}

	err = jpeg.Encode(f, img, &opt)

	if err != nil {
		fmt.Println("couldn't save")
	}
}
