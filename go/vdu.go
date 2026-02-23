package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
)

const __version__ string = "0.2.0"
const __video_extensions__ string = "mp4,mkv,avi,mov,flv,wmv,webm,m4v"

type Vdu struct {
	file_formats string
}

func NewVdu() *Vdu {
	return &Vdu{
		file_formats: __video_extensions__,
	}
}

func get_extension(filename string) string {
	dot_position := strings.LastIndex(filename, ".")
	if dot_position != -1 {
		return filename[dot_position+1:]
	}
	return ""
}

func (vdu *Vdu) is_video_file(filename string) bool {
	return strings.Contains(vdu.file_formats, get_extension(filename))
}

func get_args() (string, bool, bool, bool, bool) {
	show_version := func() {
		fmt.Fprintf(flag.CommandLine.Output(), "VDU Video Disk Used "+__version__+"\n\n")
	}
	flag.Usage = func() {
		show_version()
		fmt.Fprintf(flag.CommandLine.Output(), "Usage: %s [OPTIONS] [DIRECTORY]\n\n", os.Args[0])
		flag.PrintDefaults()
	}
	help := false
	flag.BoolVar(&help, "h", false, "")
	flag.BoolVar(&help, "help", false, "show help")

	version := false
	flag.BoolVar(&version, "v", false, "")
	flag.BoolVar(&version, "version", false, "show version")

	summarize := false
	flag.BoolVar(&summarize, "s", false, "summarize")

	sort := false
	flag.BoolVar(&sort, "S", false, "sort")

	reverse := false
	flag.BoolVar(&reverse, "r", false, "reverse the sort")
	flag.Parse()

	if flag.NArg() == 0 || flag.NArg() > 1 || help {
		flag.Usage()
		fmt.Println()
		os.Exit(0)
	}
	if version {
		show_version()
		os.Exit(0)
	}
	dir := flag.Arg(0)
	if info, err := os.Stat(dir); os.IsNotExist(err) || !info.IsDir() {
		show_version()
		fmt.Println("No directory " + dir)
		os.Exit(1)
	}
	return dir, version, summarize, sort, reverse
}

func main() {
	dir, version, summarize, sort, reverse := get_args()
	dir, version, summarize, sort, reverse = dir, version, summarize, sort, reverse
	vdu := NewVdu()
	fmt.Println(vdu.file_formats)
}
