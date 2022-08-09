package TCP

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	process "pcap_receiver/handling/process"
	"strconv"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcapgo"
)

var dirName string

var DB process.Saver

func HandleTCP(addrs, dirN string, DBase process.Saver) {
	DB = DBase
	dirName = dirN
	err := os.MkdirAll(dirName, os.ModePerm)

	if err != nil {
		log.Fatalf("couldn't create path, %v", err)
		return
	}
	server, err := net.Listen("tcp", addrs)
	if err != nil {
		log.Fatal(err)
		return
	}
	fmt.Println("TCP Server has started")

	for {
		connect, err := server.Accept()

		if err != nil {
			log.Fatal(err)
			return
		}
		go countTCPAndUDP(connect)

	}
}

type Capture struct {
	TimeStamp      time.Time     `json: "time"`
	CaptureLength  int           `json: "caplength"`
	Length         int           `json: "length"`
	InterfaceIndex int           `json :  "index"`
	AccalaryData   []interface{} `json: "accalary"`
}

func countTCPAndUDP(connect net.Conn) {

	fileName := dirName + "/" + time.Now().Format("02-01-2006-111111.456") + ".pcap"

	file, err := os.Create(fileName)

	if err != nil {
		log.Fatal(err)
		return
	}

	defer file.Close()

	w := pcapgo.NewWriter(file)
	w.WriteFileHeader(65535, layers.LinkTypeEthernet)
	for {
		read, err := process.ReceiveALL(connect, 8)

		if err != nil {
			log.Fatal(err)
			return
		}

		size := binary.BigEndian.Uint64(read)
		read, err = process.ReceiveALL(connect, size)
		if size == 4 && string(read) == "STOP" {
			break
		}
		var capi Capture
		json.Unmarshal(read, &capi)

		read, err = process.ReceiveALL(connect, 8)
		size = binary.BigEndian.Uint64(read)
		read, err = process.ReceiveALL(connect, size)

		fmt.Printf("File size: %v\n", size)

		packet := gopacket.NewPacket(read, layers.LayerTypeEthernet, gopacket.Default)

		packet.Metadata().CaptureInfo.Timestamp = capi.TimeStamp
		packet.Metadata().CaptureInfo.CaptureLength = capi.CaptureLength
		packet.Metadata().CaptureInfo.Length = capi.Length
		packet.Metadata().CaptureInfo.InterfaceIndex = capi.InterfaceIndex
		packet.Metadata().CaptureInfo.AncillaryData = capi.AccalaryData

		w.WritePacket(packet.Metadata().CaptureInfo, packet.Data())

	}
	counter, _ := process.Handle{}.Process(fileName)
	DB.SaveToDB(counter, fileName)

	res := "TCP: " + strconv.Itoa(counter.TCP) + "\n" +
		"UDP: " + strconv.Itoa(counter.UDP) + "\n" +
		"IPv4: " + strconv.Itoa(counter.IPv4) + "\n" +
		"IPv6: " + strconv.Itoa(counter.IPv6) + "\n"

	connect.Write([]byte(res))
	connect.Close()
	fmt.Println("File receiving has ended")
	fmt.Println()

}
