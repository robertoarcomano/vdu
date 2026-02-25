# vdu - utility du-like to show directory-video lengths
This is a script for getting the full length of all videos in a folder/subfolders.
## Use
### Directory listing
```
(.venv) berto@laptop:~/src/vdu$ find video_test/
video_test/
video_test/2
video_test/2/6
video_test/2/6/video1.mp4
video_test/2/video1 copy.mp4
video_test/2/video2.avi
video_test/2/video1.mp4
video_test/1
video_test/1/video2.avi
video_test/1/video1.mp4
video_test/5
video_test/5/video5.avi
video_test/4
video_test/4/video2 copy.mp4
video_test/4/video1 copy.mp4
video_test/4/video2.avi
video_test/4/video1.mp4
video_test/video2.avi
video_test/non_video
video_test/8
video_test/8/video1.mp4
video_test/3
video_test/3/video3.mp4
video_test/3/video1 copy.mp4
video_test/3/video2.avi
video_test/3/video5.avi
video_test/3/video1.mp4
```

### Sum all the video durations
```
berto@laptop:~/src/vdu$ vdu video_test/
0h 00m 30s video_test/1
0h 00m 51s video_test/2
0h 00m 21s video_test/2/6
0h 01m 22s video_test/3
0h 01m 12s video_test/4
0h 00m 10s video_test/5
0h 00m 21s video_test/8
-------------------------------
0h 04m 57s Total in video_test/
```

## Language notes
This repository includes Go and Python versions, you can specify the language by using the "LANG" Makefile parameter:
```
berto@laptop:~/src/vdu$ make
=== Instructions for Debian package vdu v0.2.1 ===

Available commands:
  make env       - Create .venv and install dependencies (requirements.txt)
  make test      - Run pytest tests
  make build     - Build .deb with dpkg-deb --build vdu-0.2.1
  make install   - Install: sudo apt install ./vdu-0.2.1.deb
  make uninstall - Uninstall: sudo apt remove vdu

Typical workflow:
  1. make env
  2. make test
  3. Prepare vdu-0.2.1/debian/control, rules...
  4. make build
  5. make install

Parameters:
  LANG=<python|go(default)>
```

## Dev notes
This tool is spawing cpu.count() processes in order to exploit the full SMP cpu capability, by means of the Pool.map whereas in Go threads are launched with no regards to the number of cpu/cores

