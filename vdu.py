import ffmpeg
import os
import datetime
import math
import sys
import argparse
from multiprocessing import Pool, cpu_count
__version__ = "0.1.6"


class Vdu:
    file_formats = {"mp4", "mkv", "avi", "mov", "flv", "wmv", "webm", "m4v"}
    @staticmethod
    def get_extension(filename):
        return filename.rsplit(".", 1)[-1].lower()

    @staticmethod
    def is_video_file(filename):
        return Vdu.get_extension(filename) in Vdu.file_formats
    
    @staticmethod
    def seconds_to_human(seconds):
        seconds = int(round(seconds))
        h = seconds // 3600
        seconds -= h*3600
        m = seconds // 60
        seconds -= m * 60
        s = int(seconds)
        return f"{h}h {m:02d}m {s:02d}s"

    @staticmethod
    def check_path(path):
        return os.path.join(path, "")

    def __init__(self, video_directory):
        self.video_directory = self.check_path(video_directory)

    def get_video_directory(self):
        return self.video_directory

    def get_video_files(self, dir):
        video_files = []
        for root, dirs, files in os.walk(dir):
            for name in files:
                if self.is_video_file(name):
                    video_files.append(os.path.join(root, name))
        return video_files

    @staticmethod
    def get_duration(file):
        try:
            return float(ffmpeg.probe(file)["format"]["duration"])
        except Exception:
            return 0        

    def get_directory_duration(self, dir, files_only = False):
        files = []
        if files_only:
            files = [os.path.join(dir, f) for f in os.listdir(dir) if os.path.isfile(os.path.join(dir, f)) and self.is_video_file(f)]
        else:
            files = self.get_video_files(self.video_directory+dir)
        # Multi process
        with Pool(processes=cpu_count()) as pool:
            sums = pool.map(self.get_duration, files)
        return sum(sums)
        
    def show_durations_summarize(self):
        durations = []
        total = 0

        main_dirs = next(os.walk(self.video_directory))[1]
        for dir in main_dirs:
            sum = self.get_directory_duration(dir)
            if sum:
                durations.append({"sum": self.seconds_to_human(sum), "dir": self.video_directory+dir})
                total += sum
        
        sum = self.get_directory_duration(self.video_directory, files_only=True)
        total += sum
        durations.append({"sum": self.seconds_to_human(total), "dir": "Total in "+self.video_directory})
        return durations        

    def show_durations(self, summarize=False):
        durations = []
        total = 0

        if summarize:
            main_dirs = next(os.walk(self.video_directory))[1]
            for dir in main_dirs:
                sum = self.get_directory_duration(dir)
                if sum:
                    durations.append({"sum": self.seconds_to_human(sum), "dir": self.video_directory+dir})
                    total += sum
        else:
            for root, dirs, files in os.walk(self.video_directory):
                for dir in dirs:
                    full_dir = os.path.join(root, dir)
                    sum = self.get_directory_duration(full_dir, files_only=True)
                    if sum:
                        durations.append({"sum": self.seconds_to_human(sum), "dir": full_dir})
                        total += sum

        sum = self.get_directory_duration(self.video_directory, files_only=True)
        total += sum
        durations.append({"sum": self.seconds_to_human(total), "dir": "Total in "+self.video_directory})
        return durations        

    def exists_video_files(self):
        for root, dirs, files in os.walk(self.video_directory):
            if any(self.is_video_file(name) for name in files):
                return True
        return False

    @staticmethod
    def main(argv=None):
        parser = argparse.ArgumentParser(description="Show video files length")      
        parser.add_argument('dir', type=str, default="./", help='directory (default=./)')
        parser.add_argument('-v', '--version', '-V', action='version', version=f"%(prog)s {__version__}", help='version')
        parser.add_argument('-s', '--summarize', action='store_true', help='Summarize')

        args = parser.parse_args()

        if not os.path.isdir(args.dir):
            sys.exit("No directory " + args.dir)

        vdu = Vdu(args.dir)
        if not vdu.exists_video_files():
            sys.exit("No video files in " + vdu.get_video_directory())
        
        durations = vdu.show_durations(args.summarize)
        print(durations)
        max_dur_len = max(len(d["sum"]) for d in durations)
        for i, duration in enumerate(durations):
            dir = duration.get("dir")
            dur = duration.get("sum")
            print(f"{dur:>{max_dur_len}}  {dir}")
            if i == len(durations) - 2:
                print("-" * (max_dur_len + 2 + max(len(d["dir"]) for d in durations)))        
            