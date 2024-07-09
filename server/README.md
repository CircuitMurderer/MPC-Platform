# 验证服务器使用说明

首先，确保`server`，`sharer`，`verifier`在同一目录下。

然后，输入以下命令启动服务器：
```shell
./server [port] [threads]
```
其中，port参数为服务器监听的端口，默认为8080；threads参数为程序使用的最大线程数量，默认为2。比如，要在9000端口启动且使用16个线程：
```shell
./server 9000 16
```

接着就可以使用http请求来与服务器进行互动。服务器提供下面四个Restful API：
```
POST  /update   ARGS: id=[calculate_id], party=[file_for_which_party], file=@[file_to_update]
GET   /verify   ARGS: id=[calculate_id], operate=[calculate_operation], scale=[precision_control], workers=[workers]
GET   /result   ARGS: id=[calculate_id]
GET   /delete   ARGS: id=[calculate_id]
```
其中：
- calculate_id：表示当前计算ID，系统会根据ID来区分计算的源数据。
- file_to_update：要上传的文件，以表单形式（Form files）提交。
- file_for_which_party：表示上传的文件是属于哪一方的，此参数可以为Alice，Bob和Result。
- calculate_operation：表示运算操作，可选值为0-ADD，1-SUB，2-MUL，3-DIV，4-CHEAPADD，5-CHEAPDIV，6-EXP
- precision_control：验证精度控制。因本系统使用32位浮点数进行验证，故源数据为64位时会有精度损失，这种情况会发生在减法中因前导0过多而导致的计算误差增大。32位小数一般可以精确计算7位左右有效数字，故此值控制验证时将验证误差小于`precision_control * 1e-6`的项标记为验证成功，默认值为1。比如将其设为10时，表示验证差值小于1e-5也可认定为true。
- workers：表示使用多少批次来分割原始数据，从而进行并行化计算，默认为1即不进行并行化。注意，因计算是分为两方进行的，每一个计算任务都会占用两个线程，所以这里的分割批次要小于等于前面设定的线程数量的一半。比如运行时设置threads为16，则workers最大可为8。

一次完整的验算步骤类似下面这样：
```shell
# Upload files
curl -F "id=1" -F "party=Alice" -F "file=@dataforAlice.csv" http://localhost:8080/update 
curl -F "id=1" -F "party=Bob" -F "file=@dataforBob.csv" http://localhost:8080/update 
curl -F "id=1" -F "party=Result" -F "file=@dataDived.csv" http://localhost:8080/update 

# Verify result
curl "http://localhost:8080/verify?id=1&operate=3"
# curl "http://localhost:8080/verify?id=1&operate=3&workers=8"
# curl "http://localhost:8080/verify?id=1&operate=3&scale=10&workers=8"

# Get checked result
curl -o result_1.csv "http://localhost:8080/result?id=1"

# Delete files
curl "http://localhost:8080/delete?id=1"
```

`/update`输出示例：
```json
{
    "filePath": "data/1/AliceData.csv",         // uploaded file path
    "message": "file uploaded successfully"     // run message
}
```

`/verify`输出示例：
```json
{
    "checked_errors": 0,                    // checked errors of uploaded result
    "share_info": {
        "error_alice": "",                  // sharer error
        "error_bob": "",
        "exitcode_alice": 0,                // sharer exitcode
        "exitcode_bob": 0,
        "output_alice": {                   // sharer output
            "comm_cost": "0 bytes",         // communication cost, in sharer, it's 0
            "total_time": "10.771 ms"       // total run time
        },
        "output_bob": {
            "comm_cost": "0 bytes",
            "total_time": "9.271 ms"
        }
    },
    "verify_info": {                    
        "error_alice": "",                  // verifier error
        "error_bob": "",
        "exitcode_alice": 0,                // verifier exitcode
        "exitcode_bob": 0,
        "output_alice": {                   // verifier output
            "comm_cost": "31246668 bytes",  // communication cost
            "total_time": "1434.41 ms"      // total run time
        },
        "output_bob": {
            "comm_cost": "65677428 bytes",
            "total_time": "1284.06 ms"
        }
    }
}
```

`/result`输出示例：
```
  % Total    % Received % Xferd  Average Speed   Time    Time     Time  Current
                                 Dload  Upload   Total   Spent    Left  Speed
100 98906  100 98906    0     0  18.8M      0 --:--:-- --:--:-- --:--:-- 18.8M
# Download a file
```

`/delete`输出示例：
```json
{
    "dirpath": "data/1",                // deleted directory
    "message": "deleted successfully"   // run message
}
```

若有错误，则返回的json中必有以下字段：
```json
{
    "error": "Error! ..."       // error message
}
```

当然，这只是一个shell的示例。实际使用中，可以使用GET和POST请求的写法都可以接受，比如使用Python：
```Python
import requests

# Upload files
files = { 'file': open('dataforAlice.csv', 'rb') }
response = requests.post("http://localhost:8080/update", data={'id': '1', 'party': 'Alice'}, files=files)
print(response.text)

files = { 'file': open('dataforBob.csv', 'rb') }
response = requests.post("http://localhost:8080/update", data={'id': '1', 'party': 'Bob'}, files=files)
print(response.text)

files = { 'file': open('dataDived.csv', 'rb') }
response = requests.post("http://localhost:8080/update", data={'id': '1', 'party': 'Result'}, files=files)
print(response.text)

# Verify result
response = requests.get("http://localhost:8080/verify", params={'id': '1', 'operate': '3'})
print(response.text)

# Get checked result
response = requests.get("http://localhost:8080/result", params={'id': '1'})
with open('result_1.csv', 'wb') as f:
    f.write(response.content)

# Delete files
response = requests.get("http://localhost:8080/delete", params={'id': '1'})
print(response.text)
```
