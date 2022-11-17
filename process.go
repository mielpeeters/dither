package main

import (
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"math"
	"os"
	"sync"
	"time"
)

// rando mcomment
type ErrorColor struct {
	R int16
	G int16
	B int16
	A int16
}

func roundDown(number float64) int {
	return int(math.Floor(number))
}

func createColorPalette(pixels *[][]color.Color, k int, samplefactor int) ColorPalette {
	pointSet := PointSet{}
	// sample only 1/samplefactor of the pixels
	for i := 0; i < len((*pixels)); i += samplefactor {
		for j := 0; j < len((*pixels)[0]); j += samplefactor {
			pointSet.Points = append(pointSet.Points, colorToPoint((*pixels)[i][j]))
			pointSet.Points[len(pointSet.Points)-1].Id = i*(len((*pixels))/samplefactor) + j
		}
	}

	KM := createKMeansProblem(pointSet, k, redMeanDistance)

	var done bool
	var iteration int
	var consecutiveDone int
	for consecutiveDone < 4 {
		iteration++
		done, _ = KM.iterate(0.0001)
		if done {
			consecutiveDone++
		} else {
			consecutiveDone = 0
		}
	}

	colorPalette := ColorPalette{}
	for index := range KM.kMeans.Points {
		colorPalette.Colors = append(colorPalette.Colors, pointToColorSlice(KM.kMeans.Points[index]))
	}

	return colorPalette
}

func pointToColorSlice(point Point) []int {
	returnValue := []int{}

	for _, value := range point.Coordinates {
		returnValue = append(returnValue, int(value))
	}

	return returnValue
}

func downscaleNoUpscale(pixels *[][]color.Color, factor int) {

	ppixels := *pixels
	iLen := len(ppixels)
	jLen := len(ppixels[0])

	//create new image
	newImage := make([][]color.Color, roundDown(float64(iLen/factor))) ///float64(factor)))
	for i := 0; i < len(newImage); i++ {
		newImage[i] = make([]color.Color, roundDown(float64(jLen/factor))) ///float64(factor)))
	}

	wg := sync.WaitGroup{}

	for i := 0; i < iLen/factor; i++ {
		wg.Add(1) //do each row in parallel
		go func(i int) {
			for j := 0; j < jLen/factor; j++ {

				sumR := float64(0)
				sumG := float64(0)
				sumB := float64(0)
				sumA := float64(0)

				for k := 0; k < factor; k++ {
					for m := 0; m < factor; m++ {
						if !(i*factor+k > iLen) {
							if !(j*factor+k > jLen) {

								pixel := ppixels[i*factor+k][j*factor+m]

								originalColor, ok := color.RGBAModel.Convert(pixel).(color.RGBA)
								if !ok {
									fmt.Println("type conversion went wrong")
								}
								sumR += float64(originalColor.R)
								sumG += float64(originalColor.G)
								sumB += float64(originalColor.B)
								sumA += float64(originalColor.A)
							}
						}
					}
				}

				col := color.RGBA{
					uint8(sumR / math.Pow(float64(factor), 2)),
					uint8(sumG / math.Pow(float64(factor), 2)),
					uint8(sumB / math.Pow(float64(factor), 2)),
					uint8(sumA / math.Pow(float64(factor), 2)),
				}

				newImage[i][j] = col

			}
			wg.Done()
		}(i)
	}
	wg.Wait()
	*pixels = newImage

}

func upscale(pixels *[][]color.Color, factor int) {

	ppixels := *pixels
	iLen := len(ppixels)
	jLen := len(ppixels[0])

	//create new image
	newImage := make([][]color.Color, roundDown(float64(iLen*factor)))
	for i := 0; i < len(newImage); i++ {
		newImage[i] = make([]color.Color, roundDown(float64(jLen*factor)))
	}

	wg := sync.WaitGroup{}

	for i := 0; i < iLen; i++ {
		wg.Add(1) //do each row in parallel
		go func(i int) {
			for j := 0; j < jLen; j++ {

				pixel := ppixels[i][j]

				for k := 0; k < factor; k++ {
					for l := 0; l < factor; l++ {
						newImage[i*factor+k][j*factor+l] = pixel
					}
				}

			}
			wg.Done()
		}(i)
	}
	wg.Wait()
	*pixels = newImage
}

