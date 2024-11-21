import React, { useState, useEffect } from "react";
import { Card, Row, Col } from "antd";

const Home = () => {
  const [summary, setSummary] = useState({
    runtime: Math.random().toFixed(2) + " 秒",
    calculations: Math.floor(Math.random() * 1000),
    errorRate: (Math.random() * 10).toFixed(2) + "%",
  });

  useEffect(() => {
    // 如果后续有接口，可用axios获取数据
    // axios.get("/api/summary").then(res => setSummary(res.data));
  }, []);

  return (
    <Row gutter={16}>
      <Col span={8}>
        <Card title="运行时间" bordered={false}>{summary.runtime}</Card>
      </Col>
      <Col span={8}>
        <Card title="已检测计算" bordered={false}>{summary.calculations}</Card>
      </Col>
      <Col span={8}>
        <Card title="错误比例" bordered={false}>{summary.errorRate}</Card>
      </Col>
    </Row>
  );
};

export default Home;
