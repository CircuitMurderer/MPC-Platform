import React, { useState } from 'react';
import { Segmented } from 'antd';
import { 
  CodeOutlined, 
  UnorderedListOutlined 
} from '@ant-design/icons';
import Compute from './Compute';
import Storage from './Storage';

const SecretsSharing: React.FC = () => {
  const [activeView, setActiveView] = useState<string | number>('console');

  return (
    <div style={{ height: '100%', display: 'flex', flexDirection: 'column' }}>
      
      {/* 顶部切换区：简洁的胶囊按钮 */}
      <div style={{ display: 'flex', justifyContent: 'left', marginBottom: 16 }}>
        <Segmented
          size="large"
          value={activeView}
          onChange={(value) => setActiveView(value)}
          options={[
            {
              label: (
                <div style={{ padding: '4px 8px', fontSize: 16 }}>
                  <CodeOutlined /> 控制台
                </div>
              ),
              value: 'console',
            },
            {
              label: (
                <div style={{ padding: '4px 8px', fontSize: 16 }}>
                  <UnorderedListOutlined /> 任务列表
                </div>
              ),
              value: 'list',
            },
          ]}
        />
      </div>

      {/* 内容显示区：直接条件渲染，无动画 */}
      <div style={{ flex: 1 }}>
        {activeView === 'console' ? (
           <Compute />
        ) : (
           <Storage />
        )}
      </div>
    </div>
  );
};

export default SecretsSharing;