import {
    DashboardOutlined,
    DollarOutlined,
    FileTextOutlined,
    LogoutOutlined,
    MenuFoldOutlined,
    MenuUnfoldOutlined,
    TruckOutlined,
    UserOutlined,
} from '@ant-design/icons'
import { App as AntdApp, Avatar, Button, Dropdown, Layout, Menu, Space, Typography } from 'antd'
import React, { useState } from 'react'
import { Route, Routes, useLocation, useNavigate } from 'react-router-dom'
import ProtectedRoute from './components/ProtectedRoute'
import { AuthProvider, useAuth } from './contexts/AuthContext'

// Import pages
import BrokerDetail from './pages/BrokerDetail'
import Brokers from './pages/Brokers'
import Dashboard from './pages/Dashboard'
import InvoiceDetail from './pages/InvoiceDetail'
import Invoices from './pages/Invoices'
import Loads from './pages/Loads'
import Login from './pages/Login'
import Payments from './pages/Payments'
import Register from './pages/Register'

const { Header, Sider, Content } = Layout
const { Title } = Typography

// Menu configuration
const menuItems = [
  {
    key: '/',
    icon: <DashboardOutlined />,
    label: 'Dashboard',
  },
  {
    key: '/brokers',
    icon: <UserOutlined />,
    label: 'Brokers',
  },
  {
    key: '/invoices',
    icon: <FileTextOutlined />,
    label: 'Invoices',
  },
  {
    key: '/payments',
    icon: <DollarOutlined />,
    label: 'Payments',
  },
  {
    key: '/loads',
    icon: <TruckOutlined />,
    label: 'Loads',
  },
]

// Main Layout Component (внутри AuthProvider)
const MainLayout = () => {
  const [collapsed, setCollapsed] = useState(false)
  const navigate = useNavigate()
  const location = useLocation()
  const { user, logout } = useAuth()

  const handleMenuClick = ({ key }) => {
    navigate(key)
  }

  const toggleCollapsed = () => {
    setCollapsed(!collapsed)
  }

  const handleLogout = () => {
    logout()
    navigate('/login')
  }

  const userMenuItems = [
    {
      key: 'profile',
      label: 'Профиль',
      icon: <UserOutlined />,
    },
    {
      type: 'divider',
    },
    {
      key: 'logout',
      label: 'Выйти',
      icon: <LogoutOutlined />,
      onClick: handleLogout,
    },
  ]

  return (
    <Layout className="app-layout">
      <Sider 
        className="app-sider"
        trigger={null} 
        collapsible 
        collapsed={collapsed}
        width={250}
        collapsedWidth={80}
      >
        <div style={{ padding: '16px', textAlign: 'center' }}>
          <Title level={collapsed ? 5 : 4} style={{ margin: 0, color: '#1890ff' }}>
            {collapsed ? 'BS' : 'Billing System'}
          </Title>
        </div>
        <Menu
          mode="inline"
          selectedKeys={[location.pathname]}
          items={menuItems}
          onClick={handleMenuClick}
          style={{ borderRight: 0 }}
        />
      </Sider>
      
      <Layout>
        <Header className="app-header">
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', width: '100%' }}>
            <Space>
              <Button
                type="text"
                icon={collapsed ? <MenuUnfoldOutlined /> : <MenuFoldOutlined />}
                onClick={toggleCollapsed}
                style={{ fontSize: '16px' }}
              />
              <Title level={4} style={{ margin: 0 }}>
                Billing Management System
              </Title>
            </Space>
            
            <Space>
              <span style={{ color: '#fff', marginRight: 8 }}>
                Добро пожаловать, {user?.username}!
              </span>
              <Dropdown
                menu={{ items: userMenuItems }}
                placement="bottomRight"
                trigger={['click']}
              >
                <Avatar 
                  style={{ backgroundColor: '#1890ff', cursor: 'pointer' }}
                  icon={<UserOutlined />}
                />
              </Dropdown>
            </Space>
          </div>
        </Header>
        
        <Content className="app-content">
          <Routes>
            <Route path="/" element={<Dashboard />} />
            <Route path="/brokers" element={<Brokers />} />
            <Route path="/brokers/:id" element={<BrokerDetail />} />
            <Route path="/invoices" element={<Invoices />} />
            <Route path="/invoices/:id" element={<InvoiceDetail />} />
            <Route path="/payments" element={<Payments />} />
            <Route path="/loads" element={<Loads />} />
          </Routes>
        </Content>
      </Layout>
    </Layout>
  )
}

// Main App Component
function App() {
  return (
    <AntdApp>
      <AuthProvider>
        <Routes>
          {/* Публичные маршруты */}
          <Route path="/login" element={<Login />} />
          <Route path="/register" element={<Register />} />
          
          {/* Защищенные маршруты */}
          <Route 
            path="/*" 
            element={
              <ProtectedRoute>
                <MainLayout />
              </ProtectedRoute>
            } 
          />
        </Routes>
      </AuthProvider>
    </AntdApp>
  )
}

export default App 