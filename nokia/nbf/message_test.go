package nbf

import (
	"testing"
)

func TestMessage_ParseFilename(t *testing.T) {
	const name = "0000186F3C52A89B0042201000500000004030000000000000000000000000000+336303132330000009F"
	var msg Message
	err := msg.ParseFilename(name)
	if err != nil {
		t.Fatal(err)
	}

	if msg.Seq != 0x186f {
		t.Errorf("bad sequence number: 0x%x", msg.Seq)
	}
	if msg.Timestamp != 0x3c52a89b {
		t.Errorf("bad timestampsequence number: 0x%x", msg.Seq)
		t.Logf("timestamp: %s", DosTime(msg.Timestamp))
	}
	if msg.MultipartSeq != 0x42 {
		t.Errorf("bad multipart sequence number: %d", msg.MultipartSeq)
	}
	if msg.Flags != 0x2010 {
		t.Errorf("bad flags: 0x%x", msg.Flags)
	}
	if msg.PartNo != 3 || msg.PartTotal != 4 {
		t.Errorf("got part %d/%d, expected 3/4",
			msg.PartNo, msg.PartTotal)
	}
	if peer := string(msg.Peer[:]); peer != "+33630313233" {
		t.Errorf("wrong peer %s, expected +33630313233", peer)
	}
}
