package main

import (
	"net/url"

	"github.com/gorilla/websocket"
	"github.com/pion/webrtc/v2"

	gst "github.com/YXL76/vr-pi-frontend/gstreamer-src"
)

// Rotation Rotation
type Rotation struct {
	Gamma float64
	Alpha float64
	Beta  float64
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

	u := url.URL{Scheme: "ws", Host: "47.96.250.166:8080", Path: "/ws"}

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		panic(err)
	}
	defer c.Close()

	done := make(chan struct{})

	go func() {
		defer close(done)
		for {
			var v webrtc.SessionDescription
			err := c.ReadJSON(&v)
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
			c.WriteJSON(answer)
			gst.CreatePipeline(webrtc.H264, []*webrtc.Track{videoTrack}, "v4l2src").Start()
		}
	}()

	select {}
}
