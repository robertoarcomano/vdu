import ffmpeg
import os
import datetime
import math
import sys
from multiprocessing import Pool, cpu_count


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

    def exists_video_files(self):
        for root, dirs, files in os.walk(self.video_directory):
            for name in files:
                if self.is_video_file(name):
                    return True
        return False

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
        
    def show_durations(self):
        durations = []
        main_dirs = next(os.walk(self.video_directory))[1]
        total = 0

        for dir in main_dirs:
            sum = self.get_directory_duration(dir)
            if sum:
                durations.append({"sum": self.seconds_to_human(sum), "dir": self.video_directory+dir})
                total += sum
        
        sum = self.get_directory_duration(self.video_directory, files_only=True)
        total += sum
        durations.append({"sum": self.seconds_to_human(total), "dir": "Total in "+self.video_directory})
        return durations        
