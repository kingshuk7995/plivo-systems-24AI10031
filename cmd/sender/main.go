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

	// Use a ring buffer instead of a map to prevent GC churn
	const ringSize = 256
	var history [ringSize][]byte
	for i := 0; i < ringSize; i++ {
		history[i] = make([]byte, 160)
	}

	buf := make([]byte, 2048)
	outBuf := make([]byte, 324)
	
	var totalRaw, totalUp uint64

	for {
		n, _, err := inConn.ReadFromUDP(buf)
		if err != nil || n < 164 {
			continue
		}

		seq := binary.BigEndian.Uint32(buf[:4])

		idx := seq % ringSize
		copy(history[idx], buf[4:164])

		copy(outBuf[0:164], buf[:164])

		var k uint32 = 3

		// 2. Append redundant payload using a running bandwidth budget.
		// Maximum allowed overhead is 2.00x of raw payloads (n * 160).
		totalRaw += 160
		haveRedundant := seq >= k
		withRedundantLen := uint64(324)
		withoutRedundantLen := uint64(164)
		
		packetLen := int(withoutRedundantLen)
		if haveRedundant && (totalUp + withRedundantLen) <= 2 * totalRaw {
			target := seq - k
			targetIdx := target % ringSize
			copy(outBuf[164:324], history[targetIdx])
			packetLen = int(withRedundantLen)
			totalUp += withRedundantLen
		} else {
			totalUp += withoutRedundantLen
		}

		outConn.Write(outBuf[:packetLen])
	}
}
