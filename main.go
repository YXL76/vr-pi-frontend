package main

import (
	"math"
	"net/url"

	"github.com/YXL76/vrpi-pi/pca9685"
	"github.com/gorilla/websocket"
	"github.com/pion/webrtc/v2"

	gst "github.com/YXL76/vrpi-pi/gstreamer-src"
)

// Rotation Rotation
type Rotation struct {
	Alpha float64
	// Beta  float64
	Gamma float64
}

func main() {
	config := webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{
				URLs: []string{"stun:stun.l.google.com:19302"},
			},
		},
		SDPSemantics: webrtc.SDPSemanticsPlanB,
	}

	peerConnection, err := webrtc.NewPeerConnection(config)
	if err != nil {
		panic(err)
	}

	videoTrack, err := peerConnection.NewTrack(webrtc.DefaultPayloadTypeH264, 2048, "video", "video")
	if err != nil {
		panic(err)
	}
	_, err = peerConnection.AddTrack(videoTrack)
	if err != nil {
		panic(err)
	}
	gst.CreatePipeline([]*webrtc.Track{videoTrack}).Start()

	u1 := url.URL{Scheme: "ws", Host: "47.96.250.166:8080", Path: "/webrtc/"}

	c1, _, err := websocket.DefaultDialer.Dial(u1.String(), nil)
	if err != nil {
		panic(err)
	}
	defer c1.Close()

	done1 := make(chan struct{})

	go func() {
		defer close(done1)
		var v webrtc.SessionDescription
		for {
			err := c1.ReadJSON(&v)
			if err != nil {
				panic(err)
			}
			err = peerConnection.SetRemoteDescription(v)
			if err != nil {
				panic(err)
			}
			answer, err := peerConnection.CreateAnswer(nil)
			if err != nil {
				panic(err)
			}
			err = peerConnection.SetLocalDescription(answer)
			if err != nil {
				panic(err)
			}
			c1.WriteJSON(answer)
		}
	}()

	device, err := pca9685.Open()
	if err != nil {
		panic(err)
	}
	defer device.Close()

	device.SetFrequency(50.0)

	u2 := url.URL{Scheme: "ws", Host: "47.96.250.166:8080", Path: "/sensor/"}

	c2, _, err := websocket.DefaultDialer.Dial(u2.String(), nil)
	if err != nil {
		panic(err)
	}
	defer c2.Close()

	done2 := make(chan struct{})

	go func() {
		defer close(done2)
		/* a := 5.71559214e-05
		b := 3.60082305e-02
		c := 2.50000000 */

		var v Rotation
		var alpha float64
		var gamma float64
		for {
			err := c2.ReadJSON(&v)
			if err != nil {
				panic(err)
			}
			gamma = v.Gamma + 40 // 0 ~ 130,-50 ~ 0
			if gamma < 0 {
				gamma += 180 // 0 ~ 180
				if v.Alpha < 90 {
					// 0 ~ 90
					alpha = v.Alpha + 110 // 110 ~ 200
				} else {
					// 250 ~ 360
					alpha = v.Alpha - 250 // 0 ~ 110
				}
			} else {
				alpha = v.Alpha - 70 // 0 ~ 200
			}

			gamma = math.Max(gamma, 0)
			gamma = math.Min(gamma, 180)

			alpha = math.Max(alpha, 0)
			alpha = math.Min(alpha, 200)

			device.SetPulse(0, gamma*9+590)
			device.SetPulse(1, alpha*10+400)

			/* verticalDirection := v.Gamma * -180.0 / math.Pi
			levelDirection := 180.0 - (v.Alpha * -180.0 / math.Pi)

			verticalDuty := a*verticalDirection*verticalDirection + b*verticalDirection + c
			levelDuty := a*levelDirection*levelDirection + b*levelDirection + c

			device.SetPulse(0, verticalDuty)
			device.SetPulse(1, levelDuty)*/
		}
	}()

	select {}
}
