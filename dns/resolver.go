// Package dns implements a minimal DNS resolver by constructing query
// packets and parsing responses by hand — no net.LookupHost involved.
package dns

import (
	"crypto/rand"
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"strings"
	"time"
)

const (
	defaultServer  = "8.8.8.8:53"
	defaultTimeout = 5 * time.Second

	maxPacketSize = 512

	typeA     = 1
	typeNS    = 2
	typeCNAME = 5
	typeAAAA  = 28
)

type Record struct {
	Name  string
	Type  uint16
	Class uint16
	TTL   uint32
	Data  string
}

type Response struct {
	ID      uint16
	Answers []Record
}

// Resolve looks up A records for hostname using the default DNS server.
func Resolve(hostname string) ([]string, error) {
	return ResolveWithServer(hostname, defaultServer)
}

// ResolveWithServer looks up A records for hostname against the given DNS
// server (host:port) and returns every IPv4 address found in the answers.
func ResolveWithServer(hostname, dnsServer string) ([]string, error) {
	query, err := buildQuery(hostname)
	if err != nil {
		return nil, fmt.Errorf("dns: building query for %q: %w", hostname, err)
	}

	conn, err := net.Dial("udp", dnsServer)
	if err != nil {
		return nil, fmt.Errorf("dns: dialing %s: %w", dnsServer, err)
	}
	defer conn.Close()

	if err := conn.SetDeadline(time.Now().Add(defaultTimeout)); err != nil {
		return nil, fmt.Errorf("dns: setting deadline: %w", err)
	}

	if _, err := conn.Write(query); err != nil {
		return nil, fmt.Errorf("dns: writing query to %s: %w", dnsServer, err)
	}

	packet := make([]byte, maxPacketSize)
	n, err := conn.Read(packet)
	if err != nil {
		return nil, fmt.Errorf("dns: reading response from %s: %w", dnsServer, err)
	}

	resp, err := parseResponse(packet[:n])
	if err != nil {
		return nil, err
	}

	var ips []string
	for _, answer := range resp.Answers {
		if answer.Type == typeA {
			ips = append(ips, answer.Data)
		}
	}
	return ips, nil
}

func buildQuery(hostname string) ([]byte, error) {
	id, err := randomID()
	if err != nil {
		return nil, err
	}
	return buildQueryWithID(hostname, id)
}

func buildQueryWithID(hostname string, id uint16) ([]byte, error) {
	qname, err := encodeName(hostname)
	if err != nil {
		return nil, err
	}

	header := make([]byte, 12)
	binary.BigEndian.PutUint16(header[0:2], id)
	binary.BigEndian.PutUint16(header[2:4], 0x0100) // recursion desired, standard query
	binary.BigEndian.PutUint16(header[4:6], 1)      // QDCOUNT: one question

	question := append([]byte(nil), qname...)
	question = binary.BigEndian.AppendUint16(question, typeA) // QTYPE
	question = binary.BigEndian.AppendUint16(question, 1)     // QCLASS: IN

	return append(header, question...), nil
}

func randomID() (uint16, error) {
	var b [2]byte
	if _, err := rand.Read(b[:]); err != nil {
		return 0, fmt.Errorf("generating query id: %w", err)
	}
	return binary.BigEndian.Uint16(b[:]), nil
}

func encodeName(hostname string) ([]byte, error) {
	hostname = strings.TrimSuffix(hostname, ".")
	if hostname == "" {
		return []byte{0}, nil
	}

	var encoded []byte
	for _, label := range strings.Split(hostname, ".") {
		if len(label) == 0 || len(label) > 63 {
			return nil, fmt.Errorf("invalid label %q in hostname %q", label, hostname)
		}
		encoded = append(encoded, byte(len(label)))
		encoded = append(encoded, label...)
	}
	return append(encoded, 0), nil
}

