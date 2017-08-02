package main

import (
	"bytes"
	"encoding/binary"
	// "encoding/hex"
	"fmt"
	"net"
	"os"
	"time"
)

const (
	CONN_HOST = "0.0.0.0"
	CONN_PORT = "4554"
	CONN_TYPE = "tcp"
)

type GPSElement struct {
	latitude  int32
	longitude int32
	angle     int16
	speed     int16
}

func main() {
	// Listen for incoming connections.
	l, err := net.Listen(CONN_TYPE, CONN_HOST+":"+CONN_PORT)
	if err != nil {
		fmt.Println("Error listening:", err.Error())
		os.Exit(1)
	}

	// Close the listener when the application closes.
	defer l.Close()

	fmt.Println("Listening on " + CONN_HOST + ":" + CONN_PORT)
	for {
		// Listen for an incoming connection.
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting: ", err.Error())
			os.Exit(1)
		}
		// Handle connections in a new goroutine.
		go handleRequest(conn)
	}
}

func handleRequest(conn net.Conn) {
	var b []byte
	knownIMEI := true
	step := 1

	// Close the connection when you're done with it.
	defer conn.Close()

	for {
		// Make a buffer to hold incoming data.
		buf := make([]byte, 1024)

		// Read the incoming connection into the buffer.
		size, err := conn.Read(buf)
		if err != nil {
			fmt.Println("Error reading:", err.Error())
			break
		}

		// Send a response if known IMEI and matches IMEI size
		if knownIMEI {
			b = []byte{1}

			fmt.Println("----------------------------------------")
			fmt.Println("Data From:", conn.RemoteAddr().String())
			// fmt.Println("Size of message: ", size)
			// fmt.Println("Message:", hex.EncodeToString(buf[:size]))
			fmt.Println("Step:", step)

			switch step {
			case 1:
				step++
				conn.Write(b)
			case 2:
				parseData(buf, size)
			}

		} else {
			b = []byte{0}
			conn.Write(b)
			break
		}
	}
}

func parseData(data []byte, size int) []GPSElement {
	reader := bytes.NewBuffer(data)
	i := 0

	reader.Next(8)                                // header
	reader.Next(1)                                // CodecID
	recordNumber := streamToInt16(reader.Next(1)) // Number of Records
	timestamp := streamToTime(reader.Next(8))     // Timestamp
	reader.Next(1)                                // Priority

	elements := make([]GPSElement, recordNumber)

	// GPS Element
	longitude := streamToInt32(reader.Next(4)) // Longitude
	latitude := streamToInt32(reader.Next(4))  // Latitude

	reader.Next(2)                         // Altitude
	angle := streamToInt16(reader.Next(2)) // Angle
	reader.Next(1)                         // Satellites
	speed := streamToInt16(reader.Next(2)) // Speed

	fmt.Println("Number of Records:", recordNumber)
	fmt.Println("Timestamp:", timestamp)
	fmt.Println("Longitude:", longitude)
	fmt.Println("Latitude:", latitude)
	i++

	return GPSElement{longitude, latitude, angle, speed}

}

func streamToInt8(data []byte) int8 {
	var y int8
	err := binary.Read(bytes.NewReader(data), binary.BigEndian, &y)
	if err != nil {
		panic(err)
	}
	return y
}

func streamToInt16(data []byte) int16 {
	var y int16
	err := binary.Read(bytes.NewReader(data), binary.BigEndian, &y)
	if err != nil {
		panic(err)
	}
	return y
}

func streamToInt32(data []byte) int32 {
	var y int32
	err := binary.Read(bytes.NewReader(data), binary.BigEndian, &y)
	if err != nil {
		panic(err)
	}
	return y
}

func streamToInt64(data []byte) int64 {
	var y int64
	err := binary.Read(bytes.NewReader(data), binary.BigEndian, &y)
	if err != nil {
		panic(err)
	}
	return y
}

func streamToTime(data []byte) time.Time {
	miliseconds := streamToInt64(data)
	seconds := int64(miliseconds / 1000)
	nanoseconds := int64(miliseconds % 1000)

	return time.Unix(seconds, nanoseconds)
}
