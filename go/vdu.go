package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/fs"
	"math"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"

	ffmpeg "github.com/u2takey/ffmpeg-go"
)

const __version__ string = "0.2.1"
const __video_extensions__ string = "mp4,mkv,avi,mov,flv,wmv,webm,m4v"

type dir_size struct {
	dir  string
	size float64
}

type finished_type struct {
	duration  float64
	directory string
}

type filesQ_type struct {
	file      string
	directory string
}

type Vdu struct {
	file_formats    string
	video_directory string
	filesQ          chan filesQ_type
	finished        chan finished_type
	workersWaiting  sync.WaitGroup
}

func NewVdu(video_directory string) *Vdu {
	return &Vdu{
		file_formats:    __video_extensions__,
		video_directory: video_directory,
	}
}

func (vdu *Vdu) get_num_files() int {
	count := 0
	filepath.WalkDir(vdu.video_directory, func(path string, d fs.DirEntry, err error) error {
		if !d.IsDir() && vdu.is_video_file(d.Name()) {
			count++
		}
		return nil
	})
	return count
}

func (vdu *Vdu) postInitialize() {
	vdu.filesQ = make(chan filesQ_type, vdu.get_num_files())
	vdu.finished = make(chan finished_type, vdu.get_num_files())
}

func (vdu *Vdu) get_extension(filename string) string {
	dot_position := strings.LastIndex(filename, ".")
	if dot_position != -1 {
		return filename[dot_position+1:]
	}
	return ""
}

func (vdu *Vdu) is_video_file(filename string) bool {
	extension := vdu.get_extension(filename)
	if extension == "" {
		return false
	}
	return strings.Contains(vdu.file_formats, extension)
}
func (vdu *Vdu) seconds_to_human(seconds float64) string {
	secondsInt := int(math.Round(seconds))

	h := secondsInt / 3600
	secondsInt -= h * 3600

	m := secondsInt / 60
	secondsInt -= m * 60

	s := secondsInt

	return fmt.Sprintf("%dh %02dm %02ds", h, m, s)
}

func (vdu *Vdu) get_video_files(dir string, files_only bool) []string {
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

func (vdu *Vdu) get_max_durations_dir_len(durations []dir_size) int {
	max_durations_dir_len := 0
	for _, duration := range durations {
		if len(duration.dir) > max_durations_dir_len {
			max_durations_dir_len = len(duration.dir)
		}
	}
	return max_durations_dir_len
}

func (vdu *Vdu) get_max_duration_size_len(durations []dir_size) int {
	max_duration_size_len := 0
	for _, duration := range durations {
		if len(vdu.seconds_to_human(duration.size)) > max_duration_size_len {
			max_duration_size_len = len(vdu.seconds_to_human(duration.size))
		}
	}
	return max_duration_size_len
}

func (vdu *Vdu) get_durations(isSummarized bool, isSorted bool, isReversed bool) ([]dir_size, int, int) {
	get_duration := func() {
		type probeData struct {
			Format struct {
				Duration string `json:"duration"`
			} `json:"format"`
		}
		defer vdu.workersWaiting.Done()
		var data probeData
		for fileQ := range vdu.filesQ {
			raw, _ := ffmpeg.Probe(fileQ.file)
			json.Unmarshal([]byte(raw), &data)
			duration, _ := strconv.ParseFloat(data.Format.Duration, 64)
			vdu.finished <- finished_type{duration: duration, directory: fileQ.directory}
		}
	}
	for w := 0; w < runtime.NumCPU(); w++ {
		vdu.workersWaiting.Add(1)
		go get_duration()
	}
	if isSummarized {
		main_dirs, _ := os.ReadDir(vdu.video_directory)
		for _, dir := range main_dirs {
			if dir.IsDir() == false {
				continue
			}
			full_directory_name := filepath.Join(vdu.video_directory, dir.Name())
			files := vdu.get_video_files(full_directory_name, false)
			for _, file := range files {
				vdu.filesQ <- filesQ_type{file: file, directory: full_directory_name}
			}
		}
	} else {
		filepath.WalkDir(vdu.video_directory, func(path string, d fs.DirEntry, err error) error {
			if d.IsDir() && path != vdu.video_directory {
				files := vdu.get_video_files(path, true)
				for _, file := range files {
					vdu.filesQ <- filesQ_type{file: file, directory: path}
				}
			}
			return nil
		})
	}

	files := vdu.get_video_files(vdu.video_directory, true)
	for _, file := range files {
		vdu.filesQ <- filesQ_type{file: file, directory: vdu.video_directory}
	}
	close(vdu.filesQ)
	vdu.workersWaiting.Wait()
	close(vdu.finished)

	durations_dir := map[string]float64{}
	for finished := range vdu.finished {
		durations_dir[finished.directory] += finished.duration
	}

	total := 0.0
	durations := []dir_size{}
	for dir, sum := range durations_dir {
		durations = append(durations, dir_size{
			dir:  dir,
			size: sum,
		})
		total += sum
	}

	if isSorted {
		sort.Slice(durations, func(i, j int) bool {
			return (durations[i].size < durations[j].size) != isReversed
		})
	}

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
	vdu.postInitialize()
	durations, max_duration_size_len, max_durations_dir_len := vdu.get_durations(summarize, sort, reverse)
	for i, item := range durations {
		fmt.Printf("%*s %s\n", max_duration_size_len, vdu.seconds_to_human(item.size), item.dir)
		if i == len(durations)-2 {
			fmt.Println(strings.Repeat("-", max_duration_size_len+max_durations_dir_len+1))
		}
	}
}
