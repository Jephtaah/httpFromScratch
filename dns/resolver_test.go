package dns

import (
	"encoding/binary"
	"net"
	"testing"
)

// cannedResponse builds a hand-crafted DNS reply for the given query ID:
// the question echoed back, plus one A record answering with 93.184.216.34.
func cannedResponse(queryID uint16) []byte {
	resp := make([]byte, 0, 64)
	resp = binary.BigEndian.AppendUint16(resp, queryID)          // ID
	resp = binary.BigEndian.AppendUint16(resp, 0x8180)           // flags: QR, RD, RA
	resp = binary.BigEndian.AppendUint16(resp, 1)                // QDCOUNT
	resp = binary.BigEndian.AppendUint16(resp, 1)                // ANCOUNT
	resp = binary.BigEndian.AppendUint16(resp, 0)                // NSCOUNT
	resp = binary.BigEndian.AppendUint16(resp, 0)                // ARCOUNT
	resp = append(resp, 0x07, 'e', 'x', 'a', 'm', 'p', 'l', 'e') // label "example"
	resp = append(resp, 0x03, 'c', 'o', 'm')                     // label "com"
	resp = append(resp, 0x00)                                    // root
	resp = binary.BigEndian.AppendUint16(resp, 1)                // QTYPE: A
	resp = binary.BigEndian.AppendUint16(resp, 1)                // QCLASS: IN
	resp = append(resp, 0xC0, 0x0C)                              // answer name: pointer to offset 12
	resp = binary.BigEndian.AppendUint16(resp, 1)                // TYPE: A
	resp = binary.BigEndian.AppendUint16(resp, 1)                // CLASS: IN
	resp = binary.BigEndian.AppendUint32(resp, 60)               // TTL
	resp = binary.BigEndian.AppendUint16(resp, 4)                // RDLENGTH
	resp = append(resp, 93, 184, 216, 34)                        // 93.184.216.34
	return resp
}

func startMockDNS(t *testing.T) net.PacketConn {
	t.Helper()

	mock, err := net.ListenPacket("udp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("starting mock dns server: %v", err)
	}

	go func() {
		buf := make([]byte, maxPacketSize)
		for {
			n, addr, err := mock.ReadFrom(buf)
			if err != nil {
				return
			}
			if n < 2 {
				continue
			}
			queryID := binary.BigEndian.Uint16(buf[0:2])
			mock.WriteTo(cannedResponse(queryID), addr)
		}
	}()

	return mock
}

func TestBuildQueryWithID(t *testing.T) {
	query, err := buildQueryWithID("example.com", 0x1234)
	if err != nil {
		t.Fatalf("buildQueryWithID: %v", err)
	}

	want := []byte{
		0x12, 0x34, // ID
		0x01, 0x00, // flags: recursion desired
		0x00, 0x01, // QDCOUNT
		0x00, 0x00, // ANCOUNT
		0x00, 0x00, // NSCOUNT
		0x00, 0x00, // ARCOUNT
		0x07, 'e', 'x', 'a', 'm', 'p', 'l', 'e',
		0x03, 'c', 'o', 'm',
		0x00,
		0x00, 0x01, // QTYPE: A
		0x00, 0x01, // QCLASS: IN
	}

	if string(query) != string(want) {
		t.Errorf("query mismatch:\n got  %x\n want %x", query, want)
	}
}

func TestEncodeNameRejectsBadLabels(t *testing.T) {
	if _, err := encodeName(""); err != nil {
		t.Errorf("empty hostname should encode to root, got error: %v", err)
	}
	if _, err := encodeName("a..b"); err == nil {
		t.Error("expected error for empty label, got nil")
	}
	if _, err := encodeName("trailing."); err != nil {
		t.Errorf("trailing dot should be accepted, got error: %v", err)
	}
}

func TestResolveWithServer(t *testing.T) {
	mock := startMockDNS(t)
	defer mock.Close()

	ips, err := ResolveWithServer("example.com", mock.LocalAddr().String())
	if err != nil {
		t.Fatalf("ResolveWithServer: %v", err)
	}
	if len(ips) != 1 {
		t.Fatalf("expected 1 IP, got %v", ips)
	}
	if ips[0] != "93.184.216.34" {
		t.Errorf("expected 93.184.216.34, got %s", ips[0])
	}
}
