package main

import (
	"fmt"
	"os"
	"os/exec"
	"flag"
	"regexp"
	"runtime"
	"strconv"
	"sort"
	"text/tabwriter"
	"io/ioutil"
)

type HostStat struct {
	host          string
	min, avg, max float64
}

type HostStats []HostStat

func (hs HostStats) Len() int { return len(hs) }

func (hs HostStats) Swap(i, j int) { hs[i], hs[j] = hs[j], hs[i] }

func (hs HostStats) Less(i, j int) bool {
	return hs[i].avg < hs[j].avg
}

var reStat *regexp.Regexp
var reHost *regexp.Regexp
var pingArgs []string

func ping(reqch chan string, respch chan *HostStat) {
	for {
		var stat HostStat
		var cmd exec.Cmd
		host := <-reqch
		stat.host = host
		cmd.Path, _ = exec.LookPath("ping")
		cmd.Args = append(pingArgs, host)
		if out, err := cmd.Output(); err == nil {
			m := reStat.FindStringSubmatch(string(out))
			stat.min, _ = strconv.ParseFloat(m[1], 32)
			stat.avg, _ = strconv.ParseFloat(m[2], 32)
			stat.max, _ = strconv.ParseFloat(m[3], 32)
		} else {
			stat.min = 10000
			stat.avg = 10000
			stat.max = 10000
			fmt.Fprintf(os.Stderr, "%v: %s\n", cmd.Args, err)
		}
		respch <- &stat
	}
}

var gonum int
var pingcount int
var infile string
func init() {
	flag.StringVar(&infile, "infile", "STDIN", "Input file containing one host by each line.")
	flag.IntVar(&gonum, "gonum", 16, "Number of parallel cmd to execute.")
	flag.IntVar(&pingcount, "count", 4, "Number of echo request to send.")
	flag.Parse()
}

func main() {
	var reqch = make(chan string)
	var respch = make(chan *HostStat)
	var i = 0

	// OS dependent initialization
	if runtime.GOOS == "linux" {
		/*
		 *   rtt min/avg/max/mdev = 0.016/0.020/0.025/0.005 ms
		 */
		reStat = regexp.MustCompile("([0-9\\.]+)/([0-9\\.]+)/([0-9\\.]+)")
		pingArgs = []string{"ping", "-c", strconv.Itoa(pingcount)}
	} else if runtime.GOOS == "windows" {
		/*
		 *   Minimum = 40ms, Maximum = 42ms, Average = 41ms
		 */
		reStat = regexp.MustCompile("Minimum = ([0-9]+)ms, Maximum = ([0-9]+)ms, Average = ([0-9]+)ms")
		pingArgs = []string{"ping", "-n", strconv.Itoa(pingcount)}
	} else {
		fmt.Fprintf(os.Stderr, "Not supported OS: %s\n", runtime.GOOS)
		os.Exit(1)
	}

	d, err := ioutil.ReadFile(infile)
	if err != nil {
		d, err = ioutil.ReadAll(os.Stdin)
	}
	reHost = regexp.MustCompile("[^\\s]+")
	m := reHost.FindAll(d, -1)
	var hosts []string
	for i := range m {
		hosts = append(hosts, string(m[i]))
	}
	if len(hosts) <= 0 {
		fmt.Fprintf(os.Stderr, "No valid host were found.\n")
		os.Exit(1)
	}

	// Start and feed gonum go routine.
	if len(hosts) < gonum {
		gonum = len(hosts)
	}
	for j := 0; j < gonum; j++ {
		go ping(reqch, respch)
	}
	for j := 0; j < gonum; j++ {
		reqch <- hosts[i]
		i++
	}

	// Collect results and send pending requests.
	var stats = []HostStat{}
	for {
		stats = append(stats, *<-respch)
		if i < len(hosts) {
			reqch <- hosts[i]
			i++
		}
		if len(stats) >= len(hosts) {
			break
		}
	}

	// Sort and print results.
	sort.Sort(HostStats(stats))
	w := new(tabwriter.Writer)
	w.Init(os.Stdout, 0, 8, 4, ' ', tabwriter.AlignRight)
	fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t\n", "Host", "Min", "=Avg", "Max", "Jit.")
	for i := range stats {
		s := stats[i]
		fmt.Fprintf(w, "%s\t%.2f\t%.2f\t%.2f\t%.2f\t\n", s.host, s.min, s.avg, s.max, s.max - s.min)
	}
	w.Flush()
}
