import {
	CarOutlined,
	CreditCardOutlined,
	DashboardOutlined,
	FileTextOutlined,
	UserOutlined,
} from '@ant-design/icons'
import { Layout, Menu } from 'antd'
import React from 'react'
import { Link, useLocation } from 'react-router-dom'

const { Sider } = Layout

const Sidebar = ({ collapsed }) => {
  const location = useLocation()

  const menuItems = [
    {
      key: '/',
      icon: <DashboardOutlined />,
      label: <Link to="/">Dashboard</Link>,
    },
    {
      key: '/brokers',
      icon: <UserOutlined />,
      label: <Link to="/brokers">Brokers</Link>,
    },
    {
      key: '/invoices',
      icon: <FileTextOutlined />,
      label: <Link to="/invoices">Invoices</Link>,
    },
    {
      key: '/payments',
      icon: <CreditCardOutlined />,
      label: <Link to="/payments">Payments</Link>,
    },
    {
      key: '/loads',
      icon: <CarOutlined />,
      label: <Link to="/loads">Loads</Link>,
    },
  ]

  return (
    <Sider trigger={null} collapsible collapsed={collapsed}>
      <div className="demo-logo-vertical" />
      <Menu
        theme="dark"
        mode="inline"
        selectedKeys={[location.pathname]}
        items={menuItems}
      />
    </Sider>
  )
}

export default Sidebar 