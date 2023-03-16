package process

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"sync"

	"github.com/mielpeeters/dither/colorpalette"
	"github.com/mielpeeters/dither/geom"
	"github.com/mielpeeters/dither/kmeans"
)

type errorColor struct {
	R int16
	G int16
	B int16
	A int16
}

func roundDown(number float64) int {
	return int(math.Floor(number))
}

func findMinIndex(arr []float64) int {
	min := math.Inf(1) // Initialize min with the highest possible float64 value
	minIndex := 0      // Initialize minIndex with 0
	for i, v := range arr {
		if v < min {
			min = v
			minIndex = i
		}
	}
	return minIndex
}

const reset = "\033[0m"
const cyan = "\033[36m"
const green = "\033[32m"
const itallic = "\033[3m"
const bold = "\033[1m"
const red = "\033[31m"
const blink = "\033[5m"

// CreateColorPalette creates a new colorpalette using the k-means clustering algorithm
//
//   - samplefactor: how many pixles to skip, during sampling for the creatrion of the KMeans problem's cluster points
//     (higher means faster, because less points to iterate over)
//   - kmTimes defines the amount of times to start the k-means algorithm with random init, the best output is choosen
func CreateColorPalette(pixels *[][]color.Color, k int, samplefactor int, kmTimes int) colorpalette.ColorPalette {
	fmt.Println(cyan + bold + "Creating color palette (knn)..." + reset)
	pointSet := geom.PointSet{}
	// sample only 1/samplefactor of the pixels
	for i := 0; i < len((*pixels)); i += samplefactor {
		for j := 0; j < len((*pixels)[0]); j += samplefactor {
			pointSet.Points = append(pointSet.Points, colorToPoint((*pixels)[i][j]))
			pointSet.Points[len(pointSet.Points)-1].ID = i*(len((*pixels))/samplefactor) + j
		}
	}

	var colorPalettes []colorpalette.ColorPalette
	var errors []float64

	// do the algorithm kmTimes
	for i := 0; i < kmTimes; i++ {
		KM := kmeans.CreateKMeansProblem(pointSet, k, geom.RedMeanDistance)

		KM.Cluster(0.001, 2)

		colorPalette := colorpalette.ColorPalette{}
		for index := range KM.KMeans.Points {
			colorPalette.Colors = append(colorPalette.Colors, pointToColorSlice(KM.KMeans.Points[index]))
		}

		colorPalettes = append(colorPalettes, colorPalette)
		errors = append(errors, KM.TotalDist())
	}

	// now select the colorpalette with the lowest error!
	minIndex := findMinIndex(errors)

	fmt.Println(green + itallic + "	Done!!" + reset)
	return colorPalettes[minIndex]
}

func pointToColorSlice(point geom.Point) []int {
	returnValue := []int{}

	for _, value := range point.Coordinates {
		returnValue = append(returnValue, int(value))
	}

	return returnValue
}

// Downscale scales the image down with a given integer factor
func Downscale(pixels *[][]color.Color, factor int) {
	fmt.Println(cyan + bold + "Downscaling..." + reset)
	iLen := len(*pixels)
	jLen := len((*pixels)[0])

	//create new image
	newImage := make([][]color.Color, roundDown(float64(iLen/factor)))
	for i := 0; i < len(newImage); i++ {
		newImage[i] = make([]color.Color, roundDown(float64(jLen/factor)))
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

				for k := 0; k < factor && i*factor+k <= iLen; k++ {
					for m := 0; m < factor && j*factor+m <= jLen; m++ {
						pixel := (*pixels)[i*factor+k][j*factor+m]

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
	fmt.Println(green + itallic + "	Done!" + reset)
}

// Upscale scales the input image up with the given integer factor
func Upscale(pixels *[][]color.Color, factor int) {
	fmt.Println(cyan + bold + "Upscaling..." + reset)
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
	fmt.Println(green + itallic + "	Done!" + reset)
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

func addErrorToColor(errorColor errorColor, origColor color.Color, factor float64) color.Color {
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

func getColorDifference(left color.Color, right color.Color) errorColor {
	leftRGBA := colorpalette.ToRGBA(left)
	rightRGBA := colorpalette.ToRGBA(right)

	col := errorColor{
		R: int16(leftRGBA.R) - int16(rightRGBA.R),
		G: int16(leftRGBA.G) - int16(rightRGBA.G),
		B: int16(leftRGBA.B) - int16(rightRGBA.B),
		A: int16(leftRGBA.A) - int16(rightRGBA.A),
	}

	return col
}

func colorToPoint(clr color.Color) geom.Point {
	clrRGBA := colorpalette.ToRGBA(clr)
	coordinates := []float32{float32(clrRGBA.R), float32(clrRGBA.G), float32(clrRGBA.B), float32(clrRGBA.A)}
	//coordinates = RGBAtoHSLA(coordinates)
	point := geom.Point{
		Coordinates: coordinates,
		ID:          0,
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

func pointToColor(point geom.Point) color.Color {
	//rgba := HSLAtoRGBA(point.Coordinates)
	col := color.RGBA{
		uint8(point.Coordinates[0]),
		uint8(point.Coordinates[1]),
		uint8(point.Coordinates[2]),
		uint8(point.Coordinates[3]),
	}

	return col
}

// FloydSteinbergDithering applies the FS dithering effect
func FloydSteinbergDithering(pixels *[][]color.Color, palette colorpalette.ColorPalette, upscale, X, Y int) (*[][]color.Color, *image.Paletted) {
	fmt.Println(cyan + bold + "Starting dithering process!" + reset)

	yLen := len(*pixels)
	xLen := len((*pixels)[0])

	upLeft := image.Point{0, 0}
	lowRight := image.Point{yLen, xLen}
	r := image.Rectangle{upLeft, lowRight}

	p := palette.ToPalette()

	newImage := image.NewPaletted(r, p)

	for y := 0; y < yLen; y++ {
		for x := 0; x < xLen; x++ {
			oldPixel := (*pixels)[y][x]

			colorIndex := uint8(p.Index(oldPixel))
			(*pixels)[y][x] = p[colorIndex]

			err := getColorDifference(oldPixel, (*pixels)[y][x])

			// for i := 0; i < upscale; i++ {
			// 	for j := 0; j < upscale; j++ {
			// 		newImage.Pix[(y*upscale+i)+(x*upscale+j)*newImage.Stride] = colorIndex
			// 	}
			// }

			newImage.Set(y, x, oldPixel)

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

	fmt.Println(green + itallic + "	Done!" + reset)

	// pixels = &newPixels
	return pixels, newImage
}
