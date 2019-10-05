package main

import (
	"encoding/binary"
	"hash/crc32"
	"math"
	"time"

	colorful "github.com/lucasb-eyer/go-colorful"
	"github.com/ngerakines/auroraops/client"
)

type SourceSkyDome struct {
	Bias map[string]float64 // maps time-periods to bias values

	BrightnessBoost                              float64
	AzimuthRange, Zenith                         float64
	BoundLeft, BoundRight, BoundTop, BoundBottom *float64
	Haze                                         float64
}

func (source SourceSkyDome) Render(now time.Time, panels ...*client.Panel) ([]*Panel, error) {
	var boundLeft, boundRight, boundTop, boundBottom float64
	for _, panel := range panels {
		if source.BoundLeft != nil {
			boundLeft = *source.BoundLeft
		} else {
			boundLeft = math.Min(boundLeft, float64(panel.X))
		}
		if source.BoundRight != nil {
			boundRight = *source.BoundRight
		} else {
			boundRight = math.Max(boundRight, float64(panel.X))
		}
		if source.BoundTop != nil {
			boundTop = *source.BoundTop
		} else {
			boundTop = math.Max(boundTop, float64(panel.Y))
		}
		if source.BoundBottom != nil {
			boundBottom = *source.BoundBottom
		} else {
			boundBottom = math.Min(boundBottom, float64(panel.Y))
		}
	}

	//fmt.Printf("boundLeft: %f, boundRight: %f, boundTop: %f, boundBottom: %f\n", boundLeft, boundRight, boundTop, boundBottom)

	colorPanels := make([]*Panel, len(panels))
	for i, panel := range panels {
		colorPanel := &Panel{Panel: *panel, source: "skydome"}

		var normX, normY float64
		if boundLeft < 0.0 {
			normX = (math.Abs(boundLeft) + float64(panel.X)) / (boundRight - boundLeft)
		} else {
			normX = (float64(panel.X) - boundLeft) / (boundRight - boundLeft)
		}

		if boundBottom < 0.0 {
			normY = (math.Abs(boundBottom) + float64(panel.Y)) / (boundTop - boundBottom)
		} else {
			normY = (float64(panel.Y) - boundBottom) / (boundTop - boundBottom)
		}
		//fmt.Printf("Panel:%d normalized from %d,%d to %f,%f\n", panel.ID, panel.X, panel.Y, normX, normY)

		r, g, b := getSkyAtPoint(normX, normY, source.Haze, 53.098697, -2.439151, 25.0, now, source.AzimuthRange, source.Zenith)
		//fmt.Printf("for sky at point %fx%f => %f, %f, %f\n--------\n", normX, normY, r, g, b)
		if math.IsNaN(r) {
			r = 0.0
		}
		if math.IsNaN(g) {
			g = 0.0
		}
		if math.IsNaN(b) {
			b = 0.0
		}

		//			fmt.Printf("panelX: %d, panelY: %d === normX: %f, normY: %f\n", panel.X, panel.Y, normX, normY)
		colorPanel.color = colorful.Color{R: r, G: g, B: b}
		colorPanels[i] = colorPanel
	}

	return colorPanels, nil
}

type SourceColor struct {
	Bias               map[string]float64
	Color              string //in hex
	PerturbationBias   float64
	PerturbationLength int64
}

func (source SourceColor) Render(now time.Time, panels ...*client.Panel) ([]*Panel, error) {
	colorPanels := make([]*Panel, len(panels))
	for i, panel := range panels {
		var err error
		colorPanel := &Panel{Panel: *panel, source: "color"}
		colorPanel.color, err = colorful.Hex(source.Color)
		if err != nil {
			return nil, err
		}

		prevBucket := now.Unix() - (now.Unix() % source.PerturbationLength)
		nextBucket := prevBucket + source.PerturbationLength

		enc := []byte{0, 0, 0, 0}
		binary.LittleEndian.PutUint32(enc, uint32(prevBucket+int64(panel.ID)))
		prevPerturbation := float64(crc32.Checksum(enc, crc32.IEEETable)) / float64(math.MaxUint32)

		binary.LittleEndian.PutUint32(enc, uint32(nextBucket+int64(panel.ID)))
		nextPerturbation := float64(crc32.Checksum(enc, crc32.IEEETable)) / float64(math.MaxUint32)

		toNextBucket := 1.0 - (time.Unix(nextBucket, 0).Sub(now).Seconds() / float64(source.PerturbationLength))
		randFloat := prevPerturbation + ((nextPerturbation - prevPerturbation) * toNextBucket)

		perturbation := ((randFloat - 0.5) * 2.0) * source.PerturbationBias
		if perturbation > 0.0 {
			colorPanel.color = colorPanel.color.BlendRgb(colorful.Color{1, 1, 1}, perturbation)
		} else {
			colorPanel.color = colorPanel.color.BlendRgb(colorful.Color{0, 0, 0}, math.Abs(perturbation))
		}
		colorPanels[i] = colorPanel
	}
	return colorPanels, nil
}
