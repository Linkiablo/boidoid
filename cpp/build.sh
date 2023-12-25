#!/bin/bash
set -x

g++ -O3 -ggdb -std=gnu++2b -Wall -Wextra -Wpedantic -Werror -march=native -lncurses -o main main.cpp
