import ffmpeg
import os
import datetime
import math
import sys
import argparse
from operator import itemgetter
from multiprocessing import Pool, cpu_count
__version__ = "0.2.0"


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
        
    def get_durations(self, isSummarized=False, isSorted=False, isReversed=False):
        durations = []
        total = 0

        if isSummarized:
            main_dirs = next(os.walk(self.video_directory))[1]
            for dir in main_dirs:
                sum = self.get_directory_duration(dir)
                if sum:
                    durations.append({"sum": sum, "dir": self.video_directory+dir})
                    total += sum
        else:
            for root, dirs, files in os.walk(self.video_directory):
                for dir in dirs:
                    full_dir = os.path.join(root, dir)
                    sum = self.get_directory_duration(full_dir, files_only=True)
                    if sum:
                        durations.append({"sum": sum, "dir": full_dir})
                        total += sum

        sum = self.get_directory_duration(self.video_directory, files_only=True)
        total += sum

        if isSorted:
            durations.sort(key=itemgetter("sum"), reverse=isReversed)

        durations.append({"sum": total, "dir": "Total in "+self.video_directory})
        return durations        

    def exists_video_files(self):
        for root, dirs, files in os.walk(self.video_directory):
            if any(self.is_video_file(name) for name in files):
                return True
        return False

    @staticmethod
    def main(argv=None):
        parser = argparse.ArgumentParser(description="Show video files length")      
        parser.add_argument('dir', type=str, default="./", help='directory')
        parser.add_argument('-v', '--version', '-V', action='version', version=f"%(prog)s {__version__}", help='version')
        parser.add_argument('-s', '--summarize', action='store_true', help='summarize')
        parser.add_argument('-S', '--sort', action='store_true', help='sort')
        parser.add_argument('-r', '--reverse', action='store_true', help='reverse the sort')

        args = parser.parse_args()

        if not os.path.isdir(args.dir):
            sys.exit("No directory " + args.dir)

        vdu = Vdu(args.dir)
        if not vdu.exists_video_files():
            sys.exit("No video files in " + vdu.get_video_directory())
        
        durations = vdu.get_durations(args.summarize, args.sort, args.reverse)
        max_sum_duration_len = max(len(vdu.seconds_to_human(d["sum"])) for d in durations)
        for i, duration in enumerate(durations):
            dir = duration.get("dir")
            sum_duration = vdu.seconds_to_human(duration.get("sum"))
            print(f"{sum_duration:>{max_sum_duration_len}}  {dir}")
            if i == len(durations) - 2:
                print("-" * (max_sum_duration_len + 2 + max(len(d["dir"]) for d in durations)))        
            