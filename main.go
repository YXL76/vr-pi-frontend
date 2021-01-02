package main

import (
	"flag"
	"fmt"
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

func setInterval(s, min, max float64) float64 {
	target := math.Max(s, min)
	target = math.Min(target, max)
	return target
}

func main() {
	// 视频格式，可选 vp8, h264，默认为 vp8
	// 树莓派支持硬解 h264，资源占用小，延迟对，但是安卓端可能会出现无法播放的情况
	// vp8 资源占用较高，不过平台支持性良好
	videoFormat := flag.String("format", "vp8", "GStreamer video format")
	// sdp方案，可选 a, b，默认为 a
	// 通常都是使用 a，对于 Android 原生应用，应使用 b
	sdpSemantics := flag.String("dsp", "a", "SDP Semantics")

	config := webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{
				URLs: []string{"stun:stun.l.google.com:19302"},
			},
		},
	}

	switch *sdpSemantics {
	case "b":
		config = webrtc.Configuration{
			ICEServers: []webrtc.ICEServer{
				{
					URLs: []string{"stun:stun.l.google.com:19302"},
				},
			},
			SDPSemantics: webrtc.SDPSemanticsPlanB,
		}
	}

	// 创建 PeerConnection
	peerConnection, err := webrtc.NewPeerConnection(config)
	if err != nil {
		panic(err)
	}

	var payloadType uint8 = webrtc.DefaultPayloadTypeVP8
	codecName := webrtc.VP8

	switch *videoFormat {
	case "vp9":
		payloadType = webrtc.DefaultPayloadTypeVP8
		codecName = webrtc.VP9
	case "h264":
		payloadType = webrtc.DefaultPayloadTypeH264
		codecName = webrtc.H264
	}

	//
	videoTrack, err := peerConnection.NewTrack(payloadType, 2048, "video", "video")
	if err != nil {
		panic(err)
	}
	_, err = peerConnection.AddTrack(videoTrack)
	if err != nil {
		panic(err)
	}
	gst.CreatePipeline(codecName, []*webrtc.Track{videoTrack}).Start()

	u1 := url.URL{Scheme: "ws", Host: "47.96.250.166:8080", Path: "/webrtc/"}

	// 连接 webrtc session
	c1, _, err := websocket.DefaultDialer.Dial(u1.String(), nil)
	if err != nil {
		panic(err)
	}
	println("Webrtc connection establish")
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
			fmt.Printf("%+v\n", v)
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
			err = c1.WriteJSON(answer)
		}
	}()

	device, err := pca9685.Open()
	if err != nil {
		panic(err)
	}
	defer device.Close()

	// 初始化 PCA9685
	device.SetFrequency(50.0)
	defer device.SetPulse(0, 0)
	defer device.SetPulse(1, 0)

	u2 := url.URL{Scheme: "ws", Host: "47.96.250.166:8080", Path: "/sensor/"}

	// 连接 sensor session
	c2, _, err := websocket.DefaultDialer.Dial(u2.String(), nil)
	if err != nil {
		panic(err)
	}
	println("Sensor connection establish")
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

			// 处理传感器数据
			if v.Gamma > 0 {
				alpha = setInterval(v.Alpha, 90, 300)
				alpha -= 270
			} else {
				if v.Alpha > 195 {
					alpha = setInterval(v.Alpha, 270, 360)
					alpha -= 270
				} else {
					alpha = setInterval(v.Alpha, 0, 120)
					alpha += 90
				}
			}

			if v.Gamma > 45 {
				gamma = -v.Gamma + 225
			} else {
				gamma = -v.Gamma + 45
			}

			device.SetPulse(0, gamma*10+500) // 大致范围为600～2200
			device.SetPulse(1, alpha*10+500) // 大致范围为600～2500

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
