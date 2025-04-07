#!/usr/bin/env python3

"""
This script just simulates log delivery to stdout
"""

import sys
from time import sleep

if len(sys.argv) != 2:
    print('invalid arguments. please provide only the log file path')
    exit(1)

filepath = sys.argv[1]

with open(filepath, 'r') as file:
    lines = file.readlines()

    for line in lines:
        try:
            sys.stdout.write(line)
            sleep(0.1)
            sys.stdout.flush()
        except KeyboardInterrupt:
            break

    file.close()
