package main

import (
	"github.com/pion/webrtc/v2"

	gst "github.com/YXL76/vr-pi-frontend/gstreamer-src"
)

func main() {
	config := webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{
				URLs: []string{"stun:stun.l.google.com:19302"},
			},
		},
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

	gst.CreatePipeline(webrtc.H264, []*webrtc.Track{videoTrack}, "v4l2src -d /dev/video0").Start()

	// Block forever
	select {}
}
