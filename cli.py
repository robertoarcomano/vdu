# cli.py
import sys
import os
from vdu import Vdu


def main(argv=None):
    if len(sys.argv) < 2:
        sys.exit("Syntax: " + sys.argv[0] + " <Video Directory>")
    if not os.path.isdir(sys.argv[1]):
        sys.exit("No directory " + sys.argv[1])

    vdu = Vdu(sys.argv[1])
    if not vdu.exists_video_files():
        sys.exit("No video files in " + sys.argv[1])
    
    durations = vdu.show_durations()
    max_dur_len = max(len(d["sum"]) for d in durations)
    for i, duration in enumerate(durations):
        dir = duration.get("dir")
        dur = duration.get("sum")
        print(f"{dur:>{max_dur_len}}  {dir}")
        if i == len(durations) - 2:
            print("-" * (max_dur_len + 2 + max(len(d["dir"]) for d in durations)))        
         