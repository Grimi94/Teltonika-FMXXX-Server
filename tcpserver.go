package main

import (
        "bytes"
        "encoding/binary"
        "encoding/hex"
        "fmt"
        "math"
        "net"
        "os"
        "time"
)

const (
        CONN_HOST = "0.0.0.0"
        CONN_PORT = "4554"
        CONN_TYPE = "tcp"
)

const PRECISION = 10000000.0

type GPSElement struct {
        latitude  float64
        longitude float64
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
                        b = []byte{1} // 0x01 if we accept the message

                        fmt.Println("----------------------------------------")
                        fmt.Println("Data From:", conn.RemoteAddr().String())
                        fmt.Println("Size of message: ", size)
                        fmt.Println("Message:", hex.EncodeToString(buf[:size]))
                        fmt.Println("Step:", step)

                        switch step {
                        case 1:
                                step++
                                conn.Write(b)
                        case 2:
                                elements := parseData(buf, size)
                                conn.Write([]byte{0, 0, 0, uint8(len(elements))})
                        }

                } else {
                        b = []byte{0} // 0x00 if we decline the message

                        conn.Write(b)
                        break
                }
        }
}

func parseData(data []byte, size int) []GPSElement {
        reader := bytes.NewBuffer(data)
        var i int8 = 0

        reader.Next(8)                               // header
        reader.Next(1)                               // CodecID
        recordNumber := streamToInt8(reader.Next(1)) // Number of Records
        fmt.Println("Number of Records:", recordNumber)
        timestamp := streamToTime(reader.Next(8)) // Timestamp
        reader.Next(1)                            // Priority

        elements := make([]GPSElement, recordNumber)

        for i < recordNumber {
                // GPS Element
                longitude := float64(streamToInt32(reader.Next(4))) / PRECISION // Longitude
                latitude := float64(streamToInt32(reader.Next(4))) / PRECISION  // Latitude

                reader.Next(2)                         // Altitude
                angle := streamToInt16(reader.Next(2)) // Angle
                reader.Next(1)                         // Satellites
                speed := streamToInt16(reader.Next(2)) // Speed

                elements[i] = GPSElement{longitude, latitude, angle, speed}

                // IO Events Elements

                _ = streamToInt8(reader.Next(1)) // ioEventID
                totalElements := streamToInt8(reader.Next(1))

                stage := 0
                var j int8 = 0

                for j < totalElements {
                        stageElements := streamToInt8(reader.Next(1))
                        var k int8 = 0

                        for k < stageElements {
                                _ = streamToInt8(reader.Next(1)) // elementID

                                switch stage {
                                case 1: // One byte IO Elements
                                        _ = streamToInt8(reader.Next(1))
                                case 2: // Two byte IO Elements
                                        _ = streamToInt16(reader.Next(2))
                                case 3: // Four byte IO Elements
                                        _ = streamToInt32(reader.Next(4))
                                case 4: // Eigth byte IO Elements
                                        _ = streamToInt64(reader.Next(8))
                                }
                                j++
                        }
                        stage++
                }

                // fmt.Println("Element:", i)
                fmt.Println("Timestamp:", timestamp)
                fmt.Println("Longitude:", longitude)
                fmt.Println("Latitude:", latitude)

                i++
        }

        // Once finished with the records we read the Record Number and the CRC

        _ = streamToInt8(reader.Next(1))  // Number of Records
        _ = streamToInt32(reader.Next(4)) // CRC

        return elements
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
        if y>>31 == 1 {
                y *= -1
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

func streamToFloat32(data []byte) float32 {
        var y float32
        err := binary.Read(bytes.NewReader(data), binary.BigEndian, &y)
        if err != nil {
                panic(err)
        }
        return y
}

func twos_complement(input int32) int32 {
        fmt.Printf("%33b", input)
        mask := int32(math.Pow(2, 31))
        return -(input & mask) + (input &^ mask)
}

func streamToTime(data []byte) time.Time {
        miliseconds := streamToInt64(data)
        seconds := int64(float64(miliseconds) / 1000.0)
        nanoseconds := int64(miliseconds % 1000)

        return time.Unix(seconds, nanoseconds)
}
