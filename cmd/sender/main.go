package main

import (
	"encoding/binary"
	"log"
	"net"
	"transport/internal/common"
)

type HarnessFrame struct {
	Seq     uint32
	Payload [common.PayloadBytes]byte
}

func main() {
	harnessAddr, err := net.ResolveUDPAddr("udp", common.PortHarnessSource)
	if err != nil {
		log.Fatal(err)
	}
	inConn, err := net.ListenUDP("udp", harnessAddr)
	if err != nil {
		log.Fatal(err)
	}
	defer inConn.Close()

	relayAddr, err := net.ResolveUDPAddr("udp", common.PortRelayUplink)
	if err != nil {
		log.Fatal(err)
	}
	outConn, err := net.DialUDP("udp", nil, relayAddr)
	if err != nil {
		log.Fatal(err)
	}
	defer outConn.Close()

	nackAddr, err := net.ResolveUDPAddr("udp", common.PortRelayFeedbackOut)
	if err != nil {
		log.Fatal(err)
	}
	nackConn, err := net.ListenUDP("udp", nackAddr)
	if err != nil {
		log.Fatal(err)
	}
	defer nackConn.Close()

	var history [common.HistoryRingSize][]byte
	for i := 0; i < common.HistoryRingSize; i++ {
		history[i] = make([]byte, common.PayloadBytes)
	}
	var highestSeqSeen uint32 = 0

	harnessChan := make(chan HarnessFrame, 200)
	nackChan := make(chan uint32, 200)

	go func() {
		buf := make([]byte, 2048)
		for {
			n, _, err := inConn.ReadFromUDP(buf)
			if err == nil && n >= common.FrameBytes {
				var frame HarnessFrame
				frame.Seq = binary.BigEndian.Uint32(buf[:common.SeqBytes])
				copy(frame.Payload[:], buf[common.SeqBytes:common.FrameBytes])
				harnessChan <- frame
			}
		}
	}()

	go func() {
		buf := make([]byte, 4)
		for {
			n, _, err := nackConn.ReadFromUDP(buf)
			if err == nil && n == 4 {
				seq := binary.BigEndian.Uint32(buf)
				nackChan <- seq
			}
		}
	}()

	outBuf := make([]byte, common.PacketBytes)
	var totalRaw, totalUsed uint64
	var k uint32 = common.RedundancyOffset

	for {
		select {
		case frame := <-harnessChan:
			if frame.Seq > highestSeqSeen {
				highestSeqSeen = frame.Seq
			}
			idx := frame.Seq % common.HistoryRingSize
			copy(history[idx], frame.Payload[:])

			binary.BigEndian.PutUint32(outBuf[0:4], frame.Seq)
			copy(outBuf[4:common.FrameBytes], frame.Payload[:])

			totalRaw += common.PayloadBytes
			haveRedundant := frame.Seq >= k
			withRedundantLen := uint64(common.PacketBytes)
			withoutRedundantLen := uint64(common.FrameBytes)

			packetLen := int(withoutRedundantLen)
			marginMax := totalRaw * 2
			if haveRedundant && (totalUsed+withRedundantLen) <= marginMax {
				target := frame.Seq - k
				targetIdx := target % common.HistoryRingSize
				copy(outBuf[common.FrameBytes:common.PacketBytes], history[targetIdx])
				packetLen = int(withRedundantLen)
				totalUsed += withRedundantLen
			} else {
				totalUsed += withoutRedundantLen
			}

			outConn.Write(outBuf[:packetLen])

		case nackSeq := <-nackChan:
			if highestSeqSeen >= nackSeq && (highestSeqSeen-nackSeq) < common.HistoryRingSize {
				retransBuf := make([]byte, common.FrameBytes)
				binary.BigEndian.PutUint32(retransBuf[0:4], nackSeq)

				idx := nackSeq % common.HistoryRingSize
				copy(retransBuf[4:], history[idx])

				totalUsed += 168
				outConn.Write(retransBuf)
			}
		}
	}
}
