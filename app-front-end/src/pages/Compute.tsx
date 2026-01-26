import React, { useState } from "react";
import { 
  Upload, 
  Button, 
  InputNumber, 
  Row, 
  Col, 
  Select, 
  message, 
  Table, 
  Input, 
  Slider, 
  Modal, 
  Skeleton, 
  FloatButton, 
  Tag, 
  Divider,
  Card,      
  Space,     
  Form,      
  Typography,
  Descriptions,
} from "antd";
import { 
  InboxOutlined, 
  InfoCircleOutlined, 
  CheckCircleOutlined, 
  PlusOutlined,
  ClockCircleOutlined,
  SyncOutlined,
  UserOutlined,       
  FileTextOutlined,   
  SettingOutlined,    
  PlayCircleOutlined, 
  DownloadOutlined    
} from "@ant-design/icons";
import axios from "axios";
import type { RcFile } from "antd/es/upload"; 
import { ABY_BASIC_URI } from "../config";

interface SummaryData {
  key: string;
  md5: string;
  items: number;
  mean: number;
  std: number;
  max: number;
  min: number;
  description: string;
}

interface TaskResult {
  data_length: number;
  operate_between: number;
  checked_mistakes: number;
  mistake_rate: string;
  conf_level: string;
  error_rate: string;
  time_cost: string;
  comm_cost: string;
}

interface TaskInfo {
  desc: string;
  sub_stage: string;
}

interface TaskStatus {
  task_id: string;
  task_stat: string;
  task_info: TaskInfo;
  task_result?: TaskResult;
}

const operate_to_text = ['ADD', 'SUB', 'MUL', 'DIV', 'Cheap ADD', 'Cheap DIV', 'EXP'];
const BASIC_URI = ABY_BASIC_URI;
const trans: {[key: string]: string} = {
  "running": "正在验证",
  "completed": "验证完成",
  "failed": "验证失败",
  "unknown": "未知状态"
};

const initialSummary: SummaryData[] = [
  { key: "Alice", md5: "-", items: 0, mean: 0, std: 0, max: 0, min: 0, description: "Alice数据文件 (未上传)" },
  { key: "Bob", md5: "-", items: 0, mean: 0, std: 0, max: 0, min: 0, description: "Bob数据文件 (未上传)" },
  { key: "Result", md5: "-", items: 0, mean: 0, std: 0, max: 0, min: 0, description: "Result数据文件 (未上传)" },
];

