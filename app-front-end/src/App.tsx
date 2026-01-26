import React from "react";
import { BrowserRouter as Router, Route, Routes, Link, useLocation } from "react-router-dom";
import { Layout, Menu, Typography, Breadcrumb } from "antd";
import Home from "./pages/Home";
import Ckks from "./pages/Ckks";
import SecretsSharing from "./pages/SecretsSharing";
import { STORAGE_PAGE } from "./config";

const { Header, Content, Footer, Sider } = Layout;
const { Title } = Typography;

const handleStoragePageClick = () => {
  window.location.href = STORAGE_PAGE;
};

const App: React.FC = () => {
  return (
    <Router>
      <style>{`
        body {
          margin: 0;
          padding: 0;
          overflow: hidden; 
        }
        ::-webkit-scrollbar {
          width: 6px;
          height: 6px;
        }
        ::-webkit-scrollbar-thumb {
          background: #ccc;
          border-radius: 3px;
        }
        ::-webkit-scrollbar-track {
          background: transparent;
        }
        .ant-tabs-card > .ant-tabs-nav .ant-tabs-tab {
           border-bottom: none;
        }
      `}</style>

      <Layout style={{ height: "100vh" }}>
        <Sider breakpoint="lg" collapsedWidth="80">
          <div style={{ padding: "16px", textAlign: "center" }}>
            <Title level={5} style={{ color: "white", margin: 0 }}>
              🧮&nbsp;计算验证模块
            </Title>
          </div>
          <SelectMenu />
        </Sider>
        
        <Layout style={{ display: 'flex', flexDirection: 'column' }}>
          
          <Header style={{ display: 'flex', justifyContent: 'flex-start', alignItems: 'center', height: '60px', background: "#fff", padding: "0 16px", boxShadow: '0 1px 4px rgba(0,21,41,.08)', zIndex: 1 }}>
            <Title level={4} style={{ margin: 0 }}> </Title>
            <div>
              <DynamicBreadcrumb />
            </div>
          </Header>
          
          <Content style={{ 
            padding: "16px", 
            background: "transparent", 
            flex: 1, 
            overflowY: "auto",
            minHeight: 0 
          }}>
            <Routes>
              <Route path="/" element={<Home />} />
              <Route path="/shr" element={<SecretsSharing />} />
              <Route path="/fhe" element={<Ckks />} />
            </Routes>
            
            <Footer style={{ textAlign: "center", background: "transparent", padding: "24px 0" }}>
              高置信的密文数据完整性验证工具集 © 2024
            </Footer>
          </Content>
        </Layout>
      </Layout>
    </Router>
  )
};

const DynamicBreadcrumb: React.FC = () => {
  const location = useLocation();

  const getBreadcrumbItems = () => {
    const path = location.pathname;
    switch (path) {
      case "/shr":
        return ["", "计算验证模块", "首页", "秘密共享验证"];
      case "/fhe":
        return ["", "计算验证模块", "首页", "同态加密验证"];
      default:
        return ["", "计算验证模块", "首页"];
    }
  };

  return (
    <Breadcrumb>
      {getBreadcrumbItems().map((item, index) => (
        <Breadcrumb.Item key={index}>{item}</Breadcrumb.Item>
      ))}
    </Breadcrumb>
  );
};

const SelectMenu: React.FC = () => {
  const location = useLocation();

  const getSelectedKey = () => {
    const path = location.pathname;
    if (path === "/shr") {
      return "2";
    } else if (path === "/fhe") {
      return "3";
    } else {
      return "1"; 
    }
  };

  return (
    <Menu theme="dark" mode="inline" selectedKeys={[getSelectedKey()]}>
      <Menu.Item key="1"><Link to="/">首页</Link></Menu.Item>
      <Menu.Item key="2"><Link to="/shr">秘密共享验证</Link></Menu.Item>
      <Menu.Item key="3"><Link to="/fhe">同态加密验证</Link></Menu.Item>
      <Menu.Item key="4" onClick={handleStoragePageClick}>存储页面</Menu.Item>
    </Menu>
  )
}

export default App;