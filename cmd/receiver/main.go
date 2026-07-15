package main

import (
	"encoding/binary"
	"log"
	"net"
	"time"
	"transport/internal/common"
)

var sent [common.HistoryRingSize]uint32
var sentValid [common.HistoryRingSize]bool

func alreadySent(seq uint32) bool {
	idx := seq % common.HistoryRingSize
	return sentValid[idx] && sent[idx] == seq
}

func markSent(seq uint32) {
	idx := seq % common.HistoryRingSize
	sent[idx] = seq
	sentValid[idx] = true
}

func main() {
	inAddr, err := net.ResolveUDPAddr("udp", common.PortRelayDownlink)
	if err != nil {
		log.Fatal(err)
	}
	inConn, err := net.ListenUDP("udp", inAddr)
	if err != nil {
		log.Fatal(err)
	}
	defer inConn.Close()

	playerAddr, err := net.ResolveUDPAddr("udp", common.PortHarnessPlayer)
	if err != nil {
		log.Fatal(err)
	}
	outConn, err := net.DialUDP("udp", nil, playerAddr)
	if err != nil {
		log.Fatal(err)
	}
	defer outConn.Close()

	feedbackAddr, err := net.ResolveUDPAddr("udp", common.PortRelayFeedbackIn)
	if err != nil {
		log.Fatal(err)
	}
	feedbackConn, err := net.DialUDP("udp", nil, feedbackAddr)
	if err != nil {
		log.Fatal(err)
	}
	defer feedbackConn.Close()

	buf := make([]byte, 2048)
	redFrame := make([]byte, common.FrameBytes)

	highestSeq := uint32(0)
	var lastNack [common.HistoryRingSize]time.Time
	var lastNackEval time.Time

	for {
		n, _, err := inConn.ReadFromUDP(buf)
		if err != nil || n < common.FrameBytes {
			continue
		}

		seq := binary.BigEndian.Uint32(buf[0:common.SeqBytes])
		if seq > highestSeq {
			highestSeq = seq
		}

		// Extract primary frame
		if !alreadySent(seq) {
			outConn.Write(buf[0:common.FrameBytes])
			markSent(seq)
		}

		// Extract redundant payload
		var k uint32 = common.RedundancyOffset
		if n >= common.PacketBytes {
			if seq >= k {
				redSeq := seq - k
				if !alreadySent(redSeq) {
					binary.BigEndian.PutUint32(redFrame[0:common.SeqBytes], redSeq)
					copy(redFrame[common.SeqBytes:common.FrameBytes], buf[common.FrameBytes:common.PacketBytes])

					outConn.Write(redFrame)
					markSent(redSeq)
				}
			}
		}

		now := time.Now()
		if highestSeq >= k+2 && now.Sub(lastNackEval) > 10*time.Millisecond {
			lastNackEval = now
			start := uint32(0)
			if highestSeq > 100 {
				start = highestSeq - 100
			}
			for m := start; m <= highestSeq-k-2; m++ {
				if !alreadySent(m) {
					idx := m % common.HistoryRingSize
					// Don't send NACKs too frequently for the same frame
					if now.Sub(lastNack[idx]) > 80*time.Millisecond {
						nackBuf := make([]byte, 4)
						binary.BigEndian.PutUint32(nackBuf, m)
						feedbackConn.Write(nackBuf)
						lastNack[idx] = now
					}
				}
			}
		}
	}
}
