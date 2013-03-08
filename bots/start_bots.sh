#!/bin/bash

for i in {1..10}; do
    ./bot.py localhost 8945 unit$i unit123 &
done
