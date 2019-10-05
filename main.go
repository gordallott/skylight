package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/muja/suncalc-go"
	"github.com/ngerakines/auroraops/client"
	toml "github.com/pelletier/go-toml"
)

var cacheDir string
var confDir string
var authFileLocation string
var configFileLocation string

//type Auth struct {
//	Name string `json:"name" toml:"name"` // Supporting both JSON and toml.
//	Age  int    `json:"name" toml:"name"`
//}
type Auth struct {
	URL string
	Key string
}

func init() {
	var err error
	cacheDir, err = os.UserCacheDir()
	if err != nil {
		cacheDir = "/tmp/"
	}
	cacheDir = filepath.Join(cacheDir, "skylight")

	if v, ok := os.LookupEnv("CONFDIR"); ok {
		fmt.Println("docker override")
		// docker running override, stuff everything in confdir
		confDir = v
		cacheDir = filepath.Join(confDir, ".cache")
	} else if v := os.Getenv("HOME"); v != "" {
		confDir = filepath.Join(v, ".config", "skylight")
	}

	authFileLocation = filepath.Join(cacheDir, "auth.toml")
	configFileLocation = filepath.Join(confDir, "config.toml")

	fmt.Printf("authFileLocation: %s\nconfigFileLocation: %s\n", authFileLocation, configFileLocation)
}

func getClient(token string) (client.AuroraClient, string, error) {
	hosts, err := client.Disocver(20 * time.Second)
	if err != nil {
		return nil, "", err
	}

	if len(hosts) < 1 {
		return nil, "", errors.New("no clients found")
	}

	if token == "" {
		client, err := client.New(hosts[0])
		return client, hosts[0], err
	}

	client, err := client.NewWithToken(hosts[0], token)
	return client, hosts[0], err
}

func getAuth() (Auth, error) {
	fmt.Printf("getting auth from: %s\n", authFileLocation)
	authf, err := os.Open(authFileLocation)
	if err != nil {
		return Auth{}, err
	}

	authr, err := ioutil.ReadAll(authf)
	if err != nil {
		return Auth{}, err
	}

	auth := Auth{}
	if err := toml.Unmarshal(authr, &auth); err != nil {
		return Auth{}, err
	}

	return auth, nil
}

func getNewAuth() error {
	fmt.Println("getting auth...")
	client, host, err := getClient("")
	if err != nil {
		//log.WithError(err).Error("Could not discover aurora.")
		fmt.Fprintf(os.Stderr, "Could not discover aurora: %s")
		os.Exit(1)
	}

	fmt.Println("Need to authenticate to device. Hold power button to enter pairing mode. Press enter when ready.")
	reader := bufio.NewReader(os.Stdin)
	reader.ReadString('\n')
	token, err := client.Authorize()
	if err != nil {
		//log.WithError(err).Error("Unable to authenticate")
		fmt.Fprintf(os.Stderr, "unable to authenticate: %s", err)
		os.Exit(1)
	}

	data := Auth{
		Key: token,
		URL: host,
	}

	d, err := toml.Marshal(&data)
	if err != nil {
		//	log.WithFields(log.Fields{
		//		"token": token,
		//		"host":  hosts[0],
		//	}).WithError(err).Error("Unable to compose toml")
		fmt.Fprintf(os.Stderr, "unable to compose toml, %s", err)
		os.Exit(1)
	}

	os.MkdirAll(filepath.Dir(authFileLocation), 0777)
	if err := ioutil.WriteFile(authFileLocation, d, 0645); err != nil {
		//log.WithFields(log.Fields{
		//	"token": token,
		//	"host":  hosts[0],
		//}).WithError(err).Error("Unable to compose toml")
		fmt.Fprintf(os.Stderr, "unable to compose toml: %s", err)
		os.Exit(1)
	}
	fmt.Printf("wrote auth file to %s\n", authFileLocation)
	return nil
}

func main() {
	config, err := getConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "could not open config at %s: %s", configFileLocation, err)
		os.Exit(1)
	}

	fmt.Printf("using config: %+v\n", config)

	fmt.Printf("uh, get auth?\n")
	auth, err := getAuth()
	fmt.Printf("auth=%+v, err=%s\n", auth, err)
	if err != nil {
		fmt.Printf("no auth?")
		fmt.Printf("warning could not open auth file: %s\n", err)
		for err != nil {
			err = getNewAuth()
		}
	}

	client, host, err := getClient(auth.Key)
	if err != nil {
		fmt.Fprintf(os.Stderr, "could not connect to %s: %s", host, err)
		os.Exit(1)
	}

	fmt.Printf("connected to %s...\n", host)
	hardwareInfo, err := client.GetInfo()
	if err != nil {
		fmt.Fprintf(os.Stderr, "could not get hardware info: %s", err)
		os.Exit(1)
	}

	// remap panels that are weird
	for _, panel := range hardwareInfo.Panels {
		id := strconv.Itoa(panel.ID)
		xy, ok := config.RemapPanels[id]
		if ok {
			fmt.Printf("remapping panel %d, %dx%d => %dx%d", panel.ID, panel.X, panel.Y, xy.X, xy.Y)
			panel.X = xy.X
			panel.Y = xy.Y
		}
	}

	// need to flip the panel axis because this is life
	if config.FlipPanels {
		for _, panel := range hardwareInfo.Panels {
			skip := false
			for _, exclude := range config.FlipExcludes {
				if panel.ID == exclude {
					skip = true
				}
			}
			if skip {
				continue
			}

			x := panel.X
			y := panel.Y
			panel.X = -y
			panel.Y = -x
		}
	}

	tmp, _ := json.MarshalIndent(hardwareInfo, "", "  ")
	fmt.Printf("hardware info: %s\n", tmp)

	fmt.Printf("currentSun Coordinates, Azimuth=%f, ZenithAngle=%f\n", config.GetSunCoordinates().Azimuth, config.GetSunCoordinates().ZenithAngle)

	ticker := time.NewTicker(config.RefreshRate)

	var accel time.Duration

	rand.Seed(time.Now().Unix())
	lastSecond := time.Now().Unix()
	for range ticker.C {
		accel = accel + config.TimeAccel
		now := time.Now().Add(accel)
		hardwarePanels := hardwareInfo.Panels

		sunTimes := suncalc.SunTimes(now, config.Location.Latitude, config.Location.Longitude)
		var sunTimeName string
		var sunTime time.Time
		for timeName, timeTime := range sunTimes {
			if timeTime.Before(now) && timeTime.After(sunTime) {
				sunTimeName = timeName
				sunTime = timeTime
			}
		}

		if lastSecond < now.Unix() {
			lastSecond = now.Unix()
			fmt.Printf("calculating panel colors for %s(started %s) [%s]...\n", sunTimeName, sunTime, now)
		}

		sourcePanels := map[string][]*Panel{}
		panels, err := config.Skydome.Render(now, hardwarePanels...)

		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %s", err)
			ticker.Stop()
			os.Exit(1)
		}

		sourcePanels[`skydome`] = panels

		panels, err = config.Color.Render(now, hardwarePanels...)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %s", err)
			ticker.Stop()
			os.Exit(1)
		}

		sourcePanels[`color`] = panels

		panels = CalculatePanels(config, sunTimeName, sourcePanels)
		if err := WritePanels(client, auth.URL, auth.Key, panels...); err != nil {
			fmt.Fprintf(os.Stderr, "error: %s", err)
			ticker.Stop()
			os.Exit(1)
		}
	}
}
