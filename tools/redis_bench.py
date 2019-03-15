# -*- coding: utf-8 -*-

"""
使用python的redis包来验证发号器的功能
"""
import os
import redis
import time

from multiprocessing import Pool

PORT = 7000
STEP = 100000
NUM_PROCESS=1


def main(pid=0):
    r = redis.Redis(port=PORT)
    pid = pid * 10000 + os.getpid()
    old = int(r.hincrby("test", pid))
    result = []
    s = time.time()
    for i in xrange(STEP):
        result.append(int(r.hincrby("test", pid)))
    ops = STEP / (time.time() - s)
    assert len(result) == len(set(result)), "存在重复的发号"
    assert int(result[-1]) == old + STEP, "存在跳号"
    print "ops:", ops

if __name__ == "__main__":
    pool = Pool(processes=NUM_PROCESS)
    pool.map(main, range(NUM_PROCESS))
