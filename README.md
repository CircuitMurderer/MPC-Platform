# MPC Platform

The Multi-Party Computation Verify Platform.

## Build & Start

### Prepare

Ensure that you have downloaded the `server`, `sharer`, and `verifier` from the release and placed them in the `run-dir` directory. We have provided the compiled programs for the Linux amd64 system in the release. If your platform is different, you can also enter the `components` folder to compile these three executable programs separately and place them in the `run-dir` directory.

If you want to compile them by yourself, please make sure that you have installed `go 1.21+`, `gcc 11+`, and `cmake 3.16+`, as well as the SCI library of EzPC. For the compilation of the SCI library, please refer to the [official manual](https://github.com/mpc-msri/EzPC/tree/master/SCI#readme). After compiling the SCI library, please set the following two lines in `components/shr_based/CMakeLists.txt` to the address of your compiled SCI library:
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

Next, please ensure that you have installed `python` and `pip` (or a `conda` virtual environment), and the `python` version should be `3.10+`; at the same time, `nodejs` and `yarn` are also necessary, and the `nodejs` version should be `18.16+`. The following commands are used to install the dependencies of `python` and `nodejs`:
```shell
cd app-front-end
yarn

cd app-back-end
pip install -r requirements.txt
```
Then the preparation is finished.

### Run

**Front-End.** The front-end of this application is developed based on [React](https://github.com/facebook/react) and [Antd](https://github.com/ant-design/ant-design). To start the front-end, open a new terminal screen (you can use tools like `screen` or `tmux` in a Linux environment), and then execute the following:
```shell
# In a new screen
cd app-front-end
yarn start
```

**Back-End.** The back-end of this application is powered by [FastAPI](https://github.com/fastapi/fastapi) and [Uvicorn](https://github.com/encode/uvicorn). To start the back-end, open another new terminal screen, and then execute the following:
```shell
# In a new screen
cd app-back-end
python main.py
```

**Scheduler.** The scheduler is built using [Gin](https://github.com/gin-gonic/gin). To start the scheduler, open another new terminal screen, and then execute the following:
```shell
# In a new screen
cd run-dir
./server
```

**Finishing.** If these components are not all deployed on the local machine, you can modify the scheduler IP in `app-back-end/main.py` and the back-end IP in `app-front-end/src/config.ts`. After making these adjustments, you can access the system page by navigating to `[front-end IP]:3000` in your browser. If the port 3000 is occupied, the application will be accessible at `[front-end IP]:3001`.

**Interaction.** After opening the browser, you can select the desired function from the menu bar on the left side of the webpage. The system provides verification functions for secret sharing computation and homomorphic encryption computation. Simply upload the original data file, specify the parameters, and click the **VERIFY** button to submit the verification task in the background. You can check the verification status at any time using the **STATUS** button and obtain the verification results through the **RESULTS** button once the process is completed.

## Test Outlines

You can certainly use the script to test the system, including the back-end and the scheduler. Following the stage like this:
```shell
#!/bin/bash

for N in 1000000 10000000 100000000
do
    echo "Running with data length: $N"

    python ../scripts/gen.py -n $N -f partyA_data.csv
    python ../scripts/look.py -f partyA_data.csv
    python ../scripts/look.py -f partyA_data.csv -l 10

    python ../scripts/gen.py -n $N -f partyB_data.csv
    python ../scripts/look.py -f partyB_data.csv
    python ../scripts/look.py -f partyB_data.csv -l 10

    for O in add sub mul div exp
    do
        echo "Running with operation: $O"

        python ../scripts/cal.py -a partyA_data.csv -b partyB_data.csv -f result_data.csv -o $O
        python ../scripts/look.py -f result_data.csv
        python ../scripts/look.py -f result_data.csv -l 10

        python ../scripts/run.py -a partyA_data.csv -b partyB_data.csv -r result_data.csv -c checked_result.csv -o $O -n 0 -w 8
        wc -l checked_result.csv
        head -n 10 checked_result.csv

        python ../scripts/break.py -f result_data.csv -b 0.1
        python ../scripts/run.py -a partyA_data.csv -b partyB_data.csv -r result_data.csv -c checked_result.csv -o $O -n 0 -w 8
        wc -l checked_result.csv
        head -n 10 checked_result.csv
    done
done
```
The commands can also be found in `scripts/tldr.sh`.
