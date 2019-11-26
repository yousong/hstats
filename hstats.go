package main

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"text/tabwriter"
)

type Stat struct {
	min, avg, max, jit float64
}

var MAX_STAT = &Stat{
	min: 10000,
	max: 10000,
	avg: 10000,
	jit: 10000,
}

type HostStat struct {
	host string
	Stat
}

type Pinger struct {
	//isiputils bool
	path      string
	iswindows bool
}

func NewPinger() (*Pinger, error) {
	path, err := exec.LookPath("ping")
	if err != nil {
		return nil, err
	}
	p := &Pinger{
		path:      path,
		iswindows: runtime.GOOS == "windows",
	}
	return p, nil
}

func (p *Pinger) cmd(ctx context.Context, count int, host string) *exec.Cmd {
	var args []string
	if p.iswindows {
		args = append(args, "-n", fmt.Sprintf("%d", count))
	} else {
		args = append(args, "-c", fmt.Sprintf("%d", count))
	}
	args = append(args, host)
	cmd := exec.CommandContext(ctx, p.path, args...)
	return cmd
}

func (p *Pinger) parseOutput(output string) *Stat {
	// iputils
	//
	//   rtt min/avg/max/mdev = 0.016/0.020/0.025/0.005 ms
	//
	// OSX
	//
	//   round-trip min/avg/max/stddev = 6.841/7.521/8.084/0.514 ms
	//
	// Windows
	//
	//   Minimum = 40ms, Maximum = 42ms, Average = 41ms
	//
	var (
		reg *regexp.Regexp
		cnt int
	)
	if p.iswindows {
		reg = regexp.MustCompile("Minimum = ([0-9]+)ms, Maximum = ([0-9]+)ms, Average = ([0-9]+)ms")
		cnt = 3
	} else {
		reg = regexp.MustCompile(` ([0-9\.]+)/([0-9\.]+)/([0-9\.]+)/([0-9\.]+) `)
		cnt = 4
	}
	m := reg.FindStringSubmatch(output)
	if len(m) != cnt+1 {
		return nil
	}
	var (
		nums = make([]float64, cnt)
		err  error
	)
	m = m[1:]
	for i := range m {
		nums[i], err = strconv.ParseFloat(m[i], 32)
		if err != nil {
			return nil
		}
	}
	var stat *Stat
	if p.iswindows {
		stat = &Stat{
			min: nums[0],
			max: nums[1],
			avg: nums[2],
		}
	} else {
		stat = &Stat{
			min: nums[0],
			avg: nums[1],
			max: nums[2],
			jit: nums[3],
		}
	}
	return stat
}

// run runs fn(ctx) n times with concurrency c
func run(ctx context.Context, c int, fn func(context.Context)) chan<- struct{} {
	ch := make(chan struct{})
	for i := 0; i < c; i++ {
		go func() {
			for {
				select {
				case _, ok := <-ch:
					if !ok {
						return
					}
					fn(ctx)
				case <-ctx.Done():
					return
				}
			}
		}()
	}
	return ch
}

func runN(ctx context.Context, c, n int, fn func(context.Context)) {
	if c > n {
		c = n
	}
	ch := run(ctx, c, fn)
	defer close(ch)
	for i := 0; i < n; i++ {
		select {
		case <-ctx.Done():
			return
		case ch <- struct{}{}:
		}
	}
}

func runForever(ctx context.Context, c int, fn func(context.Context)) {
	ch := run(ctx, c, fn)
	defer close(ch)
	for {
		select {
		case <-ctx.Done():
			return
		case ch <- struct{}{}:
		}
	}
}

var gonum int
var pingcount int
var infile string

func init() {
	flag.StringVar(&infile, "infile", "-", "Input file containing list of hosts")
	flag.IntVar(&gonum, "gonum", 16, "Number of parallel cmd to execute")
	flag.IntVar(&pingcount, "count", 4, "Number of ping probes to send for each host")
	flag.Parse()
}

func main() {
	hosts := []string{}
	{
		var (
			d   []byte
			err error
		)
		if infile == "" || infile == "-" {
			d, err = ioutil.ReadAll(os.Stdin)
		} else {
			d, err = ioutil.ReadFile(infile)
		}
		if err != nil {
			log.Fatalf("read %q: %v", infile, err)
		}
		parts := strings.Fields(string(d))
		for _, part := range parts {
			part = strings.TrimSpace(part)
			if part == "" {
				continue
			}
			hosts = append(hosts, part)
		}
		if len(hosts) == 0 {
			log.Fatalf("no hosts found")
		}
	}

	var (
		reqch  = make(chan string)
		respch = make(chan *HostStat)
		pinger *Pinger
		hstats []*HostStat
	)
	{
		var err error
		pinger, err = NewPinger()
		if err != nil {
			log.Fatalf("new pinger: %v", err)
		}
	}
	{
		wg := &sync.WaitGroup{}
		wg.Add(len(hosts) + 1)
		ctx := context.Background()
		go func() {
			defer wg.Done()
			for _, host := range hosts {
				reqch <- host
			}
		}()
		go runN(ctx, gonum, len(hosts), func(ctx context.Context) {
			defer wg.Done()
			select {
			case h := <-reqch:
				cmd := pinger.cmd(ctx, pingcount, h)
				out, err := cmd.Output()
				if err != nil {
					log.Printf("warn: ping cmd: %v", err)
					return
				}
				stat := pinger.parseOutput(string(out))
				if stat == nil {
					stat = MAX_STAT
				}
				hstat := &HostStat{
					host: h,
					Stat: *stat,
				}
				respch <- hstat
			case <-ctx.Done():
				return
			}
		})
		go func() {
			for stat := range respch {
				if stat != nil {
					hstats = append(hstats, stat)
				}
			}
		}()
		wg.Wait()
		close(respch)
	}

	sort.Slice(hstats, func(i, j int) bool {
		if hstats[i].avg != hstats[j].avg {
			return hstats[i].avg < hstats[j].avg
		}
		if hstats[i].jit != hstats[j].jit {
			return hstats[i].jit < hstats[j].jit
		}
		if hstats[i].min != hstats[j].min {
			return hstats[i].min < hstats[j].min
		}
		return hstats[i].max < hstats[j].max
	})
	w := tabwriter.NewWriter(os.Stdout, 0, 8, 4, ' ', tabwriter.AlignRight)
	fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t\n", "Host", "Min", `^Avg`, "Max", "Jit.")
	for i := range hstats {
		s := hstats[i]
		fmt.Fprintf(w, "%s\t%.2f\t%.2f\t%.2f\t%.2f\t\n", s.host, s.min, s.avg, s.max, s.max-s.min)
	}
	w.Flush()
}
