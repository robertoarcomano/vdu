#!/usr/bin/python3
import pytest
from vdu import Vdu


def test_get_video_length():
    get_video_length = Vdu("video_test")
    assert get_video_length.exists_video_files
    expected = [{'sum': '0h 00m 51s', 'dir': 'video_test/2'}, {'sum': '0h 00m 30s', 'dir': 'video_test/1'}, {'sum': '0h 00m 10s', 'dir': 'video_test/5'}, {'sum': '0h 01m 12s', 'dir': 'video_test/4'}, {'sum': '0h 01m 22s', 'dir': 'video_test/3'}, {'sum': '0h 04m 15s', 'dir': 'Total in video_test/'}]
    rows = get_video_length.show_durations()
    assert rows == expected
