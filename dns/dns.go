package dns

//DNS package - this is a slightly modified version of https://github.com/vishen/go-dnsquery - but is a fairly standard implementation

import (
	"bytes"
	"encoding/binary"
	"log"
	"math/rand"
	"strings"
)

type DNSQuery struct {
	ID     uint16 // An arbitary 16 bit request identifier (same id is used in the response)
	QR     bool   // A 1 bit flat specifying whether this message is a query (0) or a response (1)
	Opcode uint8  // A 4 bit fields that specifies the query type; 0 (standard), 1 (inverse), 2 (status), 4 (notify), 5 (update)

	AA           bool  // Authoriative answer
	TC           bool  // 1 bit flag specifying if the message has been truncated
	RD           bool  // 1 bit flag to specify if recursion is desired (if the DNS server we send out request to doesn't know the answer to our query, it can recursively ask other DNS servers)
	RA           bool  // Recursive available
	Z            uint8 // Reserved for future use
	ResponseCode uint8

	QDCount uint16 // Number of entries in the question section
	ANCount uint16 // Number of answers
	NSCount uint16 // Number of authorities
	ARCount uint16 // Number of additional records

	Questions []DNSQuestion
}

type DNSQuestion struct {
	Domain string
	Type   uint16 // DNS record type - this is set by the command line flag
	Class  uint16 // 1
}

//DNS Query function - this function takes a DNSQuery struct and encodes it into a byte array
func Uencode(q DNSQuery) []byte {

	q.QDCount = uint16(len(q.Questions))

	var buffer bytes.Buffer
	//Write the ID
	binary.Write(&buffer, binary.BigEndian, q.ID)

	b2i := func(b bool) int {
		if b {
			return 1
		}

		return 0
	}

	queryParams1 := byte(b2i(q.QR)<<7 | int(q.Opcode)<<3 | b2i(q.AA)<<1 | b2i(q.RD))
	queryParams2 := byte(b2i(q.RA)<<7 | int(q.Z)<<4)
	//Write the query parameters
	binary.Write(&buffer, binary.BigEndian, queryParams1)
	binary.Write(&buffer, binary.BigEndian, queryParams2)
	binary.Write(&buffer, binary.BigEndian, q.QDCount)
	binary.Write(&buffer, binary.BigEndian, q.ANCount)
	binary.Write(&buffer, binary.BigEndian, q.NSCount)
	binary.Write(&buffer, binary.BigEndian, q.ARCount)

	//Write the questions
	for _, question := range q.Questions {
		buffer.Write(Qencode(question))
	}

	return buffer.Bytes()
}

func Qencode(q DNSQuestion) []byte {
	var buffer bytes.Buffer

	domainParts := strings.Split(q.Domain, ".")
	for _, part := range domainParts {
		if err := binary.Write(&buffer, binary.BigEndian, byte(len(part))); err != nil {
			log.Fatalf("Error binary.Write(..) for '%s': '%s'", part, err)
		}

		for _, c := range part {
			if err := binary.Write(&buffer, binary.BigEndian, uint8(c)); err != nil {
				log.Fatalf("Error binary.Write(..) for '%s'; '%c': '%s'", part, c, err)
			}
		}
	}

	binary.Write(&buffer, binary.BigEndian, uint8(0))
	binary.Write(&buffer, binary.BigEndian, q.Type)
	binary.Write(&buffer, binary.BigEndian, q.Class)

	return buffer.Bytes()

}

//Generates a random value between 43690 and 65535 as a uint16
func RandomID() uint16 {
	id := rand.Intn(65535-43690) + 43690
	return uint16(id)
}

//Given a string, return a uint16 of the DNS type
func Type(query string) uint16 {
	var qt uint16
	switch query {
	case "A":
		qt = 0x1
	case "NS":
		qt = 0x2
	case "CNAME":
		qt = 0x5
	case "SOA":
		qt = 0x6
	case "PTR":
		qt = 0x12
	case "MX":
		qt = 0x15
	case "TXT":
		qt = 0x16
	case "AAAA":
		qt = 0x1c
	default:
		qt = 0x1
	}
	return qt
}
