#!/bin/bash
set -x

g++ -O3 -std=gnu++20 -ggdb -Wall -Wextra -Wpedantic -Werror -march=native -lncurses -lbenchmark -D_BENCHMARK -o main main.cpp
