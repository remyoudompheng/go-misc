package nbu

import (
	"encoding/binary"
	"io"
	"os"
	"time"
	"unicode/utf16"
)

// NBU format parser as produced by Nokia Communication Center.

func OpenFile(filename string) (*Reader, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	r := &Reader{File: f}
	r.Size, _ = f.Seek(0, os.SEEK_END)
	f.Seek(0, os.SEEK_SET)
	return r, nil
}

type Reader struct {
	File interface {
		io.ReaderAt
		io.Closer
	}
	Size int64
}

type FileInfo struct {
	BackupTime time.Time
	IMEI       string
	Model      string
	Name       string
	Firmware   string
	Language   string
}

// Read file metadata and TOC.
func (r *Reader) Info() (info FileInfo, err error) {
	var buf [8]byte
	// Find TOC offset.
	_, err = r.File.ReadAt(buf[:], 0x14)
	if err != nil {
		return
	}
	off := int64(binary.LittleEndian.Uint64(buf[:]))
	// Read metadata.
	sec := io.NewSectionReader(r.File, off, r.Size-off)
	info.BackupTime, err = readTime(sec)
	for _, s := range []*string{&info.IMEI, &info.Model, &info.Name, &info.Firmware, &info.Language} {
		if err == nil {
			*s, err = readString(sec)
		}
	}
	// skip 20 bytes.
	sec.Seek(0x14, os.SEEK_CUR)
	// here begin the TOC.
	_, err = sec.Read(buf[:4])
	if err != nil {
		return
	}
	parts := binary.LittleEndian.Uint32(buf[:4])
	for p := 0; p < parts; p++ {
	}
	return
}

const (
	SecFS = iota
	SecVCards
	SecGroups
	SecCalendar
	SecMemo
	SecMessages
	SecMMS
	SecBookmarks
)

var secUUID = [...][2]uint64{
	SecGroups:    {0x1f0e5865a19f3c49, 0x9e230e25eb240fe1},
	SecCalendar:  {0x16cdf8e8235e5a4e, 0xb735dddff1481222},
	SecMemo:      {0x5c62973bdca75441, 0xa1c3059de3246808},
	SecMessages:  {0x617aefd1aabea149, 0x9d9d155abb4ceb8e},
	SecMMS:       {0x471dd465efe33240, 0x8c7764caa383aa33},
	SecBookmarks: {0x7f77905631f95749, 0x8d96ee445dbebc5a},
}

// Utility functions.

// From MSDN: "A Windows file time is a 64-bit value that represents the number
// of 100-nanosecond intervals that have elapsed since 12:00 midnight, January
// 1, 1601 A.D. (C.E.) Coordinated Universal Time (UTC)."

var baseWinTime = time.Date(1601, 1, 1, 0, 0, 0, 0, time.UTC)

func readTime(r io.Reader) (time.Time, error) {
	// time is stored as: high 32 bits, low 32 bits (little-endian).
	var buf [8]byte
	_, err := io.ReadFull(r, buf[:])
	if err != nil {
		return baseWinTime, err
	}
	hi := binary.LittleEndian.Uint32(buf[:4])
	lo := binary.LittleEndian.Uint32(buf[4:])
	ticks := uint64(hi)<<32 | uint64(lo)
	days := ticks / (86400 * 1e7)
	ticks %= 86400 * 1e7
	secs, nsec := ticks/1e7, (ticks%1e7)*100
	return baseWinTime.
		AddDate(0, 0, int(days)).
		Add(time.Duration(secs) * time.Second).
		Add(time.Duration(nsec) * time.Nanosecond), nil
}

func readString(r io.Reader) (string, error) {
	// Little endian 16 bit length + UTF-16LE string.
	var buf [2]byte
	_, err := io.ReadFull(r, buf[:])
	if err != nil {
		return "", err
	}
	length := int(buf[1])<<8 | int(buf[0])
	s := make([]uint16, length)
	err = binary.Read(r, binary.LittleEndian, s)
	return string(utf16.Decode(s)), err
}
