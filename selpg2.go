package main

import (
	"bufio"
	"fmt"
	"io"
	"math"
	"os"
	"os/exec"

	flag "github.com/spf13/pflag"
)

type selpgargs struct {
	startPage  int
	endPage    int
	inFile     string
	fileDest   string
	pageLength int
	pageType   bool
}

var progname string

func usage() {
	fmt.Printf("Usage of %s:\n\n", progname)
	fmt.Printf("%s is a tool to select pages from what you want.\n\n", progname)
	fmt.Printf("Usage:\n\n")
	fmt.Printf("\tselpg -s=Number -e=Number [options] [filename]\n\n")
	fmt.Printf("The arguments are:\n\n")
	fmt.Printf("\t-s=Number\tStart from Page <number>.\n")
	fmt.Printf("\t-e=Number\tEnd to Page <number>.\n")
	fmt.Printf("\t-l=Number\t[options]Specify the number of line per page.Default is 72.\n")
	fmt.Printf("\t-f\t\t[options]Specify that the pages are sperated by \\f.\n")
	fmt.Printf("\t[filename]\t[options]Read input from the file.\n\n")
	fmt.Printf("If no file specified, %s will read input from stdin. Control-D to end.\n\n", progname)
}

func FlagInit(args *selpgargs) {
	flag.Usage = usage
	flag.IntVar(&args.startPage, "s", -1, "Start page.")
	flag.IntVar(&args.endPage, "e", -1, "End page.")
	flag.IntVar(&args.pageLength, "l", 72, "Line number per page.")
	flag.BoolVar(&args.pageType, "f", false, "Determine form-feed-delimited")
	flag.StringVar(&args.fileDest, "d", "", "specify the printer")
	flag.Parse()
	otherargs := flag.Args()
	if len(otherargs) > 0 {
		args.inFile = otherargs[0]
	} else {
		args.inFile = ""
	}
}

func ProcessArgs(args *selpgargs) {
	if args.startPage == -1 || args.endPage == -1 {
		fmt.Fprintf(os.Stderr, "%s: not enough arguments\n\n", progname)
		flag.Usage()
		os.Exit(1)
	}

	if args.startPage < -1 || args.startPage > (math.MaxInt32-1) {
		os.Stderr.Write([]byte("startPage is not valid\n"))
		flag.Usage()
		os.Exit(1)
	}
	if args.endPage < -1 || args.endPage > (math.MaxInt32-1) || args.endPage < args.startPage {
		os.Stderr.Write([]byte("endPage is not valid\n"))
		flag.Usage()
		os.Exit(2)
	}
	if args.pageLength < 1 || args.pageLength > (math.MaxInt32-1) {
		os.Stderr.Write([]byte("page length is out of range\n"))
		flag.Usage()
		os.Exit(3)
	}
}

func ProcessInput(args *selpgargs) {
	var stdin io.WriteCloser
	var err error
	var cmd *exec.Cmd

	if args.fileDest != "" {
		cmd = exec.Command("cat", "-n")
		stdin, err = cmd.StdinPipe()
		if err != nil {
			os.Stderr.Write([]byte("error happen in pipe\n"))
			os.Exit(1)
		}
	} else {
		stdin = nil
	}

	if args.inFile != "" {
		args.inFile = flag.Arg(0)
		output, err := os.Open(args.inFile)
		if err != nil {
			os.Stderr.Write([]byte("error happen in opening file\n"))
			os.Exit(1)
		}
		reader := bufio.NewReader(output)
		if args.pageType {
			for pageNum := args.startPage; pageNum <= args.endPage; pageNum++ {
				line, err := reader.ReadString('\f')
				if err != nil {
					if err == io.EOF {
						break
					}
					os.Stderr.Write([]byte("read byte from file fail\n"))
					os.Exit(1)
				}
				printOrWrite(args, string(line), stdin)
			}
		} else {
			count := 0
			for {
				line, _, err := reader.ReadLine()
				if err != nil {
					if err == io.EOF {
						break
					}
					os.Stderr.Write([]byte("read byte from file fail\n"))
					os.Exit(1)
				}
				if count/args.pageLength >= args.startPage {
					if count/args.pageLength > args.endPage {
						break
					} else {
						printOrWrite(args, string(line), stdin)
					}
				}
				count++
			}

		}
	} else {
		scanner := bufio.NewScanner(os.Stdin)
		count := 0
		target := ""
		for scanner.Scan() {
			line := scanner.Text()
			line += "\n"
			if count/args.pageLength >= args.startPage {
				if count/args.pageLength <= args.endPage {
					target += line
				}
			}
			count++
		}
		printOrWrite(args, string(target), stdin)
	}

	if args.fileDest != "" {
		stdin.Close()
		cmd.Stdout = os.Stdout
		cmd.Run()
	}
}

func printOrWrite(args *selpgargs, line string, stdin io.WriteCloser) {
	if args.fileDest != "" {
		stdin.Write([]byte(line + "\n"))
	} else {
		fmt.Println(line)
	}
}

func main() {
	progname = os.Args[0]
	var args selpgargs
	FlagInit(&args)
	ProcessArgs(&args)
	ProcessInput(&args)
}
