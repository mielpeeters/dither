// Package gifeo can be used to create gif videos, from a premade set of images
//
// ffmpeg can be used to create that set:
// `ffmpeg -i <input video> frames/frame_%05d.jpg`
package gifeo

import (
	"image"
	"image/color"
	"image/gif"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"sync"

	"github.com/mielpeeters/dither/colorpalette"
	"github.com/mielpeeters/dither/imgutil"
	"github.com/mielpeeters/dither/needle"
	"github.com/mielpeeters/dither/process"
	"github.com/mielpeeters/pacebar"
)

// Verbosity can be used to controll the cli output of the program
// 0 -> nothing
// 1 -> progress bar
var Verbosity = 1

// Giffer is a struct that contains setup information and is used
// to create gif videos
type Giffer struct {
	// Scale is the scaledown factor used in creating
	// the pixelated dither effect, on a per-frame basis
	Scale int
	// K is the amount of colors to be used in the palette
	K int
	// Palette can be set by the user, if left at default nil,
	// gifeo will create the palette from the first frame
	Palette color.Palette

	mu     sync.Mutex
	pb     pacebar.Pacebar
	frames []*image.Paletted
}

// CreateVideo is used to create the gif video
// The frames in the inputDir directory need to be of format: frame_ddddd.jpg.
// This can be achieved with ffmpeg by specifying as an output: frame_%05d.jpg
// That does mean that the maximum GIF length is 6min40s
func (gf *Giffer) CreateVideo(inputDir, outputFile string) {

	pattern := "frame_[0-9]{5}\\.jpg"

	re, err := regexp.Compile(pattern)
	if err != nil {
		panic(err)
	}

	frame := 0

	// paths maps frame numbers to their paths
	paths := make(map[int]string, 0)

	// add an entry that maps frame count to the path string per matching path
	filepath.Walk(inputDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && re.MatchString(info.Name()) {
			paths[frame] = path
			frame++
		}
		return nil
	})

	// create the pacebar if verbosity is set
	if Verbosity > 0 {
		gf.pb = pacebar.Pacebar{Work: len(paths)}
	}

	// frames keeps the processed frames in a slice
	gf.frames = make([]*image.Paletted, len(paths))

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
				gf.handleFrame((paths)[j], j)
			}
			wg.Done()
		}(&frameNumbers[i])
	}

	// wait for all child threads to finish
	wg.Wait()

	EncodeGIF(gf.frames, outputFile, 4)
}

// EncodeGIF encodes a slice of image.Paletted images with a given palette and
// saves it into the outputFile path.
func EncodeGIF(frames []*image.Paletted, outputFile string, delay int) {
	// everything from here down is encoding & saving the gif
	delays := make([]int, len(frames))
	for i := range delays {
		delays[i] = delay
	}

	// frame 0 used for config
	frame0 := *frames[0]

	g := gif.GIF{
		Image: frames,
		Delay: delays,

		// By specifying a Config, we can set a global color table for the GIF.
		// This is more efficient then each frame having its own color table, which
		// is the default when there's no config.
		Config: image.Config{
			ColorModel: frame0.Palette,
			Width:      frame0.Rect.Dx(),
			Height:     frame0.Rect.Dy(),
		},
	}

	file, err := os.Create(outputFile)
	if err != nil {
		panic(err)
	}

	err = gif.EncodeAll(file, &g)
	if err != nil {
		panic(err)
	}
}

func (gf *Giffer) handleFrame(path string, frameNo int) {
	// open the input image
	img, err := imgutil.OpenImage(path)
	if err != nil {
		return
	}

	// scale the image down with a given scale
	scaledImage := process.Downscale(img, gf.Scale)

	if gf.Palette == nil {
		gf.mu.Lock() // only one process gets through when gf.Palette is still nill
		if gf.Palette == nil {
			gf.Palette = colorpalette.Create(scaledImage, gf.K)
		}
		gf.mu.Unlock()
	}

	paletted := process.ApplyErrorDiffusion(scaledImage, gf.Palette, &process.JarvisJudiceNinke)

	gf.frames[frameNo] = paletted

	if Verbosity > 0 {
		gf.pb.Done(1)
	}
}
