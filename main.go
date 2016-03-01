package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/freb/go-blink1"
)

const (
	fade = 250 * time.Millisecond
	dur  = 2 * time.Second
)

var white = blink1.State{Red: 255, Green: 255, Blue: 255, FadeTime: fade, Duration: dur} // #FFFFFF
var black = blink1.State{Red: 0, Green: 0, Blue: 0, FadeTime: fade, Duration: dur}       // #000000
var red = blink1.State{Red: 255, Green: 0, Blue: 0, FadeTime: fade, Duration: dur}       // #FF0000
var orange = blink1.State{Red: 255, Green: 165, Blue: 0, FadeTime: fade, Duration: dur}  // #FFA500
var yellow = blink1.State{Red: 255, Green: 255, Blue: 0, FadeTime: fade, Duration: dur}  // #FFFF00
var purple = blink1.State{Red: 128, Green: 0, Blue: 128, FadeTime: fade, Duration: dur}  // #800080
var blue = blink1.State{Red: 0, Green: 0, Blue: 255, FadeTime: fade, Duration: dur}      // #0000FF
var green = blink1.State{Red: 0, Green: 128, Blue: 0, FadeTime: fade, Duration: dur}     // #008000

type status struct {
	sync.Mutex
	ifaceUp bool // red
	gwUp    bool // yellow
	tunUp   bool // purple
	inetUp  bool // blue
	dev     *blink1.Device
}

func (s *status) closeDev() {
	s.Lock()
	defer s.Unlock()
	s.dev.Close()
	var dev *blink1.Device
	s.dev = dev
}

func (s *status) setIface(up bool) {
	s.Lock()
	defer s.Unlock()
	s.ifaceUp = up
}

func (s *status) getIface() bool {
	s.Lock()
	defer s.Unlock()
	return s.ifaceUp
}

func (s *status) setGw(up bool) {
	s.Lock()
	defer s.Unlock()
	s.gwUp = up
}
func (s *status) getGw() bool {
	s.Lock()
	defer s.Unlock()
	return s.gwUp
}

func (s *status) setTun(up bool) {
	s.Lock()
	defer s.Unlock()
	s.tunUp = up
}
func (s *status) getTun() bool {
	s.Lock()
	defer s.Unlock()
	return s.tunUp
}

func (s *status) setInet(up bool) {
	s.Lock()
	defer s.Unlock()
	s.inetUp = up
}
func (s *status) getInet() bool {
	s.Lock()
	defer s.Unlock()
	return s.inetUp
}

func (s *status) Pattern() blink1.Pattern {
	s.Lock()
	defer s.Unlock()
	states := make([]blink1.State, 0)
	if s.ifaceUp == false {
		states = append(states, red)
	}
	if s.gwUp == false {
		states = append(states, yellow)
	}
	if s.tunUp == false {
		states = append(states, purple)
	}
	if s.inetUp == false {
		states = append(states, blue)
	}
	if s.ifaceUp == true && s.gwUp == true && s.tunUp == true && s.inetUp == true {
		states = append(states, green)
	}
	states = append(states, black)
	return blink1.Pattern{States: states}
}

func ifaceUp(iface string) bool {
	path := "/sys/class/net/" + iface + "/operstate"
	dat, err := ioutil.ReadFile(path)
	if err != nil {
		fmt.Println(err)
	}
	state := string(dat)
	state = strings.TrimSpace(state)
	return state == "up"
}

func gwUp(gw string) bool {
	if gw == "" {
		out, err := exec.Command("ip", "route").Output()
		if err != nil {
			fmt.Println(err)
		}
		routes := string(out)

		for _, v := range strings.Split(routes, "\n") {
			if strings.HasPrefix(v, "default") {
				gw = strings.Fields(v)[2]
			}
		}
	}

	if err := exec.Command("ping", "-c", "1", gw).Run(); err != nil {
		return false
	}
	return true
}

func tunUp(tun1, tun2 string) bool {
	if tun1 == "" && tun2 == "" {
		return true
	}

	if tun1 != "" {
		if err := exec.Command("ping", "-c", "1", tun1).Run(); err == nil {
			return true
		}
	}
	if tun2 != "" {
		if err := exec.Command("ping", "-c", "1", tun2).Run(); err == nil {
			return true
		}
	}
	return false
}

func inetUp(ip string) bool {
	if err := exec.Command("ping", "-c", "1", ip).Run(); err != nil {
		return false
	}
	return true
}

func (s *status) write() {

	defer func() {
		if r := recover(); r != nil {
			s.closeDev()
		}
	}()

	if s.dev == nil {
		// library bug requires blink1-tool to first make contact with device
		exec.Command("blink1-tool", "--list").Run()
		dev, _ := blink1.OpenNextDevice()
		if dev == nil {
			// looping terminates here when blink1 is unplugged
			time.Sleep(time.Second)
			return
		}
		s.Lock()
		s.dev = dev
		s.Unlock()
	}

	p := s.Pattern()
	start := time.Now()
	s.dev.RunPattern(&p)
	// only reliable way to tell if the blink1 device is open is to see if pattern
	// takes as long to execute as it should
	if time.Now().Sub(start) < time.Second {
		// this is only hit once per unplug
		s.closeDev()
	}

	return
}

func main() {
	var (
		iface string
		gw    string
		tun1  string
		tun2  string
		inet  string
	)

	flag.StringVar(&iface, "iface", "eth0", "network interface")
	flag.StringVar(&gw, "gw", "", "default gateway")
	flag.StringVar(&tun1, "tun1", "", "vpn tunnel endpoint 1")
	flag.StringVar(&tun2, "tun2", "", "vpn tunnel endpoint 2")
	flag.StringVar(&inet, "inet", "8.8.8.8", "internet endpoint")

	flag.Parse()

	s := &status{}

	go func() {
		for {
			up := ifaceUp(iface)
			s.setIface(up)
			time.Sleep(5 * time.Second)
		}
	}()

	go func() {
		for {
			up := gwUp(gw)
			s.setGw(up)
			time.Sleep(5 * time.Second)
		}
	}()

	go func() {
		for {
			up := tunUp(tun1, tun2)
			s.setTun(up)
			time.Sleep(5 * time.Second)
		}
	}()

	go func() {
		for {
			up := inetUp(inet)
			s.setInet(up)
			time.Sleep(5 * time.Second)
		}
	}()

	for {
		s.write()
	}
}
