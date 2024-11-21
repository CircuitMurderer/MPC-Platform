import React from "react";
import { BrowserRouter as Router, Route, Routes, Link } from "react-router-dom";
import { Layout, Menu, Typography } from "antd";
import Home from "./pages/Home";
import Compute from "./pages/Compute";

const { Header, Content, Footer, Sider } = Layout;
const { Title } = Typography;

const App: React.FC = () => (
  <Router>
    <Layout style={{ minHeight: "100vh" }}>
      <Sider breakpoint="lg" collapsedWidth="80">
        <div style={{ padding: "16px", textAlign: "center" }}>
          <Title level={4} style={{ color: "white", margin: 0 }}>
            MPC Verifier
          </Title>
        </div>
        <Menu theme="dark" mode="inline" defaultSelectedKeys={["1"]}>
          <Menu.Item key="1"><Link to="/">首页</Link></Menu.Item>
          <Menu.Item key="2"><Link to="/compute">计算页面</Link></Menu.Item>
        </Menu>
      </Sider>
      
      <Layout>
        <Header style={{ background: "#fff", padding: "0 16px" }}>
          <Title level={4} style={{ margin: 0 }}> </Title>
        </Header>
        <Content style={{ margin: "16px", padding: "16px", background: "#fff" }}>
          <Routes>
            <Route path="/" element={<Home />} />
            <Route path="/compute" element={<Compute />} />
          </Routes>
        </Content>
        <Footer style={{ textAlign: "center" }}>
          MPC Verifier ©2024
        </Footer>
      </Layout>
    </Layout>
  </Router>
);

export default App;
