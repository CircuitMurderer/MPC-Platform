import React from "react";
import { BrowserRouter as Router, Route, Routes, Link, useLocation } from "react-router-dom";
import { Layout, Menu, Typography, Breadcrumb } from "antd";
import Home from "./pages/Home";
import Compute from "./pages/Compute";
import Ckks from "./pages/Ckks";
import { STORAGE_PAGE } from "./config";

const { Header, Content, Footer, Sider } = Layout;
const { Title } = Typography;

const handleStoragePageClick = () => {
  window.location.href = STORAGE_PAGE;
};

const App: React.FC = () => {
  return (
    <Router>
      <Layout style={{ minHeight: "100vh" }}>
        <Sider breakpoint="lg" collapsedWidth="80">
          <div style={{ padding: "16px", textAlign: "center" }}>
            <Title level={5} style={{ color: "white", margin: 0 }}>
              🧮&nbsp;计算验证模块
            </Title>
          </div>
          <SelectMenu />
        </Sider>
        
        <Layout>
          <Header style={{ display: 'flex', justifyContent: 'flex-start', alignItems: 'center', height: '60px', background: "#fff", padding: "0 16px" }}>
            <Title level={4} style={{ margin: 0 }}> </Title>
            <div>
              <DynamicBreadcrumb />
            </div>
          </Header>
          <Content style={{ margin: "16px", padding: "16px", background: "#fff" }}>
            <Routes>
              <Route path="/" element={<Home />} />
              <Route path="/shr" element={<Compute />} />
              <Route path="/fhe" element={<Ckks />} />
            </Routes>
          </Content>
          <Footer style={{ textAlign: "center" }}>
            高置信的密文数据完整性验证工具集 © 2024
          </Footer>
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
