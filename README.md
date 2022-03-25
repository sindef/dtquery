Simple CLI tool for testing DNS servers. This is NOT for malicious intent and was only made for load testing of a server.
This will not return an output of the query.
Use a -h flag to see all the options.

Usage: dnstest [-h] [-c COUNT] [-s SERVER] [-t TIMEOUT] [-p PORT] [-S LENGTH] [-f LENGTH] [-r]

-S int
Use slow mode - this will send the a random name each time - Specify the length of the random name
-c int
Number of requests to send (default 1)
-d string
Top level domain to use - e.g. com
-f int
Use quick/fast mode - this will send the first domain name multiple times of the specified length
Port to send the requests to (default 53)
-q string
Type of DNS query to send (default "A")
-r    Use random domain names
-s string
Server to send the requests to
-t int
Timeout for the requests (default 5)

Examples:
	dnstest -c 10 -s 10.0.0.1