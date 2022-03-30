package main

import (
	"dtquery/dns"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net"
	"strconv"
	"sync"
	"time"
)

//Flags are defined globally here - these are the command line arguments and what will be returned when queried with -h flag
var (
	count     = flag.Int("c", 1, "Number of requests to send")
	server    = flag.String("s", "", "Server to send the requests to")
	timeout   = flag.Int("t", 5, "Timeout for the requests")
	port      = flag.Int("p", 53, "Port to send the requests to")
	queryType = flag.String("q", "A", "Type of DNS query to send")
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
	var qt = dns.Type(*queryType)

	//Create sync group & mutex for the goroutines
	wg := &sync.WaitGroup{}
	mut := &sync.Mutex{}

	//Check whether random, quick or slow mode is enabled - warn that one of these modes have to be selected:
	if *random == false && *quick == 0 && *slow == 0 {
		fmt.Println("You must select either random, fast or slow mode using the flags -r, -f or -S")
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

	//If slow mode was enabled, return a slice of all possible domain names of that length. Else return an empty slice
	var all []string
	if *slow != 0 {
		all = dns.AllWords(*slow, *domain)
	}

	//Create variables for the amount of requests to send and the domain name to send
	count := *count
	var domName string

	/*Main for loop - for each request, create a goroutine to send the request. When the goroutine finishes, decrement the waitgroup
	and unlock the mutex. I did try this by allowing race conditions and it was more of a DoS tool - sending all requests at once, rather than
	creating arbitrary load. TODO: Add some concurrency*/
	for i := 0; i < count; i++ {

		//Adds to the wait group - if the wait group reaches 0, the main function will complete (https://pkg.go.dev/sync)
		wg.Add(1)

		// For the request number, create a goroutine to send the request (this runs concurrently, so as soon as the go routine is created, the program continues)
		go func(wg *sync.WaitGroup, m *sync.Mutex) {
			/*The generation of these domain names from the dictionary is our shared resource in memory, so a mutex is used to stop memory going brr.
			This is particularly noticed in slow mode where the whole file needs to be parsed.  */
			m.Lock()
			defer wg.Done()
			defer m.Unlock()

			//If the random flag is set, set domName to the dns.Random function
			if *random {
				domName = dns.Random()
			} else {
				if *domain != "" {
					//Check if a . has been prepended, if not, add one to the beginning of the TLD string
					if (*domain)[0] != '.' {
						*domain = "." + *domain

					}
				}
			}
			//If the quick flag is set, set domName to the dns.QuickWord function
			if *quick > 0 {
				domName = dns.QuickWord(*quick, *domain)
			} else if *slow > 0 {
				//Use a random domain name from the 'all' slice
				domName = all[rand.Intn(len(all))]

			}

			//Create a new DNS question - Contains our type and domain name.
			q := dns.DNSQuestion{
				Domain: domName,
				Type:   qt,
				Class:  0x1,
			}
			//Create the query - generate random ID - rfc1035
			query := dns.DNSQuery{
				ID:        dns.RandomID(),
				RD:        true,
				Questions: []dns.DNSQuestion{q},
			}

			fmt.Println("Sending query to ", host, ": ", domName)
			// Setup a UDP connection to the host
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

			//Send the encoded query to the server
			conn.Write(encodedQuery)

		}(wg, mut)
	}
	//Wait for all the goroutines to finish
	wg.Wait()
	fmt.Println(strconv.Itoa(count) + " queries sent to " + host)
}
