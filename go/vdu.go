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
	fullPath := filepath.Join(vdu.video_directory, dir)
	filepath.WalkDir(fullPath, func(path string, d fs.DirEntry, err error) error {
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
	// files := []string{}
	size := 0.0
	if files_only {
		files, _ := os.ReadDir(vdu.video_directory)
		// fmt.Println(files)
		files = files
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

func (vdu *Vdu) get_durations(isSummarized bool, isSorted bool, isReversed bool) ([]dir_size, int) {
	durations := []dir_size{}
	// total := 0

	if isSummarized {
		main_dirs, _ := os.ReadDir(vdu.video_directory)
		for _, dir := range main_dirs {
			if dir.IsDir() == false && !vdu.is_video_file(dir.Name()) {
				continue
			}
			directory_name := dir.Name()
			sum := vdu.get_directory_duration(directory_name, false)
			// if sum > 0 {
			// 	elem := map[string]interface{}{
			// 		"sum": total,
			// 		"dir": "Total in " + vdu.video_directory,
			// 	}
			// 	durations = append(durations, elem)
			// 	total += sum
			// }
			isSorted = isSorted
			isReversed = isReversed
			fmt.Println(vdu.seconds_to_human(sum), directory_name)
			durations = append(durations, dir_size{
				dir:  directory_name,
				size: sum,
			})
		}
	} // else {
	// 	for root, dirs, files in os.walk(self.video_directory):
	// 	filepath.WalkDir(vdu.dir, func(path string, d fs.DirEntry, err error) error {
	// 		for dir in dirs:
	// 			full_dir = os.path.join(root, dir)
	// 			sum = self.get_directory_duration(full_dir, files_only=True)
	// 			if sum:
	// 				durations.append({"sum": sum, "dir": full_dir})
	// 				total += sum
	// }
	// sum = self.get_directory_duration(vdu.video_directory, files_only=True)
	// total += sum

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
	return durations, vdu.get_max_duration_len(durations)
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

func (vdu Vdu) get_max_duration_len(durations []dir_size) int {
	max_sum_duration_len := 0
	for _, duration := range durations {
		if len(vdu.seconds_to_human(duration.size)) > max_sum_duration_len {
			max_sum_duration_len = len(vdu.seconds_to_human(duration.size))
		}
	}
	return max_sum_duration_len
}

func main() {
	dir, summarize, sort, reverse := get_args()
	vdu := NewVdu(dir)
	durations, max_sum_duration_len := vdu.get_durations(summarize, sort, reverse)

	for _, item := range durations {
		fmt.Printf("%*s %s\n", max_sum_duration_len, vdu.seconds_to_human(item.size), item.dir)
	}
}
