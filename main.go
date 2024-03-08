package main

import (
	"errors"
	"fmt"
	"net"
	"os"

	"github.com/bluenviron/gortsplib/v4/pkg/format/rtpmjpeg"
	"github.com/pion/rtp"
)

/**
gst-launch-1.0 videotestsrc ! video/x-raw,width=1920,height=1080,format=I420 ! jpegenc ! rtpjpegpay ! udpsink host=127.0.0.1 port=9000
*/

func NewNetworkID(ip string, port int) *net.UDPAddr {

	addr := &net.UDPAddr{
		IP:   net.ParseIP(ip),
		Port: port,
	}

	return addr
}

func CreateUDPListener(netID *net.UDPAddr) (*net.UDPConn, error) {
	listener, err := net.ListenUDP("udp", netID)
	if err != nil {
		return nil, err
	}

	// Increase the receive buffer size
	err = listener.SetReadBuffer(5000000)
	if err != nil {
		fmt.Println("Error setting receive buffer size:", err)
		return nil, err
	}

	return listener, nil
}

func main() {

	netID := NewNetworkID("localhost", 9000)

	listener, err := CreateUDPListener(netID)
	if err != nil {
		panic(err)
	}

	// rtp decoder
	decoder := &rtpmjpeg.Decoder{}
	err = decoder.Init()
	if err != nil {
		panic(err)
	}
	
	inboundRTPPacket := make([]byte, 2048) // UDP MTU
	for {
		n, _, err := listener.ReadFrom(inboundRTPPacket)
		if err != nil {
			fmt.Println("Listener is closed.")
			return
		}

		data := inboundRTPPacket[:n]
		packet := &rtp.Packet{}

		// Decode the byte array into the RTP packet
		err = packet.Unmarshal(data)
		if err != nil {
			panic(err)
		}

		jpegImage, err := decoder.Decode(packet)
		if err != nil {
			if err != rtpmjpeg.ErrNonStartingPacketAndNoPrevious && err != rtpmjpeg.ErrMorePacketsNeeded {
				fmt.Printf("ERR: %v", err)
			}
			if errors.Is(err, rtpmjpeg.ErrMorePacketsNeeded) {
				// need more packets, so continue
				continue
			}
			panic(err)
		}

		file, err := os.OpenFile("output.jpeg", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
		if err != nil {
			// Handle error
			panic(err)
		}

		// Write data to the file
		_, err = file.Write(jpegImage)
		if err != nil {
			// Handle error
			panic(err)
		}

		// File successfully written
		println("Data written to output.jpg")
		os.Exit(0)
	}
}