func parseResponse(packet []byte) (*Response, error) {
	if len(packet) < 12 {
		return nil, fmt.Errorf("dns: response too short (%d bytes)", len(packet))
	}

	id := binary.BigEndian.Uint16(packet[0:2])
	flags := binary.BigEndian.Uint16(packet[2:4])

	if flags&0x8000 == 0 {
		return nil, errors.New("dns: response has QR bit clear, not a reply")
	}

	switch rcode := flags & 0x000F; rcode {
	case 2:
		return nil, errors.New("dns: server failure (SERVFAIL)")
	case 3:
		return nil, errors.New("dns: name does not exist (NXDOMAIN)")
	case 4:
		return nil, errors.New("dns: query type not implemented (NOTIMP)")
	case 5:
		return nil, errors.New("dns: query refused (REFUSED)")
	}

	qdcount := binary.BigEndian.Uint16(packet[4:6])
	ancount := binary.BigEndian.Uint16(packet[6:8])

	offset := 12
	for i := 0; i < int(qdcount); i++ {
		_, next, err := decodeName(packet, offset)
		if err != nil {
			return nil, fmt.Errorf("dns: decoding question %d: %w", i, err)
		}
		offset = next + 4 // skip QTYPE and QCLASS
		if offset > len(packet) {
			return nil, errors.New("dns: question section overruns packet")
		}
	}

	answers := make([]Record, 0, ancount)
	for i := 0; i < int(ancount); i++ {
		name, next, err := decodeName(packet, offset)
		if err != nil {
			return nil, fmt.Errorf("dns: decoding answer %d name: %w", i, err)
		}
		offset = next

		if offset+10 > len(packet) {
			return nil, errors.New("dns: answer header overruns packet")
		}
		rtype := binary.BigEndian.Uint16(packet[offset : offset+2])
		rclass := binary.BigEndian.Uint16(packet[offset+2 : offset+4])
		ttl := binary.BigEndian.Uint32(packet[offset+4 : offset+8])
		rdlength := int(binary.BigEndian.Uint16(packet[offset+8 : offset+10]))
		offset += 10

		if offset+rdlength > len(packet) {
			return nil, errors.New("dns: rdata overruns packet")
		}

		data, err := formatRData(rtype, packet, offset, rdlength)
		if err != nil {
			return nil, fmt.Errorf("dns: decoding answer %d rdata: %w", i, err)
		}
		offset += rdlength

		answers = append(answers, Record{
			Name:  name,
			Type:  rtype,
			Class: rclass,
			TTL:   ttl,
			Data:  data,
		})
	}

	return &Response{ID: id, Answers: answers}, nil
}

func formatRData(rtype uint16, packet []byte, rdataOffset, rdlength int) (string, error) {
	switch rtype {
	case typeA:
		if rdlength != net.IPv4len {
			return "", fmt.Errorf("A record has %d bytes of data, want %d", rdlength, net.IPv4len)
		}
		return net.IP(packet[rdataOffset : rdataOffset+rdlength]).String(), nil
	case typeAAAA:
		if rdlength != net.IPv6len {
			return "", fmt.Errorf("AAAA record has %d bytes of data, want %d", rdlength, net.IPv6len)
		}
		return net.IP(packet[rdataOffset : rdataOffset+rdlength]).String(), nil
	case typeCNAME, typeNS:
		name, _, err := decodeName(packet, rdataOffset)
		if err != nil {
			return "", err
		}
		return name, nil
	default:
		return fmt.Sprintf("0x%x", packet[rdataOffset:rdataOffset+rdlength]), nil
	}
}

func decodeName(packet []byte, offset int) (string, int, error) {
	var name strings.Builder
	pos := offset
	next := offset
	jumped := false
	jumps := 0

	for {
		if pos >= len(packet) {
			return "", 0, errors.New("name overruns packet")
		}

		length := int(packet[pos])
		switch {
		case length == 0:
			if !jumped {
				next = pos + 1
			}
			return name.String(), next, nil
		case length&0xC0 == 0xC0:
			if pos+1 >= len(packet) {
				return "", 0, errors.New("compression pointer overruns packet")
			}
			pointer := ((length & 0x3F) << 8) | int(packet[pos+1])
			jumps++
			if jumps > 10 {
				return "", 0, errors.New("compression pointer loop detected")
			}
			if !jumped {
				next = pos + 2
				jumped = true
			}
			pos = pointer
		case length&0xC0 != 0:
			return "", 0, fmt.Errorf("invalid label type %#x", length)
		default:
			if pos+1+length > len(packet) {
				return "", 0, errors.New("label overruns packet")
			}
			if name.Len() > 0 {
				name.WriteByte('.')
			}
			name.Write(packet[pos+1 : pos+1+length])
			pos += 1 + length
		}
	}
}