const Compute: React.FC = () => {
  const [loading, setLoading] = useState(false); 
  const [visible, setVisible] = useState(false);
  const [statusData, setStatusData] = useState<TaskStatus | undefined>(undefined);
  
  // --- 新增：生成数据相关状态 ---
  const [genVisible, setGenVisible] = useState(false);
  const [genLoading, setGenLoading] = useState(false);
  const [genForm] = Form.useForm();

  const [summary, setSummary] = useState<SummaryData[]>(initialSummary);
  const [verifyParams, setVerifyParams] = useState({
    id: "test",
    operate: 2,
    scale: 1,
    workers: 8,
    split_n: 0,
  });

  const handleUpload = async (party: string, file: RcFile): Promise<void> => {
    const formData = new FormData();
    formData.append("file", file);
    formData.append("party", party);
    formData.append("id", verifyParams.id);

    try {
      const res = await axios.post(`${BASIC_URI}/update`, formData);
      message.success(`${party} 文件上传成功`);
      const updatedData: SummaryData = {
        key: party,
        md5: res.data.md5,
        items: res.data.items,
        mean: res.data.mean,
        std: res.data.std,
        max: res.data.max,
        min: res.data.min,
        description: `${party} 数据文件`,
      };
      setSummary((prev) => prev.map((item) => (item.key === party ? updatedData : item)));
    } catch (error) {
      message.error(`${party} 文件上传失败`);
    }
  };

  const handleVerify = async (): Promise<void> => {
    try {
      const res = await axios.get(`${BASIC_URI}/verify`, { params: verifyParams });
      message.success(`验算任务(ID = ${res.data.task_id})已启动`);
    } catch (error) {
      message.error("验证任务启动失败");
    }
  };

  const handleDownload = async (): Promise<void> => {
    try {
      const res = await axios.get(`${BASIC_URI}/result`, { params: { id: verifyParams.id }, responseType: "blob" });
      const url = window.URL.createObjectURL(new Blob([res.data]));
      const link = document.createElement("a");
      link.href = url;
      link.setAttribute("download", "Verified.csv");
      document.body.appendChild(link);
      link.click();
      message.success("结果下载成功");
    } catch (error) {
      message.error("结果下载失败");
    }
  };

  const handleStatus = async (): Promise<void> => {
    setLoading(true); 
    try {
      const res = await axios.get(`${BASIC_URI}/stat`, { params: { id: verifyParams.id } });
      setStatusData(res.data); 
      console.log(res.data);
    } catch (error) {
      message.error("任务未运行或获取状态失败");
    } finally {
      setLoading(false); 
    }
  };

  const handleGenSubmit = async () => {
    try {
      const values = await genForm.validateFields();
      setGenLoading(true);

      const res = await axios.get(`${BASIC_URI}/generate`, {
        params: {
          id: values.id,
          operator: values.operator,
          data_length: values.data_length,
        }
      });

      message.success(res.data.message || "数据生成成功");

      const { a, b, r } = res.data.summary;
      const mapSummary = (key: string, data: any, desc: string): SummaryData => ({
        key,
        md5: data.md5 || "-",
        items: data.count || values.data_length,
        mean: data.mean || 0,
        std: data.std || 0,
        max: data.max || 0,
        min: data.min || 0,
        description: desc,
      });

      const newSummary = [
        mapSummary("Alice", a, "Alice 数据文件 (生成)"),
        mapSummary("Bob", b, "Bob 数据文件 (生成)"),
        mapSummary("Result", r, "Result 数据文件 (生成)"),
      ];

      setSummary(newSummary);
      
      setVerifyParams(prev => ({ ...prev, id: values.id }));

      setGenVisible(false);
    } catch (error) {
      console.error(error);
      message.error("数据生成失败，请检查参数");
    } finally {
      setGenLoading(false);
    }
  };

  const handleModalOpen = () => {
    setVisible(true);
    handleStatus();
  };

  const handleModalClose = () => {
    setVisible(false);
    setStatusData(undefined);
  }; 

  const columns = [
    { title: "文件描述", dataIndex: "description", key: "description" },
    { title: "MD5", dataIndex: "md5", key: "md5" },
    { title: "条目数", dataIndex: "items", key: "items" },
    { title: "平均值", dataIndex: "mean", key: "mean" },
    { title: "标准差", dataIndex: "std", key: "std" },
    { title: "最大值", dataIndex: "max", key: "max" },
    { title: "最小值", dataIndex: "min", key: "min" },
  ];

  return (
    <div style={{ background: 'transparent' }}>
      <Space direction="vertical" size="middle" style={{ display: 'flex' }}>
        
        <Row gutter={16}>
          {["Alice", "Bob", "Result"].map((party, index) => {
            const icon = party === "Result" ? <FileTextOutlined /> : <UserOutlined />;
            const title = party === "Result" ? "结果文件" : `${party} 端文件`;
            
            return (
              <Col span={8} key={index}>
                <Card title={<Space>{icon} {title}</Space>} size="small" bordered={false}>
                  <Upload.Dragger
                    name="file"
                    customRequest={({ file }) => handleUpload(party, file as RcFile)}
                    showUploadList={false}
                    style={{ background: '#fafafa', border: '1px dashed #d9d9d9' }}
                  >
                    <p className="ant-upload-drag-icon" style={{ marginBottom: 10 }}>
                      <InboxOutlined style={{ fontSize: 28, color: '#1890ff' }} />
                    </p>
                    <p className="ant-upload-text" style={{ fontSize: 14, color: '#666' }}>
                      点击或拖拽上传
                    </p>
                  </Upload.Dragger>
                </Card>
              </Col>
            );
          })}
        </Row>

        <Card title="数据概览" size="small" bordered={false}>
          <Table 
            dataSource={summary} 
            columns={columns} 
            pagination={false} 
            size="middle"
            bordered 
          />
        </Card>

        <Card 
          title={<Space><SettingOutlined /> 验证参数配置</Space>} 
          bordered={false}
          extra={
            <Space>
               <Button onClick={handleModalOpen}>查看状态</Button>
               <Button icon={<DownloadOutlined />} onClick={handleDownload}>下载结果</Button>
               <Button type="primary" icon={<PlayCircleOutlined />} onClick={handleVerify}>
                 开始验证
               </Button>
            </Space>
          }
        >
          <Form layout="vertical">
            <Row gutter={24}>
              <Col span={6}>
                <Form.Item label="计算ID">
                  <Input
                    value={verifyParams.id}
                    onChange={(e) => setVerifyParams({ ...verifyParams, id: e.target.value })}
                    placeholder="输入唯一ID"
                  />
                </Form.Item>
              </Col>
              <Col span={6}>
                <Form.Item label="串行批次 (Split N)">
                  <InputNumber
                    style={{ width: '100%' }}
                    min={0}
                    value={verifyParams.split_n}
                    onChange={(value) => setVerifyParams({ ...verifyParams, split_n: value || 0 })}
                  />
                </Form.Item>
              </Col>
              <Col span={6}>
                <Form.Item label="并行 Workers">
                  <InputNumber
                    style={{ width: '100%' }}
                    min={1}
                    value={verifyParams.workers}
                    onChange={(value) => setVerifyParams({ ...verifyParams, workers: value || 8 })}
                  />
                </Form.Item>
              </Col>
              
              <Col span={6}>
                <Form.Item label="运算操作">
                  <Select
                    value={verifyParams.operate}
                    onChange={(value) => setVerifyParams({ ...verifyParams, operate: value })}
                  >
                    <Select.Option value={0}>ADD (加法)</Select.Option>
                    <Select.Option value={1}>SUB (减法)</Select.Option>
                    <Select.Option value={2}>MUL (乘法)</Select.Option>
                    <Select.Option value={3}>DIV (除法)</Select.Option>
                    <Select.Option value={4}>Cheap ADD</Select.Option>
                    <Select.Option value={5}>Cheap DIV</Select.Option>
                    <Select.Option value={6}>EXP (指数)</Select.Option>
                  </Select>
                </Form.Item>
              </Col>

              <Col span={24}>
                <Form.Item label={`精度控制 (当前: 1e${Math.log10(verifyParams.scale)})`} style={{ marginBottom: 0 }}>
                   <Slider
                     min={0}
                     max={4}
                     step={1}
                     marks={{ 0: '1', 1: '10', 2: '100', 3: '1k', 4: '10k' }}
                     value={Math.log10(verifyParams.scale)}
                     onChange={(value) => setVerifyParams({ ...verifyParams, scale: 10 ** value })}
                   />
                </Form.Item>
              </Col>
            </Row>
          </Form>
        </Card>
      </Space>

      <FloatButton 
        icon={<PlusOutlined />} 
        tooltip="生成测试数据"
        onClick={() => {
          setGenVisible(true);
          genForm.setFieldsValue({
            id: `auto_${Math.floor(Math.random() * 1000)}`,
            operator: 'mul',
            data_length: 10000
          });
        }}
      />

      <Modal
        title={
          <Space>
            <PlayCircleOutlined style={{ color: '#722ed1' }} /> 
            <span>生成测试数据</span>
          </Space>
        }
        open={genVisible}
        onCancel={() => setGenVisible(false)}
        confirmLoading={genLoading}
        onOk={handleGenSubmit}
        okText="生成并加载"
        cancelText="取消"
        width={500}
      >
        <Form form={genForm} layout="vertical" name="gen_data_form">
          <Form.Item
            name="id"
            label="任务 ID (Task ID)"
            rules={[{ required: true, message: '请输入任务 ID' }]}
            help="生成的数据将存储在该 ID 目录下"
          >
            <Input placeholder="例如: test_01" prefix={<SettingOutlined style={{color: '#bfbfbf'}}/>} />
          </Form.Item>

          <Row gutter={16}>
            <Col span={12}>
              <Form.Item
                name="operator"
                label="运算类型 (Operator)"
                rules={[{ required: true, message: '请选择运算类型' }]}
              >
                <Select>
                  <Select.Option value="add">ADD (加法)</Select.Option>
                  <Select.Option value="sub">SUB (减法)</Select.Option>
                  <Select.Option value="mul">MUL (乘法)</Select.Option>
                  <Select.Option value="div">DIV (除法)</Select.Option>
                  <Select.Option value="exp">EXP (指数)</Select.Option>
                </Select>
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item
                name="data_length"
                label="数据长度 (Length)"
                rules={[{ required: true, message: '请输入数据长度' }]}
              >
                <InputNumber<number> 
                  style={{ width: '100%' }} 
                  min={1} 
                  max={10000000}
                  formatter={value => `${value}`.replace(/\B(?=(\d{3})+(?!\d))/g, ',')}
                  parser={(value) => value?.replace(/\$\s?|(,*)/g, '') as unknown as number}
                />
              </Form.Item>
            </Col>
          </Row>
          
          <div style={{ background: '#f5f5f5', padding: '12px', borderRadius: '4px', color: '#666', fontSize: '12px' }}>
            <InfoCircleOutlined style={{ marginRight: 6 }} />
            点击确定后，服务器将自动生成 Alice, Bob 和 Result 的 CSV 文件并计算统计信息。
          </div>
        </Form>
      </Modal>

      <Modal
        title={
          <div style={{ display: 'flex', alignItems: 'center' }}>
            {statusData?.task_stat === 'completed' ? (
              <CheckCircleOutlined style={{ color: '#52c41a', marginRight: 8, fontSize: 18 }} />
            ) : (
              <InfoCircleOutlined style={{ color: '#1890ff', marginRight: 8, fontSize: 18 }} />
            )}
            <span>任务详情</span>
          </div>
        }
        open={visible}
        onCancel={handleModalClose}
        footer={null}
        width={700}
      >
        {loading ? (
          <Skeleton active paragraph={{ rows: 6 }} />
        ) : statusData ? (
          <>
            <Descriptions bordered column={2} size="small">
              <Descriptions.Item label="任务ID" span={2}>
                <code>{statusData.task_id}</code>
              </Descriptions.Item>
              
              <Descriptions.Item label="当前状态">
                {statusData.task_stat === 'completed' ? (
                  <Tag color="success" icon={<CheckCircleOutlined />}>{trans[statusData.task_stat]}</Tag>
                ) : statusData.task_stat === 'running' ? (
                  <Tag color="processing" icon={<SyncOutlined spin />}>{trans[statusData.task_stat]}</Tag>
                ) : (
                  <Tag color="default">{trans[statusData.task_stat]}</Tag>
                )}
              </Descriptions.Item>

              {statusData.task_stat === 'running' && (
                <Descriptions.Item label="验证阶段">
                  {statusData.task_info.desc} <span style={{color: '#8c8c8c'}}>({statusData.task_info.sub_stage})</span>
                </Descriptions.Item>
              )}
            </Descriptions>

            {statusData.task_stat === 'completed' && statusData.task_result && (
              <>
                <Divider orientation="left" orientationMargin="0" style={{ margin: '16px 0' }}>执行结果报告</Divider>
                <Descriptions bordered column={2} size="small" layout="vertical">
                  <Descriptions.Item label="数据规模">{statusData.task_result.data_length}</Descriptions.Item>
                  <Descriptions.Item label="运算类型">{operate_to_text[statusData.task_result.operate_between]}</Descriptions.Item>
                  <Descriptions.Item label="错误数量">
                    <span style={{ color: statusData.task_result.checked_mistakes > 0 ? '#cf1322' : '#3f8600' }}>
                      {statusData.task_result.checked_mistakes}
                    </span>
                  </Descriptions.Item>
                  <Descriptions.Item label="错误比例">{statusData.task_result.mistake_rate}</Descriptions.Item>
                  <Descriptions.Item label="结果置信度">{statusData.task_result.conf_level}</Descriptions.Item>
                  <Descriptions.Item label="结果误差率">{statusData.task_result.error_rate}</Descriptions.Item>
                  <Descriptions.Item label="时间开销">
                    <ClockCircleOutlined style={{ marginRight: 4 }} />
                    {statusData.task_result.time_cost}
                  </Descriptions.Item>
                  <Descriptions.Item label="通信开销">{statusData.task_result.comm_cost}</Descriptions.Item>
                </Descriptions>
              </>
            )}
          </>
        ) : (
          <div style={{ textAlign: 'center', color: '#999', padding: '20px 0' }}>暂无任务状态信息</div>
        )}
      </Modal>
    </div>
  );
};

export default Compute;