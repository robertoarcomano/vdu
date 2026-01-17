# vdu - a Python utility du-like to show directory-video lengths
This is Python script for getting the full length of all videos in a folder/subfolders.
## Use
### Directory listing
```
(.venv) berto@laptop:~/src/vdu$ find video_test/
video_test/
video_test/2
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
video_test/3
video_test/3/video3.mp4
video_test/3/video1 copy.mp4
video_test/3/video2.avi
video_test/3/video5.avi
video_test/3/video1.mp4
```

### Sum all the video durations
```
(.venv) berto@laptop:~/src/vdu$ ./vdu ./video_test/
0h 00m 51s  ./video_test/2
0h 00m 30s  ./video_test/1
0h 00m 10s  ./video_test/5
0h 01m 12s  ./video_test/4
0h 01m 22s  ./video_test/3
----------------------------------
0h 04m 15s  Total in ./video_test/
```

## Dev notes
This tool is spawing cpu.count() processes in order to exploit the full SMP cpu capability, by means of the Pool.map

