## Usage

`hstats` is a tool for parallelly obtaining statistics of connectivity and
response time for a list of remote hosts.

	yousong@jumper:~/go/hstats$ ./hstats -h
	Usage of ./hstats:
	  -count int
			Number of ping probes to send for each host (default 4)
	  -gonum int
			Number of parallel cmd to execute (default 16)
	  -infile string
			Input file containing list of hosts (default "-")

## Build and Run

Input is a list of hosts, addresses to ping.  They need to be separated by
whitespaces.

	go build .
	./hstats.go <henet-tunnelserver.txt

## Examples

Read host list from file.

	yousong@jumper:~/go/hstats$ ./hstats -infile henet-tunnelserver.txt
		       Host       Min      =Avg       Max     Jit.
	      216.218.221.6     51.06     60.28     71.65    20.59
	     216.218.221.42     86.40     98.42    107.52    21.12
		 74.82.46.6    109.00    115.57    123.22    14.22
	       66.220.18.42    194.23    197.71    201.63     7.39
	       72.52.104.74    204.38    214.41    225.39    21.01
	       216.66.80.26    214.67    221.42    229.38    14.71
	     184.105.253.10    224.37    230.68    234.56    10.19
	    216.218.226.238    226.34    233.79    249.25    22.91
	       216.66.80.30    231.05    240.64    249.58    18.52
	      216.66.77.230    233.53    242.52    247.61    14.08
	       216.66.80.98    243.36    246.92    251.85     8.49
	      216.66.80.162    247.76    253.68    263.19    15.43
	     184.105.255.26    250.88    259.64    279.47    28.60
	      209.51.161.14    253.04    260.27    275.79    22.74
	     184.105.253.14    253.32    261.46    268.90    15.58
	       216.66.38.58    253.24    261.73    277.64    24.41
		216.66.22.2    260.89    268.85    276.41    15.52
	       216.66.80.90    260.87    269.14    279.12    18.25
	      209.51.161.58    263.99    271.98    282.56    18.56
	     184.105.250.46    269.50    274.14    278.63     9.13
	      64.62.134.130    270.53    275.51    281.76    11.23
	       216.66.84.42    286.87    293.30    311.30    24.43
	       216.66.84.46    284.66    293.73    314.19    29.53
	      216.66.86.114    295.67    306.42    324.93    29.26
	      216.66.86.122    298.64    308.13    329.49    30.86
	       216.66.87.14    313.58    319.87    331.25    17.66

Read host list from stdin (normally ends with Ctrl-D).

	yousong@jumper:~/go/hstats$ ./hstats
	8.8.8.8
	114.114.114.114
	[ping -c 4 8.8.8.8]: exit status 1
		       Host         Min        =Avg         Max     Jit.
	    114.114.114.114        8.07       18.36       25.80    17.74
		    8.8.8.8    10000.00    10000.00    10000.00     0.00

