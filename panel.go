package main

import (
	"fmt"
	"math"
	"sort"

	"github.com/dghubble/sling"
	"github.com/fogleman/ease"
	colorful "github.com/lucasb-eyer/go-colorful"
	auroraClient "github.com/ngerakines/auroraops/client"
)

const postBrightnessPath = `api/v1/%s/state`

type postbrightness struct {
	Value    int `json:"value"`
	Duration int `json:"duration"`
}

type postBrightnessData struct {
	Brightness postbrightness `json:"brightness"`
}

func postBrightness(url, authToken string, brightness float64) error {
	fmt.Printf("brightness %f\n", brightness)
	req := sling.New().Base(url).Put(fmt.Sprintf(postBrightnessPath, authToken)).BodyJSON(postBrightnessData{Brightness: postbrightness{Value: int(brightness * 100), Duration: 1}})
	r, _ := req.Request()
	fmt.Printf("request: %+v\n", r)
	resp, err := req.Receive(nil, nil)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Resp status %d: %s\n", resp.StatusCode, resp.Status)
	return err
}

type Panel struct {
	Panel           auroraClient.Panel
	source          string
	color           colorful.Color
	brightnessBoost float64
}

func setPanelColor(aurora auroraClient.AuroraClient, panel *Panel) error {
	r, g, b := panel.color.RGB255()
	//fmt.Printf("setting color for panel %d => %d, %d, %d\n", panel.Panel.ID, r, g, b)
	return aurora.SetPanelColor(byte(panel.Panel.ID), r, g, b)
}

func WritePanels(client auroraClient.AuroraClient, url, authToken string, panels ...*Panel) error {
	if len(panels) < 1 {
		return nil
	}

	var brightnesses = make([]float64, len(panels))
	for i, panel := range panels {
		_, _, v := panel.color.Hsv()
		brightnesses[i] = v * panel.brightnessBoost
	}

	sort.Float64s(brightnesses)
	brightness := 0.0
	if bLen := len(brightnesses); bLen%2 == 0 {
		brightness = (brightnesses[bLen/2-1] + brightnesses[bLen/2]) / 2.0
	} else {
		brightness = brightnesses[bLen/2]
	}

	// apply an easing to the brightness to try and get a kinda higher dynamic range
	//fmt.Printf("setting brightness to %f(%f)\n", brightness, ease.InOutCirc(brightness))
	brightness = ease.InOutCirc(brightness)

	err := postBrightness(url, authToken, brightness)
	if err != nil {
		return err
	}

	for _, panel := range panels {
		setPanelColor(client, panel)
	}
	return nil
}

func findPanel(id int, panels ...*Panel) (*Panel, bool) {
	for _, panel := range panels {
		if id == panel.Panel.ID {
			return panel, true
		}
	}

	return nil, false
}

func CalculatePanels(config *Config, timePeriod string, sources map[string][]*Panel) []*Panel {
	panels := []*Panel{}
	var order = []string{"color", "skydome"}
	for _, sourceName := range order {
		sourcePanels := sources[sourceName]
		for _, sourcePanel := range sourcePanels {
			panel, ok := findPanel(sourcePanel.Panel.ID, panels...)
			if ok == false {
				panel = &Panel{Panel: sourcePanel.Panel, source: "mixed", brightnessBoost: 1.0}
				panels = append(panels, panel)
			}

			bias := 0.0
			switch sourcePanel.source {
			case "skydome":
				bias = config.Skydome.Bias[timePeriod]
				panel.brightnessBoost = math.Max(panel.brightnessBoost, config.Skydome.BrightnessBoost*bias)
			case "color":
				bias = config.Color.Bias[timePeriod]
				if timePeriod == "" {
					bias = 1.0
				}
				panel.brightnessBoost = math.Max(panel.brightnessBoost, 1.0*bias)
			}
			//r, g, b := panel.color.RGB255()
			//x, y, z := sourcePanel.color.RGB255()
			panel.color = panel.color.BlendRgb(sourcePanel.color, bias)

			//r1, g1, b1 := panel.color.RGB255()
			//	fmt.Printf("merging (%s) %d,%d,%d => %d,%d,%d @ %f\t\t: %d,%d,%d\n", sourcePanel.source, r, g, b, x, y, z, bias, r1, g1, b1)

		}
	}
	return panels
}
