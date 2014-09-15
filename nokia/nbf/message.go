package nbf

import (
	"strconv"
	"time"
)

// predefmessages/1: inbox
// predefmessages/3: outbox

type Message struct {
	// Zip directory information
	Date time.Time
	// Filename information
	Seq          uint32
	Timestamp    uint32
	MultipartSeq uint16
	Flags        uint16
	PartNo       uint8
	PartTotal    uint8
	Peer         [12]byte
	Pad          byte
}

// ParseFilename decomposes the filename of messages found in NBF archives.
// 00001DFC: sequence number of message
// 3CEAC364: Dos timestamp (seconds since 01 Jan 1980, 32-bit integer)
// 00B7: 16-bit multipart sequence number (identical for parts of the same message)
// 2010: 1st byte 0x20 for sms, 0x10 for mms
// 00500000:
// 00302000: for multipart: 2 out of 3.
// 00000000: zero
// 00000000: zero
// 000000000: zero (9 digits)
// 36300XXXXXXX : 12 digit number
// 0000007C : a checksum ?
func (msg *Message) ParseFilename(filename string) (err error) {
	s := filename
	s, msg.Seq, err = getUint32(s)
	if err != nil {
		return err
	}
	s, msg.Timestamp, err = getUint32(s)
	if err != nil {
		return err
	}
	s, n, err := getUint32(s)
	if err != nil {
		return err
	}
	msg.MultipartSeq = uint16(n >> 16)
	msg.Flags = uint16(n)
	s = s[8:] // skip
	s, n, err = getUint32(s)
	if err != nil {
		return err
	}
	msg.PartNo = uint8(n >> 12)
	msg.PartTotal = uint8(n >> 20)
	s = s[25:] // skip
	copy(msg.Peer[:], s[:len(msg.Peer)])
	s = s[len(msg.Peer):]
	msg.Pad = uint8(s[7])
	return nil
}

func getUint32(s string) (rest string, n uint32, err error) {
	x, err := strconv.ParseUint(s[:8], 16, 32)
	return s[8:], uint32(x), err
}

const (
	FLAGS_SMS = 0x2000
	FLAGS_MMS = 0x1000
)

func DosTime(stamp uint32) time.Time {
	t := time.Unix(int64(stamp), 0)
	// Add 10 years
	t = t.Add(3652 * 24 * time.Hour)
	return t
}

// A big-endian interpretation of the binary format.
type rawMessage struct {
	N0, N1 uint16
	N2     uint32
	Size   uint32     // The total size in bytes of the message (including headers).
	Peer   [53]uint16 // A UTF-16 big endian string
	Pad    [29]uint16 // Pads the first 176 bytes
	Box    byte       // 4 for inbox, 0xX1 for outbox
}

// 0001 0003 size(uint16) + size bytes (NUL-terminated UTF-16-BE)  (text)
// 02 size(uint16) + NUL-terminated [size]byte                     (SMS center)
// 04 0001 002b size(uint16) + [size]byte (NUL-terminated UTF16BE) (peer)
// 23 bytes
