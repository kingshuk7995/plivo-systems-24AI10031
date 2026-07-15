package main

import (
	"encoding/binary"
	"log"
	"net"
)

func main() {
	harnessAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:47010")
	if err != nil {
		log.Fatal(err)
	}
	inConn, err := net.ListenUDP("udp", harnessAddr)
	if err != nil {
		log.Fatal(err)
	}
	defer inConn.Close()

	relayAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:47001")
	if err != nil {
		log.Fatal(err)
	}
	outConn, err := net.DialUDP("udp", nil, relayAddr)
	if err != nil {
		log.Fatal(err)
	}
	defer outConn.Close()

	const k = 1
	history := make(map[uint32][]byte)

	buf := make([]byte, 2048)
	for {
		n, _, err := inConn.ReadFromUDP(buf)
		if err != nil || n < 4 {
			continue
		}

		seq := binary.BigEndian.Uint32(buf[:4])

		payload := make([]byte, n-4)
		copy(payload, buf[4:n])
		history[seq] = payload

		if seq >= k+10 {
			delete(history, seq-(k+10))
		}

		outBuf := make([]byte, 0, 324)

		outBuf = append(outBuf, buf[:n]...)

		if seq%20 != 0 && seq >= k {
			target := seq - k
			if red, ok := history[target]; ok {
				outBuf = append(outBuf, red...)
			}
		}

		outConn.Write(outBuf)
	}
}
