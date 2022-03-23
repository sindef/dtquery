package main

import (
	"bytes"
	"encoding/binary"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"
)

//Global Variables - Flags & Wait Group
var wg sync.WaitGroup

var (
	count   = flag.Int("c", 1, "Number of requests to send")
	server  = flag.String("s", "", "Server to send the requests to")
	timeout = flag.Int("t", 5, "Timeout for the requests")
	port    = flag.Int("p", 53, "Port to send the requests to")
	//Type of DNS query to send
	queryType = flag.String("q", "A", "Type of DNS query to send")
)

//Type of DNS query

type DNSQuery struct {
	ID     uint16 // An arbitary 16 bit request identifier (same id is used in the response)
	QR     bool   // A 1 bit flat specifying whether this message is a query (0) or a response (1)
	Opcode uint8  // A 4 bit fields that specifies the query type; 0 (standard), 1 (inverse), 2 (status), 4 (notify), 5 (update)

	AA           bool  // Authoriative answer
	TC           bool  // 1 bit flag specifying if the message has been truncated
	RD           bool  // 1 bit flag to specify if recursion is desired (if the DNS server we secnd out request to doesn't know the answer to our query, it can recursively ask other DNS servers)
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
func (q DNSQuery) encode() []byte {

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
		buffer.Write(question.encode())
	}

	return buffer.Bytes()
}

func (q DNSQuestion) encode() []byte {
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

//Function to generate random domain names
func randomDomain() string {
	var buffer bytes.Buffer
	//Generate a random number between 1 and 14
	length := uint8(rand.Intn(14) + 1)
	//Generate a random string of length length
	for i := uint8(0); i < length; i++ {
		buffer.WriteByte(byte(rand.Intn(26) + 97))
		//Add a period before the penultimate character
		if i == length-3 {
			buffer.WriteByte(46)
		}
	}
	return buffer.String()
}

func main() {
	//Parse the flags (Command Line Arguments)
	flag.Parse()
	//Format the server and port (i.e. "10.0.0.1:53")
	host := *server + ":" + strconv.Itoa(*port)

	//Check the DNS type - return the hex value for the type (rfc1035 & rfc3596)
	var qt = uint16(0)
	switch *queryType {
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

	//Create a buffer to run only 10000 goroutines at a time (I only tested this on an 8 core CPU with 32G RAM - but with more resources you can open more sockets)
	count := *count
	maxGoroutines := 10000
	maxChan := make(chan struct{}, maxGoroutines)

	for i := 0; i < count; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			//Wait for a goroutine to be available
			maxChan <- struct{}{}
			defer func() {
				<-maxChan
			}()
			q := DNSQuestion{
				Domain: randomDomain(),
				Type:   qt,  // Hex value for the type
				Class:  0x1, // Internet
			}

			query := DNSQuery{
				ID:        0xAAAA,
				RD:        true,
				Questions: []DNSQuestion{q},
			}

			// Setup a UDP connection
			conn, err := net.Dial("udp", host)
			if err != nil {
				log.Fatal("failed to connect:", err)
			}
			defer conn.Close()

			if err := conn.SetDeadline(time.Now().Add(15 * time.Second)); err != nil {
				log.Fatal("failed to set deadline: ", err)
			}
			//Print the query

			encodedQuery := query.encode()

			conn.Write(encodedQuery)

		}()
	}
	//Wait for all the goroutines to finish
	wg.Wait()
	fmt.Println(strconv.Itoa(count) + " queries sent to " + host)
}