func downscale(pixels *[][]color.Color, factor int) {

	ppixels := *pixels
	iLen := len(ppixels)
	jLen := len(ppixels[0])

	//create new image
	newImage := make([][]color.Color, roundDown(float64(iLen))) ///float64(factor)))
	for i := 0; i < len(newImage); i++ {
		newImage[i] = make([]color.Color, roundDown(float64(jLen))) ///float64(factor)))
	}

	wg := sync.WaitGroup{}

	for i := 0; i < iLen/factor; i++ {

		for j := 0; j < jLen/factor; j++ {

			wg.Add(1)
			go func(i, j int) {
				sumR := float64(0)
				sumG := float64(0)
				sumB := float64(0)
				sumA := float64(0)

				for k := 0; k < factor; k++ {
					for m := 0; m < factor; m++ {
						if !(i*factor+k > iLen) {
							if !(j*factor+k > jLen) {

								pixel := ppixels[i*factor+k][j*factor+m]

								originalColor, ok := color.RGBAModel.Convert(pixel).(color.RGBA)
								if !ok {
									fmt.Println("type conversion went wrong")
								}
								sumR += float64(originalColor.R)
								sumG += float64(originalColor.G)
								sumB += float64(originalColor.B)
								sumA += float64(originalColor.A)
							}
						}
					}
				}

				col := color.RGBA{
					uint8(sumR / math.Pow(float64(factor), 2)),
					uint8(sumG / math.Pow(float64(factor), 2)),
					uint8(sumB / math.Pow(float64(factor), 2)),
					uint8(sumA / math.Pow(float64(factor), 2)),
				}

				for k := 0; k < factor; k++ {
					for m := 0; m < factor; m++ {
						newImage[i*factor+k][j*factor+m] = col
					}
				}
				wg.Done()
			}(i, j)
		}
	}
	wg.Wait()
	*pixels = newImage
}

func addColorComponents(left int16, right int16) uint8 {
	result := left + right

	if result < 0 {
		result = 0
	}

	if result > 255 {
		result = 255
	}

	return uint8(result)
}

func addErrorToColor(errorColor ErrorColor, origColor color.Color, factor float64) color.Color {
	orig, ok := color.RGBAModel.Convert(origColor).(color.RGBA)
	if !ok {
		fmt.Println("type conversion (to rgba color) went wrong")
	}

	col := color.RGBA{
		addColorComponents(int16(orig.R), int16(float64(errorColor.R)*factor)),
		addColorComponents(int16(orig.G), int16(float64(errorColor.G)*factor)),
		addColorComponents(int16(orig.B), int16(float64(errorColor.B)*factor)),
		addColorComponents(int16(orig.A), int16(float64(errorColor.A)*factor)),
	}

	// printRGBAColor(orig, "original neighboring color")
	// printRGBAColor(col, "added noise neighboring color")

	return col
}

func getColorDifference(left color.Color, right color.Color) ErrorColor {
	leftRGBA := toRGBA(left)
	rightRGBA := toRGBA(right)

	// fmt.Println("differences: ", int8(leftRGBA.R)-int8(rightRGBA.R),
	// 	int8(leftRGBA.G)-int8(rightRGBA.G),
	// 	int8(leftRGBA.B)-int8(rightRGBA.B),
	// 	int8(leftRGBA.A)-int8(rightRGBA.A))

	col := ErrorColor{
		int16(leftRGBA.R) - int16(rightRGBA.R),
		int16(leftRGBA.G) - int16(rightRGBA.G),
		int16(leftRGBA.B) - int16(rightRGBA.B),
		int16(leftRGBA.A) - int16(rightRGBA.A),
	}

	return col
}

func toRGBA(origColor color.Color) color.RGBA {
	orig, ok := color.RGBAModel.Convert(origColor).(color.RGBA)
	if !ok {
		fmt.Println("type conversion (to rgba color) went wrong")
	}
	return orig
}

func colorToPoint(clr color.Color) Point {
	clrRGBA := toRGBA(clr)
	coordinates := []float64{float64(clrRGBA.R), float64(clrRGBA.G), float64(clrRGBA.B), float64(clrRGBA.A)}
	//coordinates = RGBAtoHSLA(coordinates)
	point := Point{
		coordinates,
		0,
	}
	return point
}

func makeColor(R, G, B, A int) color.Color {
	col := color.RGBA{
		uint8(R),
		uint8(G),
		uint8(B),
		uint8(A),
	}

	return col
}

func pointToColor(point Point) color.Color {
	//rgba := HSLAtoRGBA(point.Coordinates)
	col := color.RGBA{
		uint8(point.Coordinates[0]),
		uint8(point.Coordinates[1]),
		uint8(point.Coordinates[2]),
		uint8(point.Coordinates[3]),
	}

	return col
}

func paletteToNeighbors(palette ColorPalette) []Point {
	var neighbors []Point
	for _, clr := range palette.Colors {
		colour := color.RGBA{
			uint8(clr[0]),
			uint8(clr[1]),
			uint8(clr[2]),
			uint8(clr[3]),
		}
		neighbors = append(neighbors, colorToPoint(colour))
	}
	return neighbors
}

func squaresDistance(pnt1 Point, pnt2 Point) float64 {
	var dist float64
	for index := range pnt1.Coordinates {
		dist += math.Pow((pnt1.Coordinates[index] - pnt2.Coordinates[index]), 2)
	}

	return dist
}

