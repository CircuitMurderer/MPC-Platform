# MPC all in one

The MPC Calculate & Verify Platform.

## Test Outlines

```shell
#
python ../scripts/gen.py -n N -f partyA_data.csv
../scripts/look.py -f partyA_data.csv
../scripts/look.py -f partyA_data.csv -l 10

python ../scripts/gen.py -n N -f partyB_data.csv
../scripts/look.py -f partyB_data.csv
../scripts/look.py -f partyB_data.csv -l 10

##
python ../scripts/cal.py -a partyA_data.csv -b partyB_data.csv -f result_data.csv -o O
../scripts/look.py -f result_data.csv
../scripts/look.py -f result_data.csv -l 10

python ../scripts/run.py -a partyA_data.csv -b partyB_data.csv -r result_data.csv -c checked_result.csv -o O -n 0 -w 8
wc -l checked_result.csv
head -n 10 checked_result.csv

../scripts/break.py -f result_data.csv -b 0.1
python ../scripts/run.py -a partyA_data.csv -b partyB_data.csv -r result_data.csv -c checked_result.csv -o O -n 0 -w 8
wc -l checked_result.csv
head -n 10 checked_result.csv
```
