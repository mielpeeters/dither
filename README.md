# dither
Implements the Error Diffusion Dithering algorithm, and some other image manipulation techniques.
It is a work in progress, more effects are in the work.
This go module contains multiple packages:
- **process**: the core functionality, implementing the image algorithms (like FSD).
- **geom**: some simple geometry types (like `Point`) and functions to manipulate them.
- **imgutil**: some image utilities, like `OpenImage`.
- **kmeans**: k-means clustering implementation, useful for finding the k most prominent colors in an image.
- **nearneigh**: a work in progress, impelementation of a nearest neighbour search algorithm.
- **kdtree**: a work in progress, implements a kd tree search structure for fast nearest neighbour.
- **colorpalette**: a custom defined colorpalette type, with accompanying functions.
- **gifeo**: a package for creating dithered gif videos
- **needle**: some functions that are useful for multithreading (the needle)

<div style="display: flex; flex-direction: row; justify-content:space-evenly;"><img src="https://user-images.githubusercontent.com/72082402/225594701-a15c3d26-5ad9-4d42-9d25-cdc7751c8ad2.png" alt="example image created using the dither module." height="400">
<img src="https://user-images.githubusercontent.com/72082402/227805266-be47ad7d-c4d4-47cd-9cec-d24196aa07b9.gif" alt="example gif video created using the dither module." height="400"></div>

## Usage
In your go module (after having run `go mod init mymodule`), get this module by running `go get github.com/mielpeeters/dither`.
You can then import the various packages in your code.

As an example, here is code you could write to apply the Floyd Steinberg Dithering algorithm:
```go
package main

import (
	"github.com/mielpeeters/dither/colorpalette"
	"github.com/mielpeeters/dither/imgutil"
	"github.com/mielpeeters/dither/process"
)

func main() {
    // open the input image
    img, err := imgutil.OpenImage("path/to/imputImage.any")
    if err != nil {
    return
    }

    // scale the image down with a given scale
    scale := 10
    scaledImage := process.Downscale(img, scale)

    // get the palette in which to create the new image
    amountOfColors := 5
    colorpalette.SampleFactor = 2
	colorpalette.KMTimes = 4
    palette = colorpalette.Create(pixels, amountOfColors)

    // apply dithering to the image
    paletted := process.ApplyErrorDiffusion(pixels, palette, &process.FloydSteinBerg)

    // save the image as a GIF (efficient for paletted images)
    imgutil.SaveGIF(paletted, "path/to/outputImage.gif")
}
``` 

## License
This module is licensed under version 3 of the GNU General Public License.
