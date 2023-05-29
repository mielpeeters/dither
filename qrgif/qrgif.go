package qrgif

import (
	"image"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"sync"
	"time"

	"github.com/mielpeeters/dither/colorpalette"
	"github.com/mielpeeters/dither/gifeo"
	"github.com/mielpeeters/dither/imgutil"
	"github.com/mielpeeters/dither/needle"
	"github.com/mielpeeters/dither/process"
	"github.com/skip2/go-qrcode"
)

// QRGif represents the qr gif video
type QRGif struct {
	// VideoPath is the directory of the frames, stored in format frame_%05d.jpg
	VideoPath string

	OutputPath string

	Code *qrcode.QRCode

	ChangeFraction float64

	frames []*image.Paletted

	codeimg *image.Paletted
}

// NewQRGif creates a new QRGif object
func NewQRGif(videoPath, outputPath, content string, changeFraction float64) *QRGif {
	code, err := qrcode.NewWithForcedVersion(content, 6, qrcode.Highest)
	if err != nil {
		log.Fatal(err)
	}

	code.WriteFile(49, "tmp.png")

	img, err := imgutil.OpenImage("tmp.png")
	if err != nil {
		log.Fatal(err)
	}

	rgbaImg := process.Resize(img, 49, 49)

	paletted := process.ApplyErrorDiffusion(rgbaImg, colorpalette.BW(), &process.JarvisJudiceNinke)

	rand.Seed(time.Now().UnixNano())

	return &QRGif{
		VideoPath:      videoPath,
		Code:           code,
		codeimg:        paletted,
		OutputPath:     outputPath,
		ChangeFraction: changeFraction,
	}
}

// EmbedVideo embeds the Video into the QRCode
func (qrg *QRGif) EmbedVideo() {

	pattern := "frame_[0-9]{5}\\.jpg"

	re, err := regexp.Compile(pattern)
	if err != nil {
		panic(err)
	}

	frame := 0

	// paths maps frame numbers to their paths
	paths := make(map[int]string, 0)

	// add an entry that maps frame count to the path string per matching path
	filepath.Walk(qrg.VideoPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && re.MatchString(info.Name()) {
			paths[frame] = path
			frame++
		}
		return nil
	})

	// frames keeps the processed frames in a slice
	qrg.frames = make([]*image.Paletted, len(paths))

	// make a slice of the keys of the paths map (frame numbers)
	// will be used to spread the multithreaded load
	keys := make([]int, len(paths))
	i := 0
	for key := range paths {
		keys[i] = key
		i++
	}

	// divide the frameNumbers in chunks, each to be dealth with by one thread
	frameNumbers := needle.ChunkSlice(keys, runtime.GOMAXPROCS(0))

	// start multithreaded processing of frames
	wg := sync.WaitGroup{}

	for i := range frameNumbers {
		wg.Add(1)
		go func(myFrameNumbers *[]int) {
			for _, j := range *myFrameNumbers {
				// here, all of the frames that are my responsibility will be dealth with
				qrg.handleFrame((paths)[j], j)
			}
			wg.Done()
		}(&frameNumbers[i])
	}

	// wait for all child threads to finish
	wg.Wait()

	gifeo.EncodeGIF(qrg.frames, qrg.OutputPath, 8)
}

func (qrg *QRGif) handleFrame(inputPath string, no int) {
	// open the input image
	img, err := imgutil.OpenImage(inputPath)
	if err != nil {
		return
	}
	// scale the image down with a given scale
	scaledImage := process.Resize(img, 49, 49)

	paletted := process.ApplyErrorDiffusion(scaledImage, colorpalette.BW(), &process.Nothing)

	if no == 40 {
		imgutil.SaveGIF(paletted, "TEST.gif")
	}

	adjusted := 0

	// apply QR code filter on top
	for x := 0; x < 49; x++ {
		for y := 0; y < 49; y++ {
			imagePixel := false

			if !mask(x-4, y-4) {
				if paletted.ColorIndexAt(x, y) != qrg.codeimg.ColorIndexAt(x, y) && paletted.ColorIndexAt(x, y) != 1 {
					if rand.Float64() < qrg.ChangeFraction && adjusted < int(qrg.ChangeFraction*41*30) {
						adjusted++
						imagePixel = true
						// qrg.ChangeFraction = orig * 5
					}
				}
			}

			if !imagePixel {
				paletted.SetColorIndex(x, y, qrg.codeimg.ColorIndexAt(x, y))
			}
		}
	}

	qrg.frames[no] = paletted
}

func mask(x, y int) bool {
	// Quiet Zone
	if x < 0 || y < 0 || x > 40 || y > 40 {
		return true
	}
	// check Timing Patterns
	if x == 6 {
		return true
	}
	if y == 6 {
		return true
	}

	// check Position Detection Patterns
	// left top
	if x < 8 && y < 8 {
		return true
	}
	// right top
	if x >= 33 && y < 8 {
		return true
	}
	// left bottom
	if x < 8 && y >= 33 {
		return true
	}

	// check Alignment Pattern
	if x >= 32 && x < 37 && y >= 32 && y < 37 {
		return true
	}

	// check Format Information
	if x == 8 {
		if y < 9 {
			return true
		}
		if y >= 33 {
			return true
		}
	}
	if y == 8 {
		if x < 9 {
			return true
		}
		if x >= 33 {
			return true
		}
	}

	// if here, x, y is not part of the patterns
	return false
}
