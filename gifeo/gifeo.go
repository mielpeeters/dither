// Package gifeo can be used to create gif videos, from a premade set of images
//
// ffmpeg can be used to create that set:
// `ffmpeg -i <input video> frames/frame_%05d.jpg`
package gifeo

import (
	"fmt"
	"image"
	"image/color"
	"image/gif"
	"os"
	"path/filepath"
	"regexp"

	"github.com/mielpeeters/dither/colorpalette"
	"github.com/mielpeeters/dither/imgutil"
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
	// Frames is used to indicate the amount of frames the input,
	// and thus output video have
	Frames int
	// Palette can be set by the user, if left at default nil,
	// gifeo will create the palette from the first frame
	Palette color.Palette

	pb           pacebar.Pacebar
	first        bool
	frame        int
	ditherMatrix process.ErrorDiffusionMatrix
}

// CreateVideo is used to create the gif video
func (gf *Giffer) CreateVideo(inputDir, outputFile string) {
	if Verbosity > 0 {
		gf.pb = pacebar.Pacebar{Work: gf.Frames}
	}

	pattern := "frame_[0-9]{5}\\.jpg"

	re, err := regexp.Compile(pattern)
	if err != nil {
		panic(err)
	}

	frames := make([]*image.Paletted, 0)

	gf.first = true
	gf.frame = 0

	filepath.Walk(inputDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && re.MatchString(info.Name()) {
			gf.handleFrame(path, &frames)
		}
		return nil
	})

	delays := make([]int, len(frames))
	for i := range delays {
		delays[i] = 4
	}

	g := gif.GIF{
		Image: frames,
		Delay: delays,

		// By specifying a Config, we can set a global color table for the GIF.
		// This is more efficient then each frame having its own color table, which
		// is the default when there's no config.
		Config: image.Config{
			ColorModel: gf.Palette,
			Width:      (*frames[0]).Rect.Dx(),
			Height:     (*frames[0]).Rect.Dy(),
		},
	}

	f2, err := os.Create(outputFile)
	if err != nil {
		panic(err)
	}

	err = gif.EncodeAll(f2, &g)
	if err != nil {
		panic(err)
	}

	fmt.Println("Gifeo: GIF video saved.")
}

func (gf *Giffer) handleFrame(path string, frames *[]*image.Paletted) {
	// open the input image
	img, err := imgutil.OpenImage(path)
	if err != nil {
		return
	}

	// scale the image down with a given scale
	scaledImage := process.Downscale(img, gf.Scale)

	if gf.first {
		if gf.Palette == nil {
			gf.Palette = colorpalette.Create(scaledImage, gf.K)
		}
		gf.ditherMatrix = process.JarvisJudiceNinke
		gf.first = false
	}

	paletted := process.ApplyErrorDiffusion(scaledImage, gf.Palette, &gf.ditherMatrix)

	*frames = append(*frames, paletted)

	gf.frame++
	gf.pb.Done(1)
}
