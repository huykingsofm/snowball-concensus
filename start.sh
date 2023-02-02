#!/bin/bash
N_PROCESS=${N_PROCESS:-10}
N_ELEMENTS=${N_ELEMENTS:-10}

echo "We will run $N_PROCESS processes"

mkdir -p logs
mkdir -p storage

MIN_PORT=${MIN_PORT:-60000}
MAX_PORT=${MAX_PORT:-$(($MIN_PORT+$N_PROCESS))}

echo "Start nodes $MIN_PORT..$(($MAX_PORT-1))"
echo "Check logs/consensus_[port].log to see logs for each node"
echo "Ctrl-C to terminate all processes!"

trap_ctrlc() {
    echo "You are stopping all processes. Wait a minute..."
    sleep 1s
    pkill -f consensus
}

trap trap_ctrlc INT

START=$(date +'%s')

for i in $(seq 1 $N_PROCESS)
do
    ./consensus --n $N_ELEMENTS --folder storage --p $(($i+$MIN_PORT-1)) --minP $MIN_PORT --maxP $MAX_PORT &> logs/consensus_$(($i+$MIN_PORT-1)).log &
    pids[${i}]=$!
done

# wait for all pids
for pid in ${pids[*]}; do
    wait $pid
    echo "$pid stopped"
done

echo "Time slapse: $(($(date +'%s') - $START))s"
echo "All nodes stopped!!"
echo "Check final results in storage/[port].result directory"
