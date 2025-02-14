import React, { useState } from "react";
import { Button, Table, Drawer, Form, InputNumber, Select, Typography, message } from "antd";
import axios from "axios";

const { Title } = Typography;

const Ckks: React.FC = () => {
  // 表格数据
  const [logs, setLogs] = useState([
    {
      key: "1",
      time: "2024-11-27 12:00:00",
      storage: "1.2MB",
      result: "成功",
      id: "12345",
      operation: "加减运算",
    },
    {
      key: "2",
      time: "2024-11-27 13:00:00",
      storage: "2.5MB",
      result: "失败",
      id: "67890",
      operation: "乘除运算",
    },
  ]);
  const [drawerVisible, setDrawerVisible] = useState(false);

  // 表单数据
  const [datasetSize, setDatasetSize] = useState<number>(1000000);
  const [operationCount, setOperationCount] = useState<number>(1000000);
  const [operationForm, setOperationForm] = useState<string>("-a");

  // 表格列定义
  const columns = [
    {
      title: "验证时间",
      dataIndex: "time",
      key: "time",
    },
    {
      title: "存储开销",
      dataIndex: "storage",
      key: "storage",
    },
    {
      title: "验证结果",
      dataIndex: "result",
      key: "result",
    },
    {
      title: "验证ID",
      dataIndex: "id",
      key: "id",
    },
    {
      title: "验证操作",
      dataIndex: "operation",
      key: "operation",
    },
  ];

  // 生成随机数据集
  const handleGenerateDataset = async () => {
    try {
      const res = await axios.post("http://127.0.0.1:9889/ckksv/generate", {
        "dataset size": datasetSize,
      });
      message.success(`生成数据集成功: ${res.data}`);
    } catch (error) {
      message.error("生成数据集失败");
    }
  };

  // 运算
  const handleCalculate = async () => {
    try {
      const res = await axios.post("http://127.0.0.1:9889/ckksv/calcuate", {
        "number of operations": operationCount,
        "operational form": operationForm,
      });
      message.success(`运算成功: ${res.data}`);
    } catch (error) {
      message.error("运算失败");
    }
  };

  // 验证
  const handleVerify = async () => {
    try {
      const res = await axios.post("http://127.0.0.1:9889/ckksv/verify", {
        "number of operations": operationCount,
        "operational form": operationForm,
      });
      message.success(`验证成功: ${res.data}`);
    } catch (error) {
      message.error("验证失败");
    }
  };

  return (
    <div style={{ padding: "20px" }}>
      {/* 表格部分 */}
      <Title level={3}>验证日志</Title>
      <Table dataSource={logs} columns={columns} pagination={{ pageSize: 5 }} />

      {/* 悬浮按钮 */}
      <Button
        type="primary"
        shape="circle"
        icon="+"
        style={{
          position: "fixed",
          bottom: 40,
          right: 40,
          zIndex: 1000,
        }}
        onClick={() => setDrawerVisible(true)}
      />

      {/* 右侧弹出表单 */}
      <Drawer
        title="参数输入"
        placement="right"
        onClose={() => setDrawerVisible(false)}
        visible={drawerVisible}
        width={400}
      >
        <Form layout="vertical">
          {/* 数据集生成 */}
          <Form.Item label="数据集大小:">
            <InputNumber
              min={1}
              value={datasetSize}
              onChange={(value) => setDatasetSize(value || 1000000)}
              style={{ width: "100%" }}
            />
          </Form.Item>
          <Button type="primary" block onClick={handleGenerateDataset}>
            生成数据集
          </Button>

          {/* 运算 */}
          <Form.Item label="操作次数:" style={{ marginTop: "20px" }}>
            <InputNumber
              min={1}
              value={operationCount}
              onChange={(value) => setOperationCount(value || 1000000)}
              style={{ width: "100%" }}
            />
          </Form.Item>
          <Form.Item label="运算方式:">
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
          <Button type="primary" block onClick={handleCalculate}>
            开始运算
          </Button>

          {/* 验证 */}
          <Form.Item label="操作次数:" style={{ marginTop: "20px" }}>
            <InputNumber
              min={1}
              value={operationCount}
              onChange={(value) => setOperationCount(value || 1000000)}
              style={{ width: "100%" }}
            />
          </Form.Item>
          <Form.Item label="运算方式:">
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
          <Button type="primary" block onClick={handleVerify}>
            开始验证
          </Button>
        </Form>
      </Drawer>
    </div>
  );
};

export default Ckks;
