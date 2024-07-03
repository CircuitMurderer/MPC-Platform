# 验证服务器使用说明

首先，确保`server`，`sharer`，`verifier`在同一目录下。

然后，输入以下命令启动服务器：
```shell
./server [port]
```
其中，port参数为服务器监听的端口，默认为8080。

接着就可以使用http请求来与服务器进行互动。服务器提供下面四个Restful API：
```
POST  /update   ARGS: id=[calculate_id], party=[file_for_party], file=@[file_to_update]
GET   /verify   ARGS: id=[calculate_id], operate=[calculate_operation]
GET   /result   ARGS: id=[calculate_id]
GET   /delete   ARGS: id=[calculate_id]
```
其中：
- calculate_id：表示当前计算ID，系统在计算时会根据ID来区分计算的源数据。
- file_for_party：表示上传的文件是属于哪一方的，此参数可以为Alice，Bob和Result。
- file_to_update：要上传的文件，以表单形式提交。
- calculate_operation：表示运算操作，可选值为0-ADD，1-SUB，2-MUL，3-DIV，4-CHEAPADD，5-CHEAPDIV，6-EXP

一次完整的验算步骤类似下面这样：
```shell
# Upload files
curl -F "id=1" -F "party=Alice" -F "file=@dataforAlice.csv" http://localhost:8080/update 
curl -F "id=1" -F "party=Bob" -F "file=@dataforBob.csv" http://localhost:8080/update 
curl -F "id=1" -F "party=Result" -F "file=@dataDived.csv" http://localhost:8080/update 

# Verify result
curl "http://localhost:8080/verify?id=1&operate=3"

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
