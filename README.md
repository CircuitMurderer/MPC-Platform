# MPC all in one

The MPC Calculate & Verify Platform.

## Build & Start

### Prepare

Ensure that you have downloaded the `server`, `sharer`, and `verifier` from the release and placed them in the `run-dir` directory. We have provided the compiled programs for the Linux amd64 system in the release. If your platform is different, you can also enter the `components` folder to compile these three executable programs separately and place them in the `run-dir` directory.

If you want to compile them by yourself, please make sure that you have installed `go 1.21+`, `gcc 11+`, and `cmake 3.16+`, as well as the SCI library of EzPC. For the compilation of the SCI library, please refer to the official manual. After compiling the SCI library, please set the following two lines in `components/shr_based/CMakeLists.txt` to the address of your compiled SCI library:
```cmake
set(SCI_DIR "/path/to/SCI/install/dir/lib/cmake/SCI")
include_directories("/path/to/SCI/install/dir/include")
```
Then, execute the following commands to compile the `verifier` and `sharer`:
```shell
cd components/shr_based
mkdir build && cd build
cmake ..
cmake --build .
```
The compiled `verifier` and `sharer` are located in the `components/shr_based/build` folder. Next is the compilation of the scheduler `server`. Execute the following commands:
```shell
cd components/server
go mod tidy
go mod vendor
go build -ldflags="-s -w"
```
The compiled `server` is located in the `components/server` folder. After all the above compilations are completed, place the `server`, `sharer`, and `verifier` in the `run-dir` directory to complete the preparation work of the core programs.

Next, please ensure that you have installed `python` and `pip` (or a `conda` virtual environment), and the `python` version should be `3.10+`; at the same time, `Nodejs` and `yarn` are also necessary, and the `Nodejs` version should be `18.16+`. The following commands are used to install the dependencies of `python` and `nodejs`:
```shell
cd app-front-end
yarn

cd app-back-end
pip install -r requirements.txt
```
Then the preparation is finished.

### Run

Front End:

```shell
# In a new screen
cd app-front-end
yarn start
```

Back End:

```shell
# In a new screen
cd app-back-end
python main.py
```

Schedular:

```shell
# In a new screen
cd run-dir
./server
```

## Test Outlines

```shell
# Generate
# export N=100000000

python ../scripts/gen.py -n $N -f partyA_data.csv
../scripts/look.py -f partyA_data.csv
../scripts/look.py -f partyA_data.csv -l 10

python ../scripts/gen.py -n $N -f partyB_data.csv
../scripts/look.py -f partyB_data.csv
../scripts/look.py -f partyB_data.csv -l 10

# Verify
# export O=mul
## O can be add, sub, mul, div, exp

python ../scripts/cal.py -a partyA_data.csv -b partyB_data.csv -f result_data.csv -o $O
../scripts/look.py -f result_data.csv
../scripts/look.py -f result_data.csv -l 10

python ../scripts/run.py -a partyA_data.csv -b partyB_data.csv -r result_data.csv -c checked_result.csv -o $O -n 0 -w 8
wc -l checked_result.csv
head -n 10 checked_result.csv

../scripts/break.py -f result_data.csv -b 0.1
python ../scripts/run.py -a partyA_data.csv -b partyB_data.csv -r result_data.csv -c checked_result.csv -o $O -n 0 -w 8
wc -l checked_result.csv
head -n 10 checked_result.csv
```
