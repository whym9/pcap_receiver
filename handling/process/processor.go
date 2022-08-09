package config

import (
	"database/sql"
	"fmt"
	"io"
	"log"
	"net"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type PcapProcess interface {
	Process(fileName string) (Protocols, error)
}

type DBSaver interface {
	Create(DBMS, dbName string)
	SaveToDB(counter Protocols, filePath string)
}

type Handle struct{}

type Protocols struct {
	TCP  int `json: "TCP"`
	UDP  int `json: "UDP"`
	IPv4 int `json: "IPv4"`
	IPv6 int `json: "IPv6"`
}

var (
	eth layers.Ethernet
	ip4 layers.IPv4
	ip6 layers.IPv6
	tcp layers.TCP
	udp layers.UDP
	dns layers.DNS
)

func (h Handle) Process(fileName string) (Protocols, error) {

	parser := gopacket.NewDecodingLayerParser(
		layers.LayerTypeEthernet,
		&eth,
		&ip4,
		&ip6,
		&tcp,
		&udp,
		&dns,
	)

	decoded := make([]gopacket.LayerType, 0, 10)

	handle, err := pcap.OpenOffline(fileName)

	defer handle.Close()

	if err != nil {
		log.Fatal(err)
		return Protocols{}, err
	}

	counter := Protocols{}

	for {
		data, _, err := handle.ZeroCopyReadPacketData()
		if err == io.EOF {
			break
		}

		if err != nil {
			log.Fatal(err)
			return Protocols{}, err
		}

		parser.DecodeLayers(data, &decoded)

		for _, layer := range decoded {
			if layer == layers.LayerTypeTCP {
				counter.TCP++
			}
			if layer == layers.LayerTypeUDP {
				counter.UDP++
			}
			if layer == layers.LayerTypeIPv4 {
				counter.IPv4++
			}
			if layer == layers.LayerTypeIPv6 {
				counter.IPv6++
			}
		}

	}
	return counter, nil

}

func ReceiveALL(connect net.Conn, size uint64) ([]byte, error) {
	read := make([]byte, size)

	_, err := io.ReadFull(connect, read)
	if err != nil {
		log.Fatal(err)
		return []byte{}, err
	}

	return read, nil
}

type Saver struct {
	DB gorm.DB
}

func (saver *Saver) Create(DBMS, dbName string) {
	sqlDB, err := sql.Open("mysql", "myserver_dsn")
	if err != nil {

		log.Fatal(err)
		return
	}
	gormDB, err := gorm.Open(mysql.New(mysql.Config{
		Conn: sqlDB,
	}), &gorm.Config{})

	saver.DB = *gormDB
}

//var gormDB *gorm.DB

// func DB() *gorm.DB {

// 	return gormDB
// }

type Statistics struct {
	gorm.Model
	PathToFile string
	TCP        int
	UDP        int
	IPv4       int
	IPv6       int
}

func (saver *Saver) SaveToDB(counter Protocols, filePath string) {

	result := saver.DB.Create(&Statistics{
		Model:      gorm.Model{},
		PathToFile: filePath,
		TCP:        counter.TCP,
		UDP:        counter.UDP,
		IPv4:       counter.IPv4,
		IPv6:       counter.IPv6,
	})

	if result.Error != nil {
		log.Fatal(result.Error)
		return
	}

	fmt.Printf("Record saved to Database!")

}
