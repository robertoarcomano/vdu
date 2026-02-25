#!/usr/bin/python3
import pytest
from vdu import Vdu


def test_get_video_length():
    get_video_length = Vdu("video_test")
    assert get_video_length.exists_video_files
    expected = [{'sum': 51.335066000000005, 'dir': 'video_test/2'}, {'sum': 30.447533, 'dir': 'video_test/1'}, {'sum': 9.56, 'dir': 'video_test/5'}, {'sum': 72.222599, 'dir': 'video_test/4'}, {'sum': 20.887533, 'dir': 'video_test/8'}, {'sum': 81.782599, 'dir': 'video_test/3'}, {'sum': 20.887533, 'dir': 'video_test/2/6'}, {'sum': 296.682863, 'dir': 'Total in video_test/'}]
    rows = get_video_length.get_durations()
    assert rows == expected
