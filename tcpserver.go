package main

import (
        "bytes"
        "encoding/binary"
        "encoding/hex"
        "fmt"
        "gopkg.in/mgo.v2"
        "math"
        "net"
        "os"
        "strconv"
        "time"
)

const (
        CONN_HOST = "0.0.0.0"
        CONN_PORT = "4554"
        CONN_TYPE = "tcp"
)

const PRECISION = 10000000.0
const MONGO_DATABASE = ""
const MONGO_PLACES_COLLECTION = ""
const MONGO_RECORD_COLLECTION = ""
const MONGO_URL = ""

type GPSElement struct {
        latitude  float64
        longitude float64
        angle     int16
        speed     int16
}

var pc *mgo.Collection = nil
var rc *mgo.Collection = nil

func main() {
        // Initialize mongo connector
        session, err := mgo.Dial(MONGO_URL)
        if err != nil {
                fmt.Println("mgo error: ", err.Error())
                os.Exit(1)
        }

        pc = session.DB(MONGO_DATABASE).C(MONGO_PLACES_COLLECTION)
        rc = session.DB(MONGO_DATABASE).C(MONGO_RECORD_COLLECTION)
        fmt.Println(pc.Database)
        fmt.Println(rc.Database)

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
                                elements, err := parseData(buf, size)
                                if err != nil {
                                        fmt.Println("Error while parsing data", err)
                                        break
                                }

                                for i := 0; i < len(elements); i++ {
                                        element := elements[i]
                                        err := rc.Insert(&element)
                                        if err != nil {
                                                fmt.Println("Error inserting element to database", err)
                                        }
                                }

                                conn.Write([]byte{0, 0, 0, uint8(len(elements))})
                        }

                } else {
                        b = []byte{0} // 0x00 if we decline the message

                        conn.Write(b)
                        break
                }
        }
}

func parseData(data []byte, size int) (elements []GPSElement, err error) {
        reader := bytes.NewBuffer(data)
        fmt.Println("Reader Size:", reader.Len())

        // Header
        reader.Next(4)                                                        // 4 Zero Bytes
        dataLength, err := streamToInt32(reader.Next(4))                      // Header
        reader.Next(1)                                                        // CodecID
        recordNumber, err := strconv.Atoi(hex.EncodeToString(reader.Next(1))) // Number of Records
        fmt.Println("Length of data:", dataLength)

        elements = make([]GPSElement, recordNumber)

        i := 0
        for i < recordNumber {
                timestamp, err := streamToTime(reader.Next(8)) // Timestamp
                reader.Next(1)                                 // Priority

                // GPS Element
                longitudeInt, err := streamToInt32(reader.Next(4)) // Longitude
                longitude := float64(longitudeInt) / PRECISION
                latitudeInt, err := streamToInt32(reader.Next(4)) // Latitude
                latitude := float64(latitudeInt) / PRECISION

                reader.Next(2)                              // Altitude
                angle, err := streamToInt16(reader.Next(2)) // Angle
                reader.Next(1)                              // Satellites
                speed, err := streamToInt16(reader.Next(2)) // Speed

                if err != nil {
                        fmt.Println("Error while reading GPS Element")
                        break
                }

                elements[i] = GPSElement{longitude, latitude, angle, speed}

                // IO Events Elements

                reader.Next(1) // ioEventID
                reader.Next(1) // total Elements

                stage := 1
                for stage <= 4 {
                        stageElements, err := streamToInt8(reader.Next(1))
                        if err != nil {
                                break
                        }

                        var j int8 = 0
                        for j < stageElements {
                                reader.Next(1) // elementID

                                switch stage {
                                case 1: // One byte IO Elements
                                        _, err = streamToInt8(reader.Next(1))
                                case 2: // Two byte IO Elements
                                        _, err = streamToInt16(reader.Next(2))
                                case 3: // Four byte IO Elements
                                        _, err = streamToInt32(reader.Next(4))
                                case 4: // Eigth byte IO Elements
                                        _, err = streamToInt64(reader.Next(8))
                                }
                                j++
                        }
                        stage++
                }

                if err != nil {
                        fmt.Println("Error while reading IO Elements")
                        break
                }

                fmt.Println("Timestamp:", timestamp)
                fmt.Println("Longitude:", longitude, "Latitude:", latitude)

                i++
        }

        // Once finished with the records we read the Record Number and the CRC

        _, err = streamToInt8(reader.Next(1))  // Number of Records
        _, err = streamToInt32(reader.Next(4)) // CRC

        return
}

func streamToInt8(data []byte) (int8, error) {
        var y int8
        err := binary.Read(bytes.NewReader(data), binary.BigEndian, &y)
        return y, err
}

func streamToInt16(data []byte) (int16, error) {
        var y int16
        err := binary.Read(bytes.NewReader(data), binary.BigEndian, &y)
        return y, err
}

func streamToInt32(data []byte) (int32, error) {
        var y int32
        err := binary.Read(bytes.NewReader(data), binary.BigEndian, &y)
        if y>>31 == 1 {
                y *= -1
        }
        return y, err
}

func streamToInt64(data []byte) (int64, error) {
        var y int64
        err := binary.Read(bytes.NewReader(data), binary.BigEndian, &y)
        return y, err
}

func streamToFloat32(data []byte) (float32, error) {
        var y float32
        err := binary.Read(bytes.NewReader(data), binary.BigEndian, &y)
        return y, err
}

func twos_complement(input int32) int32 {
        mask := int32(math.Pow(2, 31))
        return -(input & mask) + (input &^ mask)
}

func streamToTime(data []byte) (time.Time, error) {
        miliseconds, err := streamToInt64(data)
        seconds := int64(float64(miliseconds) / 1000.0)
        nanoseconds := int64(miliseconds % 1000)

        return time.Unix(seconds, nanoseconds), err
}
