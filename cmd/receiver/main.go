package main

import (
	"encoding/binary"
	"log"
	"net"
)

const winSize = 256

var sent [winSize]uint32
var sentValid [winSize]bool

func alreadySent(seq uint32) bool {
	idx := seq % winSize
	return sentValid[idx] && sent[idx] == seq
}

func markSent(seq uint32) {
	idx := seq % winSize
	sent[idx] = seq
	sentValid[idx] = true
}

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
	redFrame := make([]byte, 164)

	for {
		n, _, err := inConn.ReadFromUDP(buf)
		if err != nil || n < 164 {
			continue
		}

		seq := binary.BigEndian.Uint32(buf[0:4])
		if !alreadySent(seq) {
			outConn.Write(buf[0:164])
			markSent(seq)
		}

		if n >= 324 {
			var k uint32 = 3

			if seq >= k {
				redSeq := seq - k
				if !alreadySent(redSeq) {
					binary.BigEndian.PutUint32(redFrame[0:4], redSeq)
					copy(redFrame[4:164], buf[164:324])

					outConn.Write(redFrame)
					markSent(redSeq)
				}
			}
		}
	}
}