func redMeanDistance(pnt1, pnt2 Point) float64 {
	// only to use with colors!
	redMean := (pnt1.Coordinates[0] + pnt2.Coordinates[0]) / 2

	output := (2 + redMean/256) * math.Pow(pnt1.Coordinates[0]-pnt2.Coordinates[0], 2)

	output += 4 * math.Pow(pnt1.Coordinates[1]-pnt2.Coordinates[1], 2)

	output += (2 + (255-redMean)/256) * math.Pow(pnt1.Coordinates[2]-pnt2.Coordinates[2], 2)

	return output
}

func floydSteinbergDithering(pixels *[][]color.Color, palette ColorPalette, upscale, X, Y int) *image.Paletted {
	var neighborTime time.Duration

	newPixels := *pixels
	yLen := len(newPixels)
	xLen := len(newPixels[0])

	upLeft := image.Point{0, 0}
	lowRight := image.Point{Y, X}
	r := image.Rectangle{upLeft, lowRight}

	p := colorPaletteToPalette(palette)

	newImage := image.NewPaletted(r, p)

	for y := 0; y < yLen; y++ {
		for x := 0; x < xLen; x++ {
			oldPixel := newPixels[y][x]

			start := time.Now()

			newPixel := p.Convert(oldPixel)

			neighborTime += time.Since(start)

			err := getColorDifference(oldPixel, newPixel)

			index := p.Index(oldPixel)

			for i := 0; i < upscale; i++ {
				for j := 0; j < upscale; j++ {
					newImage.Pix[(y*upscale+i)+(x*upscale+j)*newImage.Stride] = uint8(index)
				}
			}

			(*pixels)[y][x] = newPixel

			if x+1 < xLen {
				(*pixels)[y][x+1] = addErrorToColor(err, (*pixels)[y][x+1], 7.0/16.0)
			}
			if x-1 > 0 && y+1 < yLen {
				(*pixels)[y+1][x-1] = addErrorToColor(err, (*pixels)[y+1][x-1], 3.0/16.0)
			}
			if y+1 < yLen {
				(*pixels)[y+1][x] = addErrorToColor(err, (*pixels)[y+1][x], 5.0/16.0)
			}
			if x+1 < xLen && y+1 < yLen {
				(*pixels)[y+1][x+1] = addErrorToColor(err, (*pixels)[y+1][x+1], 1.0/16.0)
			}
		}
	}

	return newImage
}

func printRGBAColor(col color.RGBA, title string) {
	fmt.Println(title)
	fmt.Println("R: ", col.R)
	fmt.Println("G: ", col.G)
	fmt.Println("B: ", col.B)
	fmt.Println("A: ", col.A)
}

func printErrorColor(col ErrorColor, title string) {
	fmt.Println(title)
	fmt.Println("R: ", col.R)
	fmt.Println("G: ", col.G)
	fmt.Println("B: ", col.B)
	fmt.Println("A: ", col.A)
}

func openImage(path string) (image.Image, error) {
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
	// if format != "jpeg" {
	// 	fmt.Println("image format is not jpeg")
	// 	return nil, errors.New("")
	// }
	return img, nil
}

func imageToPixels(img image.Image) *[][]color.Color {
	size := img.Bounds().Size()
	var pixels [][]color.Color
	//put pixels into two three two dimensional array
	for i := 0; i < size.X; i++ {
		var y []color.Color
		for j := 0; j < size.Y; j++ {
			y = append(y, img.At(i, j))
		}
		pixels = append(pixels, y)
	}

	return &pixels
}

func pixelsToImage(pixels *[][]color.Color) *image.RGBA {
	rect := image.Rect(0, 0, len(*pixels), len((*pixels)[0]))
	nImg := image.NewRGBA(rect)

	for x := 0; x < len(*pixels); x++ {
		for y := 0; y < len((*pixels)[0]); y++ {
			q := (*pixels)[x]
			if q == nil {
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
	}

	return nImg
}

func savePNG(img image.Image, name string) {
	f, err := os.Create(name + ".png")
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

func saveJPEG(img image.Image, name string, quality int) {
	f, err := os.Create(name + ".jpeg")
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

func colorPaletteToPalette(colorpalette ColorPalette) color.Palette {
	colors := []color.Color{}
	var paletteColor color.Color

	for i := 0; i < len(colorpalette.Colors); i++ {
		paletteColor = color.RGBA{
			uint8(colorpalette.Colors[i][0]),
			uint8(colorpalette.Colors[i][1]),
			uint8(colorpalette.Colors[i][2]),
			uint8(colorpalette.Colors[i][3]),
		}
		colors = append(colors, paletteColor)
	}

	var palette color.Palette
	palette = colors

	return palette
}
