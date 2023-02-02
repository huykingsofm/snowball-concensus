#!/bin/bash
N_PROCESS=${N_PROCESS:-10}
N_ELEMENTS=${N_ELEMENTS:-10}

echo "We will run $N_PROCESS processes"

mkdir -p logs
mkdir -p storage

MIN_PORT=${MIN_PORT:-60000}
MAX_PORT=${MAX_PORT:-$(($MIN_PORT+$N_PROCESS))}

echo "$MIN_PORT $MAX_PORT"

for i in $(seq 1 $N_PROCESS)
do
    ./consensus --n $N_ELEMENTS --folder storage --p $(($i+$MIN_PORT-1)) --minP $MIN_PORT --maxP $MAX_PORT &> logs/consensus_$i.log &
done
