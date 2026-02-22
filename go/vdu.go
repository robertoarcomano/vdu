package main

import (
	"fmt"
	"strings"
)

const __version__ string = "0.2.0"

type Vdu struct {
	file_formats string
}

func NewVdu() *Vdu {
	return &Vdu{
		file_formats: "mp4,mkv,avi,mov,flv,wmv,webm,m4v",
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

func main() {
	vdu := Vdu{"mp4,mkv,avi,mov,flv,wmv,webm,m4v"}
	fmt.Println(vdu)
}
