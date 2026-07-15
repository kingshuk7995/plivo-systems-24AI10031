package common

const (
	// Network Ports
	PortHarnessSource = "127.0.0.1:47010"
	PortRelayUplink   = "127.0.0.1:47001"
	PortRelayDownlink = "127.0.0.1:47002"
	PortRelayFeedbackIn = "127.0.0.1:47003"
	PortRelayFeedbackOut = "127.0.0.1:47004"
	PortHarnessPlayer = "127.0.0.1:47020"
	
	// Sizes & Limits
	SeqBytes       = 4
	PayloadBytes   = 160
	FrameBytes     = SeqBytes + PayloadBytes // 164
	RedundantBytes = PayloadBytes            // 160
	PacketBytes    = FrameBytes + RedundantBytes // 324
	
	// Buffers & Offsets
	HistoryRingSize = 256
	RedundancyOffset = 3 // k=3
)
