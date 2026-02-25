package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	ffmpeg "github.com/u2takey/ffmpeg-go"
)

const __version__ string = "0.2.0"
const __video_extensions__ string = "mp4,mkv,avi,mov,flv,wmv,webm,m4v"

type dir_size struct {
	dir  string
	size float64
}

type probeData struct {
	Format struct {
		Duration string `json:"duration"`
	} `json:"format"`
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

func get_extension(filename string) string {
	dot_position := strings.LastIndex(filename, ".")
	if dot_position != -1 {
		return filename[dot_position+1:]
	}
	return ""
}

func (vdu Vdu) is_video_file(filename string) bool {
	extension := get_extension(filename)
	if extension == "" {
		return false
	}
	return strings.Contains(vdu.file_formats, get_extension(filename))
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
func (vdu Vdu) get_video_files(dir string) []string {
	video_files := []string{}
	filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		// if err != nil {
		// 	return err
		// }
		// if d.IsDir() {
		// 	return nil
		// }
		if vdu.is_video_file(d.Name()) {
			video_files = append(video_files, path)
		}
		return nil
	})
	return video_files
}
func (vdu Vdu) getVideoDurationSeconds(filePath string) (float64, error) {
	if filePath == "" {
		return 0, errors.New("filePath is empty")
	}

	raw, err := ffmpeg.Probe(filePath)
	if err != nil {
		return 0, err
	}

	var data probeData
	if err := json.Unmarshal([]byte(raw), &data); err != nil {
		return 0, err
	}

	if data.Format.Duration == "" {
		return 0, errors.New("duration not found in metadata")
	}

	duration, err := strconv.ParseFloat(data.Format.Duration, 64)
	if err != nil {
		return 0, err
	}

	return duration, nil
}

func (vdu Vdu) get_directory_duration(dir string, files_only bool) float64 {
	size := 0.0
	if files_only {
		files, _ := os.ReadDir(vdu.video_directory)
		for _, file := range files {
			if file.IsDir() == false && vdu.is_video_file(file.Name()) {
				sub, err := vdu.getVideoDurationSeconds(file.Name())
				size = size + sub
				err = err
			}
		}
	} else {
		files := vdu.get_video_files(dir)
		for _, file := range files {
			sub, err := vdu.getVideoDurationSeconds(file)
			size = size + sub
			err = err
			// fmt.Println(file, sub)
		}
		// Multi process
		// with Pool(processes=cpu_count()) as pool:
		// sums = pool.map(self.get_duration, files)
	}
	return size
	// return sum(sums)
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
			directory_name := dir.Name()
			sum := vdu.get_directory_duration(directory_name, false)
			total += sum
			isSorted = isSorted
			isReversed = isReversed
			durations = append(durations, dir_size{
				dir:  filepath.Join(vdu.video_directory, directory_name),
				size: sum,
			})
		}
	} else {
		filepath.WalkDir(vdu.video_directory, func(path string, d fs.DirEntry, err error) error {
			if d.IsDir() && d.Name() != vdu.video_directory {
				sum := vdu.get_directory_duration(path, false)
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
		isSorted = isSorted
		isReversed = isReversed
	}
	sum := vdu.get_directory_duration(vdu.video_directory, true)
	total += sum
	durations = append(durations, dir_size{
		dir:  "Total in " + vdu.video_directory,
		size: total,
	})

	// if isSorted:
	// durations.sort(key=itemgetter("sum"), reverse=isReversed)

	// durations = durations.append({"sum": total, "dir": "Total in "+self.video_directory})
	// return durations

	// filepath.WalkDir(vdu.dir, func(path string, d fs.DirEntry, err error) error {
	// 	if err != nil {
	// 		return err
	// 	}
	// 	info, _ := d.Info()
	// 	if !info.IsDir() && vdu.is_video_file(path) {
	// 		duration, err := GetVideoDurationSeconds(path)
	// 		if err == nil {
	// 			fmt.Println(path, duration)
	// 		}
	// 	}
	// 	return nil
	// })
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
