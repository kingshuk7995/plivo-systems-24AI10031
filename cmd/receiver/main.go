package main

import (
	"encoding/binary"
	"log"
	"net"
)

func main() {
	inAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:47002")
	if err != nil {
		log.Fatal(err)
	}
	inConn, err := net.ListenUDP("udp", inAddr)
	if err != nil {
		log.Fatal(err)
	}
	defer inConn.Close()

	playerAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:47020")
	if err != nil {
		log.Fatal(err)
	}
	outConn, err := net.DialUDP("udp", nil, playerAddr)
	if err != nil {
		log.Fatal(err)
	}
	defer outConn.Close()

	buf := make([]byte, 2048)
	const k = 1

	for {
		n, _, err := inConn.ReadFromUDP(buf)
		if err != nil || n < 164 {
			continue
		}

		seq := binary.BigEndian.Uint32(buf[0:4])
		outConn.Write(buf[0:164])

		if n >= 324 {
			redSeq := seq - k

			redFrame := make([]byte, 164)
			binary.BigEndian.PutUint32(redFrame[0:4], redSeq)
			copy(redFrame[4:], buf[164:324])

			outConn.Write(redFrame)
		}
	}
}
