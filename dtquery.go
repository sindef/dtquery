package main

import (
	"dtquery/dns"
	"flag"
	"fmt"
	"log"
	"net"
	"strconv"
	"sync"
	"time"
)

//Global Variables - Flags & Wait Group
var wg sync.WaitGroup

var (
	count     = flag.Int("c", 1, "Number of requests to send")
	server    = flag.String("s", "", "Server to send the requests to")
	timeout   = flag.Int("t", 5, "Timeout for the requests")
	port      = flag.Int("p", 53, "Port to send the requests to")
	queryType = flag.String("q", "A", "Type of DNS query to send") //Type of DNS query to send
	random    = flag.Bool("r", false, "Use random domain names")
	domain    = flag.String("d", "", "Top level domain to use - e.g. com")
	quick     = flag.Int("f", 0, "Use quick/fast mode - this will send the first domain name multiple times of the specified length")
	slow      = flag.Int("S", 0, "Use slow mode - this will send the a random name each time - Specify the length of the random name")
)

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

	var domName string

	//Check whether random, quick or slow mode is enabled - warn that one of these modes have to be selected:
	if *random == false && *quick == 0 && *slow == 0 {
		fmt.Println("You must select either random, quick or slow mode using the flags -r, -f or -S")
		return
	}

	//If quick or slow mode are enabled, check that the domain is specified
	if *quick != 0 || *slow != 0 {
		if *domain == "" {
			fmt.Println("You must specify a top level domain using the flag -d")
			fmt.Println("e.g. dtquery -c 10 -s 127.0.0.1 -S 10 -d com")
			return
		}
	}

	for i := 0; i < count; i++ {
		wg.Add(1)
		go func() {
			//Wait for a goroutine to be available
			maxChan <- struct{}{}
			defer func() {
				<-maxChan
			}()

			//If the random flag is set, set domName to the dns.Random function
			if *random {
				domName = dns.Random()
			} else {
				//Check if the domain flag is set
				if *domain != "" {
					//Check if a . has been prepended, if not, add one to the *domain var
					if (*domain)[0] != '.' {
						*domain = "." + *domain

					}
				}
				//If the quick flag is set, set domName to the dns.QuickWord function
				if *quick > 0 {
					domName = dns.QuickWord(*quick, *domain)
				} else {
					//Else set domName to the slower dns.RandomWords function
					domName = dns.RandomWords(*slow, *domain)

				}
			}
			defer wg.Done()

			q := dns.DNSQuestion{
				Domain: domName,
				Type:   qt,  // Hex value for the type
				Class:  0x1, // Internet
			}

			query := dns.DNSQuery{
				ID:        0xAAAA,
				RD:        true,
				Questions: []dns.DNSQuestion{q},
			}
			fmt.Println("Sending query to ", host, ": ", query)
			// Setup a UDP connection
			conn, err := net.Dial("udp", host)
			if err != nil {
				log.Fatal("failed to connect:", err)
			}
			defer conn.Close()

			//Set deadline for the request to the specified timeout
			if err := conn.SetDeadline(time.Now().Add(time.Duration(*timeout) * time.Second)); err != nil {
				log.Fatal("failed to set deadline: ", err)
			}
			encodedQuery := dns.Uencode(query)

			conn.Write(encodedQuery)

		}()
	}
	//Wait for all the goroutines to finish
	wg.Wait()
	fmt.Println(strconv.Itoa(count) + " queries sent to " + host)
}
