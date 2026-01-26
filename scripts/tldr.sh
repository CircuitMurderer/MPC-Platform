#!/bin/bash

for N in 100000000
do
    echo "Running with data length: $N"

    # If not generated
    python ../scripts/gen.py -n $N -f partyA_data.csv
    ../scripts/look.py -f partyA_data.csv
    ../scripts/look.py -f partyA_data.csv -l 10
    read -n 1 -s -r -p "Alice部分数据，数据量$N"
    echo ""

    # If not generated
    python ../scripts/gen.py -n $N -f partyB_data.csv
    ../scripts/look.py -f partyB_data.csv
    ../scripts/look.py -f partyB_data.csv -l 10
    read -n 1 -s -r -p "Bob部分数据，数据量$N"
    echo ""

    for O in add sub mul div exp
    do
        echo "Running with operation: $O"

        python ../scripts/cal.py -a partyA_data.csv -b partyB_data.csv -f result_data.csv -o $O
        ../scripts/look.py -f result_data.csv
        ../scripts/look.py -f result_data.csv -l 10
        read -n 1 -s -r -p "明文计算结果，运算$O，数据量$N"
        echo ""

        python ../scripts/run.py -a partyA_data.csv -b partyB_data.csv -r result_data.csv -c checked_result.csv -o $O -n 0 -w 8
        wc -l checked_result.csv
        head -n 10 checked_result.csv
        
        read -n 1 -s -r -p "密文验证结果（未篡改数据），运算$O，数据量$N"
        echo "" 

        ../scripts/break.py -f result_data.csv -b 0.2
        python ../scripts/run.py -a partyA_data.csv -b partyB_data.csv -r result_data.csv -c checked_result.csv -o $O -n 0 -w 8
        wc -l checked_result.csv
        head -n 10 checked_result.csv

        read -n 1 -s -r -p "密文验证结果（已篡改数据），运算$O，数据量$N"
        echo "" 
    done
done