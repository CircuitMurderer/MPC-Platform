import React, { useState } from "react";
import { Button, Row, Col, Form, InputNumber, Select, Typography, Input, Upload, Card } from "antd";
import axios from "axios";
import { InboxOutlined } from "@ant-design/icons";
import { CKKS_BASIC_URI } from "../config";

const { Title } = Typography;
const { TextArea } = Input;

const BASIC_URI = CKKS_BASIC_URI;

const Ckks: React.FC = () => {
  // 表格数据
  const [logs, setLogs] = useState<string[]>([]);

  // 表单数据
  const [datasetSize, setDatasetSize] = useState<number>(1000000);
  const [operationCount, setOperationCount] = useState<number>(1000000);
  const [operationForm, setOperationForm] = useState<string>("-a");

  // 上传文件（空接口）
  const handleUpload = async (file: any) => {
    const formData = new FormData();
    formData.append("file", file);
    formData.append("dataset_size", datasetSize.toString());

    try {
      const res = await axios.post(`${BASIC_URI}/ckksv/upload`, formData, {
        headers: { "Content-Type": "multipart/form-data" },
      });
      setLogs(prevLogs => [...prevLogs, `文件上传成功: ${file.name}`]);
    } catch (error) {
      setLogs(prevLogs => [...prevLogs, `文件上传失败: ${file.name}`]);
    }

    return false; // 防止自动上传
  };

  // 清除日志
  const handleClearLogs = () => {
    setLogs([]);
  };

  // 生成随机数据集
  const handleGenerateDataset = async () => {
    try {
      const res = await axios.post(`${BASIC_URI}/ckksv/generate`, {
        "dataset size": datasetSize,
      });
      setLogs(prevLogs => [...prevLogs, `生成数据集成功: \n${res.data["message"]}`]);
    } catch (error) {
      setLogs(prevLogs => [...prevLogs, "生成数据集失败"]);
    }
  };

  // 运算
  const handleCalculate = async () => {
    try {
      const res = await axios.post(`${BASIC_URI}/ckksv/calcuate`, {
        "number of operations": operationCount,
        "operational form": operationForm,
      });
      setLogs(prevLogs => [...prevLogs, `运算成功: \n${res.data["message"]}`]);
    } catch (error) {
      setLogs(prevLogs => [...prevLogs, `运算失败: \n${error}`]);
    }
  };

  // 验证
  const handleVerify = async () => {
    try {
      const res = await axios.post(`${BASIC_URI}/ckksv/verify`, {
        "number of operations": operationCount,
        "operational form": operationForm,
      });
      setLogs(prevLogs => [...prevLogs, `验证成功: \n${res.data["message"]}`]);
    } catch (error) {
      setLogs(prevLogs => [...prevLogs, "验证失败"]);
    }
  };

  // 下载数据
  const handleDownloadData = async () => {
    try {
      const res = await axios.get(`${BASIC_URI}/ckksv/download_data`, {
        params: { dataset_size: datasetSize },
        responseType: "blob",
      });

      const url = window.URL.createObjectURL(new Blob([res.data]));
      const link = document.createElement("a");
      link.href = url;
      link.setAttribute("download", `data${datasetSize}.csv`);
      document.body.appendChild(link);
      link.click();
      setLogs(prevLogs => [...prevLogs, "数据下载成功"]);
    } catch (error) {
      setLogs(prevLogs => [...prevLogs, "数据下载失败"]);
    }
  };

  // 下载结果
  const handleDownloadResults = async () => {
    try {
      const res = await axios.get(`${BASIC_URI}/ckksv/download_results`, {
        responseType: "blob",
      });

      const url = window.URL.createObjectURL(new Blob([res.data]));
      const link = document.createElement("a");
      link.href = url;
      link.setAttribute("download", "logs.txt");
      document.body.appendChild(link);
      link.click();
      setLogs(prevLogs => [...prevLogs, "结果下载成功"]);
    } catch (error) {
      setLogs(prevLogs => [...prevLogs, "结果下载失败"]);
    }
  };

  return (
    <div style={{ padding: "20px" }}>
      <Row gutter={16}>
        {/* 第一行 - 上传文件、按钮、日志 */}
        <Col span={6}>
          <Upload.Dragger
            name="file"
            customRequest={(options) => handleUpload(options.file)}
            showUploadList={false}
            multiple={false}
            accept=".csv"
            style={{ height: 20 }} // 设置上传文件框高度
          >
            <p className="ant-upload-drag-icon">
              <InboxOutlined />
            </p>
            <p className="ant-upload-text">点击或拖动来上传数据集 (*.csv)</p>
          </Upload.Dragger>
        </Col>

        <Col span={6}>
          <Card title="操作" style={{ height: "100%" }} bordered={false}>
            <Button type="default" block style={{ marginBottom: 10, height: 55 }} onClick={handleDownloadData}>
              下载数据
            </Button>
            <Button type="default" block style={{ marginBottom: 10, height: 55 }} onClick={handleDownloadResults}>
              下载结果
            </Button>
            <Button type="dashed" block onClick={handleClearLogs} style={{ height: 55 }}>
              清除日志
            </Button>
          </Card>
        </Col>

        <Col span={12}>
          <Card title="验证日志" style={{ height: "100%" }} bordered={false}>
            <TextArea
              rows={8}
              value={logs.join("\n")}
              readOnly
              style={{ width: "100%", height: "100%" }} // 设置 TextArea 高度
            />
          </Card>
        </Col>
      </Row>

      {/* 第二行 - 验证参数设置 */}
      <Row gutter={16} style={{ marginTop: "20px" }}> {/* 增加了上下间距 */}
        <Col span={24}>
          <Card title="验证参数设置" bordered={false}>
            <Row gutter={16}>
              <Col span={8}>
                <Form.Item label="数据集大小:" layout="vertical">
                  <InputNumber
                    min={1}
                    value={datasetSize}
                    onChange={(value) => setDatasetSize(value || 1000000)}
                    style={{ width: "100%" }}
                  />
                </Form.Item>
              </Col>

              <Col span={8}>
                <Form.Item label="操作次数:" layout="vertical">
                  <InputNumber
                    min={1}
                    value={operationCount}
                    onChange={(value) => setOperationCount(value || 1000000)}
                    style={{ width: "100%" }}
                  />
                </Form.Item>
              </Col>

              <Col span={8}>
                <Form.Item label="运算方式:" layout="vertical">
                  <Select
                    value={operationForm}
                    onChange={(value) => setOperationForm(value)}
                    style={{ width: "100%" }}
                  >
                    <Select.Option value="-a">加减运算 (-a)</Select.Option>
                    <Select.Option value="-m">乘除运算 (-m)</Select.Option>
                    <Select.Option value="-e">指数运算 (-e)</Select.Option>
                  </Select>
                </Form.Item>
              </Col>
            </Row>

            <Row gutter={16}>
              <Col span={8}>
                <Button type="dashed" block onClick={handleGenerateDataset}>
                  生成数据集
                </Button>
              </Col>

              <Col span={8}>
                <Button type="default" block onClick={handleCalculate}>
                  开始运算
                </Button>
              </Col>

              <Col span={8}>
                <Button type="primary" block onClick={handleVerify}>
                  开始验证
                </Button>
              </Col>
            </Row>
          </Card>
        </Col>
      </Row>
    </div>
  );
};

export default Ckks;
