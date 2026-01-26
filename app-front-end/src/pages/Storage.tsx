import React, { useState, useEffect } from "react";
import { 
  Table, 
  Card, 
  Tag, 
  Button, 
  Space, 
  Descriptions, 
  Typography, 
  Badge, 
  Divider,
  message,
  Tooltip
} from "antd";
import { 
  ReloadOutlined, 
  ClockCircleOutlined, 
  CheckCircleOutlined, 
  CloseCircleOutlined, 
  SyncOutlined,
  DatabaseOutlined,
  DeploymentUnitOutlined
} from "@ant-design/icons";
import axios from "axios";
import { ABY_BASIC_URI } from "../config";

const { Title, Text } = Typography;
const BASIC_URI = ABY_BASIC_URI;
const operate_to_text = ['ADD', 'SUB', 'MUL', 'DIV', 'Cheap ADD', 'Cheap DIV', 'EXP'];

interface SummaryStats {
  md5: string;
  items: number;
  mean: number;
  std: number;
  max: number;
  min: number;
}

interface PartyData {
  summary: SummaryStats;
}

interface CheckedResult {
  status: string;
  data_length: number;
  operate_between: number;
  checked_mistakes: string; // 后端返回的是字符串 "0"
  mistake_rate: string;
  conf_level: string;
  error_rate: string;
  comm_cost: string;
  time_cost: string;
}

interface TaskInfo {
  desc: string;
  sub_stage: string;
}

interface TaskItem {
  id: string; 
  status: string;
  stage: string;
  length: number;
  info: TaskInfo;
  Alice?: PartyData;
  Bob?: PartyData;
  Result?: PartyData;
  checked?: CheckedResult; 
}

interface ApiResponse {
  task_order: string[];
  tasks: Record<string, Omit<TaskItem, 'id'>>;
}

