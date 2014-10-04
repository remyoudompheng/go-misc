package nbf

import (
	"encoding/binary"
	"fmt"
	"strconv"
	"time"
	"unicode/utf16"
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
	Peer string
	Text string
	// From PDU
	Number  string
	Stamp   time.Time
	Payload []byte // raw SMS-encoded payload
}

// SMS encoding.
// Inspired by libgammu's libgammu/phone/nokia/dct4s40/6510/6510file.c

// Structure: all integers are big-endian
// u16 u16 u32 u32(size)
// [82]byte (zero)
// [41]uint16 (NUL-terminated peer name)
// PDU (offset is 0xb0)
// 64 unknown bytes
// 0001 0003 size(uint16) [size/2]uint16 (NUL-terminated text)
// 02 size(uint16) + NUL-terminated [size]byte (SMS center)
// 04 0001 002b size(uint16) + [size]byte (NUL-terminated UTF16BE) (peer)
// [23]byte unknown data

func parseMessage(s []byte) (rawMessage, error) {
	// peer (fixed offset 0x5e)
	var runes []uint16
	for off := 0x5e; s[off]|s[off+1] != 0; off += 2 {
		runes = append(runes, binary.BigEndian.Uint16(s[off:off+2]))
	}
	peer := string(utf16.Decode(runes))

	// PDU frame starts at 0xb0
	// incoming PDU frame:
	// * NN 91 <NN/2 bytes> (NN : number of BCD digits, little endian)
	//   source number, padded with 0xf halfbyte.
	// * 00 FF (data format, GSM 03.40 section 9.2.3.10)
	// * YY MM DD HH MM SS ZZ (BCD date time, little endian)
	// * NN <NN septets> (NN : number of packed 7-bit data)
	// received SMS: 04 0b 91
	pdu := s[0xb0:]
	nbLen := pdu[1]
	number := decodeBCD(pdu[3 : 3+(nbLen+1)/2])
	//log.Printf("number: %s", number)
	pdu = pdu[3+(nbLen+1)/2:]

	// Date time
	pdu = pdu[2:]
	stamp := parseDateTime(pdu[:7])
	//log.Printf("stamp: %s", stamp)
	pdu = pdu[7:]

	// Payload
	length := int(pdu[0])
	packedLen := length - length/8
	data := unpack7bit(pdu[1 : 1+packedLen])
	//log.Printf("payload: %q", translateSMS(data, &basicSMSset))
	pdu = pdu[1+packedLen:]

	// END of PDU.
	if len(pdu) < 72 {
		return rawMessage{}, fmt.Errorf("truncated message")
	}
	pdu = pdu[65:]
	length = int(pdu[5])
	pdu = pdu[6:]
	text := make([]rune, length/2)
	for i := range text {
		text[i] = rune(binary.BigEndian.Uint16(pdu[2*i : 2*i+2]))
	}
	//log.Printf("%q", string(text))

	m := rawMessage{
		Peer:    peer,
		Number:  number,
		Stamp:   stamp,
		Payload: data,
		Text:    string(text),
	}
	return m, nil
}

// Ref: GSM 03.40 section 9.2.3.11
func parseDateTime(b []byte) time.Time {
	var dt [7]int
	for i := range dt {
		dt[i] = int(b[i]&0xf)*10 + int(b[i]>>4)
	}
	return time.Date(
		2000+dt[0],
		time.Month(dt[1]),
		dt[2],
		dt[3], dt[4], dt[5], 0, time.FixedZone("", dt[6]*3600/4))
}

func decodeBCD(b []byte) string {
	s := make([]byte, 0, len(b)*2)
	for _, c := range b {
		s = append(s, '0'+(c&0xf))
		if c>>4 == 0xf {
			break
		} else {
			s = append(s, '0'+(c>>4))
		}
	}
	return string(s)
}

func unpack7bit(s []byte) []byte {
	// each byte may contain a part of septet i in lower bits
	// and septet i+1 in higher bits.
	buf := uint16(0)
	buflen := uint(0)
	out := make([]byte, 0, len(s)+len(s)/7+1)
	for len(s) > 0 {
		buf |= uint16(s[0]) << buflen
		buflen += 8
		s = s[1:]
		for buflen >= 7 {
			out = append(out, byte(buf&0x7f))
			buflen -= 7
			buf >>= 7
		}
	}
	return out
}

// translateSMS decodes a 7-bit encoded SMS text into a standard
// UTF-8 encoded string.
func translateSMS(s []byte, charset *[128]rune) string {
	r := make([]rune, len(s))
	for i, b := range s {
		r[i] = charset[b]
	}
	return string(r)
}

// See http://en.wikipedia.org/wiki/GSM_03.38

var basicSMSset = [128]rune{
	// 0x00
	'@', '£', '$', '¥', 'è', 'é', 'ù', 'ì',
	'ò', 'Ç', '\n', 'Ø', 'ø', '\r', 'Å', 'å',
	// 0x10
	'Δ', '_', 'Φ', 'Γ', 'Λ', 'Ω', 'Π', 'Ψ',
	'Σ', 'Θ', 'Ξ', -1 /* ESC */, 'Æ', 'æ', 'ß', 'É',
	// 0x20
	' ', '!', '"', '#', '¤', '%', '&', '\'',
	'(', ')', '*', '+', ',', '-', '.', '/',
	// 0x30
	'0', '1', '2', '3', '4', '5', '6', '7',
	'8', '9', ':', ';', '<', '=', '>', '?',
	// 0x40
	'¡', 'A', 'B', 'C', 'D', 'E', 'F', 'G',
	'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O',
	// 0x50
	'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W',
	'X', 'Y', 'Z', 'Ä', 'Ö', 'Ñ', 'Ü', '§',
	// 0x60
	'¿', 'a', 'b', 'c', 'd', 'e', 'f', 'g',
	'h', 'i', 'j', 'k', 'l', 'm', 'n', 'o',
	// 0x70
	'p', 'q', 'r', 's', 't', 'u', 'v', 'w',
	'x', 'y', 'z', 'ä', 'ö', 'ñ', 'ü', 'à',
}
