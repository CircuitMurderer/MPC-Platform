import React, { useState } from "react";
import { 
  Upload, Button, InputNumber, Row, Col, Select, message, Table,
  Descriptions, Input, Slider, Modal, Skeleton, FloatButton
} from "antd";
import { InboxOutlined, InfoCircleOutlined, CheckCircleOutlined, PlusOutlined } from "@ant-design/icons";
import axios from "axios";
import { RcFile } from "antd/lib/upload";
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
  
  const [addVisible, setAddVisible] = useState(false);
  const [genParams, setGenParams] = useState({
    operator: "mul",
    data_length: 1000000,
  })

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
      message.success(`验算任务(ID = '${res.data.task_id}')已启动`);
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
      // const status: string = res.data.task_stat;
      // const info = `${res.data.task_info.desc} ${res.data.task_info.sub_stage}`;
      setStatusData(res.data); 
      console.log(res.data);
    } catch (error) {
      message.error("任务未运行");
    } finally {
      setLoading(false); 
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
    <div>
      <Row gutter={[16, 16]}>
        {["Alice", "Bob", "Result"].map((party, index) => (
          <Col span={8} key={index}>
            <Upload.Dragger
              name="file"
              customRequest={({ file }) => handleUpload(party, file as RcFile)}
              showUploadList={false}
              style={{ padding: 16 }}
            >
              <p className="ant-upload-drag-icon">
                <InboxOutlined />
              </p>
              <p className="ant-upload-text">拖动文件到此处，或点击上传 {party} 文件</p>
            </Upload.Dragger>
          </Col>
        ))}

        <Col span={24}>
          <Table dataSource={summary} columns={columns} pagination={false} bordered />
        </Col>

        <Col span={24}>
          <Descriptions title="验证参数配置" bordered column={3}>
            <Descriptions.Item label="计算ID">
              <Input
                value={verifyParams.id}
                onChange={(e) => setVerifyParams({ ...verifyParams, id: e.target.value })}
                style={{ width: "100%" }}
                variant="borderless"
              />
            </Descriptions.Item>
            <Descriptions.Item label="串行批次">
              <InputNumber
                min={0}
                value={verifyParams.split_n}
                onChange={(value) => setVerifyParams({ ...verifyParams, split_n: value || 0 })}
                style={{ width: "100%" }}
                variant="borderless"
              />
            </Descriptions.Item>
            <Descriptions.Item label="并行批次">
              <InputNumber
                min={1}
                value={verifyParams.workers}
                onChange={(value) => setVerifyParams({ ...verifyParams, workers: value || 8 })}
                style={{ width: "100%" }}
                variant="borderless"
              />
            </Descriptions.Item>
            <Descriptions.Item label="运算操作">
              <Select
                value={verifyParams.operate}
                onChange={(value) => setVerifyParams({ ...verifyParams, operate: value })}
                style={{ width: "100%" }}
                variant="borderless"
              >
                <Select.Option value={0}>ADD</Select.Option>
                <Select.Option value={1}>SUB</Select.Option>
                <Select.Option value={2}>MUL</Select.Option>
                <Select.Option value={3}>DIV</Select.Option>
                <Select.Option value={4}>Cheap ADD</Select.Option>
                <Select.Option value={5}>Cheap DIV</Select.Option>
                <Select.Option value={6}>EXP</Select.Option>
              </Select>
            </Descriptions.Item>
            <Descriptions.Item label="精度控制">
              <Slider
                min={0}
                max={4}
                step={1}
                value={Math.log10(verifyParams.scale)}
                tooltip={{ formatter: (value) => value ? 10 ** value : 1 }}
                onChange={(value) => setVerifyParams({ ...verifyParams, scale: 10 ** value })}
              />
            </Descriptions.Item>
            <Descriptions.Item label="操作">
              <Button type="link" onClick={handleVerify}>验证</Button>
              <Button type="text" style={{ marginLeft: 16 }} onClick={handleDownload}>结果</Button>
              <Button type="text" style={{ marginLeft: 16 }} onClick={handleModalOpen}>状态</Button>
            </Descriptions.Item>
          </Descriptions>
        </Col>
      </Row>
      <FloatButton icon={<PlusOutlined />} />

      <Modal
        title={
          <span>{statusData && statusData.task_stat === 'completed' ? (
            <CheckCircleOutlined style={{ marginRight: 8 }} />
          ) : (
            <InfoCircleOutlined style={{ marginRight: 8 }} />
          )}任务状态</span>
        }

        open={visible}
        onCancel={handleModalClose}
        footer={null}
      >
        {loading ? (
          <Skeleton active /> 
        ) : statusData ? (
          <div>
            <p>任务ID: {statusData.task_id}</p>
            <p>状态: {trans[statusData.task_stat]}</p>

            {statusData.task_stat === 'running' && (
              <p>验证阶段: {statusData.task_info.desc} {statusData.task_info.sub_stage}</p>
            )}
            
            {statusData.task_stat === 'completed' && statusData.task_result && (
              <div>
                <p>数据规模: {statusData.task_result.data_length}</p>
                <p>运算类型: {operate_to_text[statusData.task_result.operate_between]}</p>
                <p><strong>错误数量: </strong>{statusData.task_result.checked_mistakes}</p>
                <p><strong>错误比例: </strong>{statusData.task_result.mistake_rate}</p>
                <p><strong>结果置信度: </strong>{statusData.task_result.conf_level}</p>
                <p><strong>结果误差率: </strong>{statusData.task_result.error_rate}</p>
                <p><strong>时间开销: </strong>{statusData.task_result.time_cost}</p>
                <p><strong>通信开销: </strong>{statusData.task_result.comm_cost}</p>
              </div>
            )}
          </div>
        ) : (
          <div>暂无任务状态信息</div> 
        )}
      </Modal>
    </div>
  );
};

export default Compute;