const Storage: React.FC = () => {
  const [loading, setLoading] = useState(false);
  const [data, setData] = useState<TaskItem[]>([]);

  const fetchData = async (showToast: boolean = false) => { // 默认为 false
    setLoading(true);
    try {
      const res = await axios.get<ApiResponse>(`${BASIC_URI}/all`);
      const { task_order, tasks } = res.data;

      const orderedList = task_order.map((id) => ({
        id,
        ...tasks[id],
      }));

      setData(orderedList);

      if (showToast) {
        message.success("历史记录已更新");
      }
    } catch (error) {
      console.error(error);
      message.error("获取任务列表失败");
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchData(false); 
  }, []);

  const columns = [
    {
      title: '任务 ID',
      dataIndex: 'id',
      key: 'id',
      render: (text: string) => <Text strong copyable>{text}</Text>,
      width: 150,
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      width: 120,
      render: (status: string) => {
        let color = 'default';
        let icon = null;
        let text = '未知';

        if (status === 'completed') {
          color = 'success';
          icon = <CheckCircleOutlined />;
          text = '已完成';
        } else if (status === 'running') {
          color = 'processing';
          icon = <SyncOutlined spin />;
          text = '进行中';
        } else if (status === 'failed') {
          color = 'error';
          icon = <CloseCircleOutlined />;
          text = '失败';
        }

        return <Tag icon={icon} color={color}>{text}</Tag>;
      },
    },
    {
      title: '数据量',
      dataIndex: 'length',
      key: 'length',
      render: (val: number) => (
        <Space>
          <DatabaseOutlined style={{ color: '#1890ff' }} />
          {val.toLocaleString()}
        </Space>
      ),
    },
    {
      title: '验证耗时',
      key: 'time',
      render: (_: any, record: TaskItem) => {
        if (!record.checked) return '-';
        return (
          <Space>
            <ClockCircleOutlined style={{ color: '#faad14' }} />
            {record.checked.time_cost}
          </Space>
        );
      },
    },
    {
      title: '错误数',
      key: 'mistakes',
      render: (_: any, record: TaskItem) => {
        if (!record.checked) return '-';
        const mistakes = parseInt(record.checked.checked_mistakes);
        return (
          <Badge 
            count={mistakes} 
            showZero 
            overflowCount={99999} 
            style={{ backgroundColor: mistakes > 0 ? '#ff4d4f' : '#52c41a' }} 
          />
        );
      },
    },
    {
      title: '置信度',
      key: 'conf',
      render: (_: any, record: TaskItem) => record.checked?.conf_level || '-',
    },
  ];

  const expandedRowRender = (record: TaskItem) => {
    return (
      <div style={{ padding: '16px 24px', background: '#f8f9fa', borderRadius: 8 }}>
        {record.checked && (
          <>
            <Descriptions title="验证性能报告" size="small" column={4}>
              <Descriptions.Item label="通信开销">{record.checked.comm_cost}</Descriptions.Item>
              <Descriptions.Item label="错误比例">{record.checked.mistake_rate}</Descriptions.Item>
              <Descriptions.Item label="结果误差率">{record.checked.error_rate}</Descriptions.Item>
              <Descriptions.Item label="运算类型">{operate_to_text[record.checked.operate_between]}</Descriptions.Item>
            </Descriptions>
            <Divider style={{ margin: '12px 0' }} />
          </>
        )}

        <Title level={5} style={{ fontSize: 14, color: '#666', marginBottom: 16 }}>
          <DeploymentUnitOutlined /> 多方数据指纹 (Data Fingerprints)
        </Title>
        
        <Space align="start" size="large" wrap>
            {record.Alice && (
                <Card size="small" title="Alice" style={{ width: 300, borderTop: '3px solid #1890ff' }}>
                    <Descriptions column={1} size="small" bordered>
                        <Descriptions.Item label="MD5"><Text type="secondary" style={{fontSize: 12}} copyable>{record.Alice.summary.md5}</Text></Descriptions.Item>
                        <Descriptions.Item label="均值">{record.Alice.summary.mean}</Descriptions.Item>
                        <Descriptions.Item label="标准差">{record.Alice.summary.std}</Descriptions.Item>
                        <Descriptions.Item label="极值">[{record.Alice.summary.min}, {record.Alice.summary.max}]</Descriptions.Item>
                    </Descriptions>
                </Card>
            )}

            {record.Bob && (
                <Card size="small" title="Bob" style={{ width: 300, borderTop: '3px solid #722ed1' }}>
                    <Descriptions column={1} size="small" bordered>
                        <Descriptions.Item label="MD5"><Text type="secondary" style={{fontSize: 12}} copyable>{record.Bob.summary.md5}</Text></Descriptions.Item>
                        <Descriptions.Item label="均值">{record.Bob.summary.mean}</Descriptions.Item>
                        <Descriptions.Item label="标准差">{record.Bob.summary.std}</Descriptions.Item>
                        <Descriptions.Item label="极值">[{record.Bob.summary.min}, {record.Bob.summary.max}]</Descriptions.Item>
                    </Descriptions>
                </Card>
            )}

            {record.Result && (
                <Card size="small" title="Result (计算结果)" style={{ width: 300, borderTop: '3px solid #52c41a' }}>
                    <Descriptions column={1} size="small" bordered>
                        <Descriptions.Item label="MD5"><Text type="secondary" style={{fontSize: 12}} copyable>{record.Result.summary.md5}</Text></Descriptions.Item>
                        <Descriptions.Item label="均值">{record.Result.summary.mean}</Descriptions.Item>
                        <Descriptions.Item label="标准差">{record.Result.summary.std}</Descriptions.Item>
                        <Descriptions.Item label="极值">[{record.Result.summary.min}, {record.Result.summary.max}]</Descriptions.Item>
                    </Descriptions>
                </Card>
            )}
        </Space>
      </div>
    );
  };

  return (
    <div style={{  }}>
      <Card 
        bordered={false} 
        title={<Space><DatabaseOutlined /> 验证任务历史记录</Space>}
        extra={
          <Button 
            type="primary" 
            ghost 
            icon={<ReloadOutlined />} 
            onClick={() => fetchData(true)}
            loading={loading}
          >
            刷新列表
          </Button>
        }
      >
        <Table
          columns={columns}
          dataSource={data}
          rowKey="id"
          loading={loading}
          expandable={{
            expandedRowRender,
            rowExpandable: (record) => record.status === 'completed', 
          }}
          pagination={{ pageSize: 10 }}
        />
      </Card>
    </div>
  );
};

export default Storage;