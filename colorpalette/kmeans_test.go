package colorpalette

import (
	"encoding/json"
	"fmt"
	"image"
	"io/ioutil"
	"strings"
	"testing"
	"time"

	"github.com/mielpeeters/dither/imgutil"
	"github.com/mielpeeters/dither/process"
)

var img image.Image
var scale int
var k int
var Max int = 20
var Step int = 4
var scaleMax = 20
var scaleStep = 4
var results []*Result

type Result struct {
	Seconds float64
	Scale   int
	K       int
}

func testInit(t *testing.T) {
	imgOrig, err := imgutil.OpenImage("../data/sample-image.jpg")
	if err != nil {
		fmt.Println(err)
		t.Errorf("couldn't open image")
	}

	img = process.Downscale(imgOrig, scale)
}

// TestKMSpeed tests the speed of the KM algorithm for a range of parameters
func TestCreateSpeed(t *testing.T) {
	fmt.Printf("\n\n\033[1mStart Create Speed Test\033[0m\n\n")

	totalRuns := ((scaleMax-1)/scaleStep + 1) * ((Max-1)/Step + 1)
	run := 0

	KMAccuracy = 0.01
	SampleFactor = 4

	for scale = 1; scale <= scaleMax; scale += scaleStep {
		for k = 1; k < Max; k += Step {
			run++
			testInit(t)
			start := time.Now()
			Create(img, k)
			duration := time.Since(start)

			results = append(results, &Result{
				Seconds: duration.Seconds(),
				Scale:   scale,
				K:       k,
			})

			fmt.Printf("\r\033[1mProgress: \033[32m%s\033[31m%s \033[0m(%d / %d)", strings.Repeat("―", run), strings.Repeat("―", totalRuns-run), run, totalRuns)
		}
	}

	output, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		fmt.Println(err)
		return
	}

	err = ioutil.WriteFile("results2.json", output, 0644)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Printf("\n\033[1m\033[32mDONE\033[0m\n\n")
}
