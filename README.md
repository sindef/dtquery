Simple CLI tool for testing DNS servers. This is NOT for malicious intent and was only made for load testing of a server.
This will not return an output of the query.
Use a -h flag to see all the options.

Usage: 

	dtquery [-h] [-c COUNT] [-s SERVER] [-t TIMEOUT] [-p PORT] [-S LENGTH] [-f LENGTH] [-r]

Examples:

	dtquery -c 10 -s 10.0.0.1
	
	dtquery -c 1 -s 10.0.0.1 -S 11 -d com
	
	dtquery -c 100 -s 10.0.0.1 -f 8 -d net
	
