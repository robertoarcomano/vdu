package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/fs"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	ffmpeg "github.com/u2takey/ffmpeg-go"
)

const __version__ string = "0.2.1"
const __video_extensions__ string = "mp4,mkv,avi,mov,flv,wmv,webm,m4v"

type dir_size struct {
	dir  string
	size float64
}

type Vdu struct {
	file_formats    string
	video_directory string
}

func NewVdu(video_directory string) *Vdu {
	return &Vdu{
		file_formats:    __video_extensions__,
		video_directory: video_directory,
	}
}

func (vdu Vdu) get_extension(filename string) string {
	dot_position := strings.LastIndex(filename, ".")
	if dot_position != -1 {
		return filename[dot_position+1:]
	}
	return ""
}

func (vdu Vdu) is_video_file(filename string) bool {
	extension := vdu.get_extension(filename)
	if extension == "" {
		return false
	}
	return strings.Contains(vdu.file_formats, vdu.get_extension(filename))
}
func (vdu Vdu) seconds_to_human(seconds float64) string {
	secondsInt := int(math.Round(seconds))

	h := secondsInt / 3600
	secondsInt -= h * 3600

	m := secondsInt / 60
	secondsInt -= m * 60

	s := secondsInt

	return fmt.Sprintf("%dh %02dm %02ds", h, m, s)
}

func (vdu Vdu) get_video_files(dir string, files_only bool) []string {
	video_files := []string{}
	if files_only {
		file_list, _ := os.ReadDir(dir)
		for _, file := range file_list {
			if file.IsDir() || !vdu.is_video_file(file.Name()) {
				continue
			}
			fileName := filepath.Join(dir, file.Name())
			video_files = append(video_files, fileName)
		}
	} else {
		filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
			if vdu.is_video_file(d.Name()) {
				video_files = append(video_files, path)
			}
			return nil
		})
	}
	return video_files
}

func (vdu Vdu) get_duration(filePath string, finished chan float64) {
	type probeData struct {
		Format struct {
			Duration string `json:"duration"`
		} `json:"format"`
	}
	var data probeData
	raw, _ := ffmpeg.Probe(filePath)
	json.Unmarshal([]byte(raw), &data)
	duration, _ := strconv.ParseFloat(data.Format.Duration, 64)
	finished <- duration
}

func (vdu Vdu) get_directory_duration(dir string, files_only bool) float64 {
	size := 0.0
	files := vdu.get_video_files(dir, files_only)
	finished := make(chan float64, len(files))
	for _, file := range files {
		go vdu.get_duration(file, finished)
	}
	for range files {
		size = size + <-finished
	}
	return size
}

func (vdu Vdu) get_max_durations_dir_len(durations []dir_size) int {
	max_durations_dir_len := 0
	for _, duration := range durations {
		if len(duration.dir) > max_durations_dir_len {
			max_durations_dir_len = len(duration.dir)
		}
	}
	return max_durations_dir_len
}

func (vdu Vdu) get_max_duration_size_len(durations []dir_size) int {
	max_duration_size_len := 0
	for _, duration := range durations {
		if len(vdu.seconds_to_human(duration.size)) > max_duration_size_len {
			max_duration_size_len = len(vdu.seconds_to_human(duration.size))
		}
	}
	return max_duration_size_len
}

func (vdu *Vdu) get_durations(isSummarized bool, isSorted bool, isReversed bool) ([]dir_size, int, int) {
	durations := []dir_size{}
	total := 0.0

	if isSummarized {
		main_dirs, _ := os.ReadDir(vdu.video_directory)
		for _, dir := range main_dirs {
			if dir.IsDir() == false {
				continue
			}
			full_directory_name := filepath.Join(vdu.video_directory, dir.Name())
			sum := vdu.get_directory_duration(full_directory_name, false)
			total += sum
			durations = append(durations, dir_size{
				dir:  filepath.Join(vdu.video_directory, dir.Name()),
				size: sum,
			})
		}
	} else {
		filepath.WalkDir(vdu.video_directory, func(path string, d fs.DirEntry, err error) error {
			if d.IsDir() && path != vdu.video_directory {
				sum := vdu.get_directory_duration(path, true)
				if sum > 0 {
					durations = append(durations, dir_size{
						dir:  path,
						size: sum,
					})
					total += sum
				}
			}
			return nil
		})
	}

	if isSorted {
		sort.Slice(durations, func(i, j int) bool {
			return (durations[i].size < durations[j].size) != isReversed
		})
	}

	sum := vdu.get_directory_duration(vdu.video_directory, true)
	total += sum
	durations = append(durations, dir_size{
		dir:  "Total in " + vdu.video_directory,
		size: total,
	})
	return durations, vdu.get_max_duration_size_len(durations), vdu.get_max_durations_dir_len(durations)
}

func get_args() (string, bool, bool, bool) {
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
	return dir, summarize, sort, reverse
}

func main() {
	dir, summarize, sort, reverse := get_args()
	vdu := NewVdu(dir)
	durations, max_duration_size_len, max_durations_dir_len := vdu.get_durations(summarize, sort, reverse)
	for i, item := range durations {
		fmt.Printf("%*s %s\n", max_duration_size_len, vdu.seconds_to_human(item.size), item.dir)
		if i == len(durations)-2 {
			fmt.Println(strings.Repeat("-", max_duration_size_len+max_durations_dir_len+1))
		}
	}
}
