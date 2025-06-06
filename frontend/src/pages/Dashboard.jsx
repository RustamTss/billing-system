import {
  CreditCardOutlined,
  DollarOutlined,
  FileTextOutlined,
  UserOutlined
} from '@ant-design/icons'
import { Card, Col, Row, Statistic, Typography } from 'antd'
import React from 'react'
import { useQuery } from 'react-query'
import { dashboardApi } from '../services/api'

const { Title } = Typography

const Dashboard = () => {
  const { data: metrics, isLoading, error } = useQuery(
    'dashboard-metrics', 
    dashboardApi.getMetrics,
    {
      onSuccess: (data) => {
        console.log('Dashboard metrics loaded:', data)
      },
      onError: (error) => {
        console.error('Dashboard metrics error:', error)
      }
    }
  )

  return (
    <div>
      <Title level={2}>Dashboard</Title>
      
      <Row gutter={16}>
        <Col span={6}>
          <Card>
            <Statistic
              title="Active Brokers"
              value={metrics?.data?.active_brokers || 0}
              prefix={<UserOutlined />}
              loading={isLoading}
            />
          </Card>
        </Col>
        <Col span={6}>
          <Card>
            <Statistic
              title="Total Invoices"
              value={metrics?.data?.total_invoices || 0}
              prefix={<FileTextOutlined />}
              loading={isLoading}
            />
          </Card>
        </Col>
        <Col span={6}>
          <Card>
            <Statistic
              title="Total Loads"
              value={metrics?.data?.total_loads || 0}
              prefix={<CreditCardOutlined />}
              loading={isLoading}
            />
          </Card>
        </Col>
        <Col span={6}>
          <Card>
            <Statistic
              title="Total Debt"
              value={metrics?.data?.total_debt || 0}
              prefix={<DollarOutlined />}
              precision={2}
              loading={isLoading}
            />
          </Card>
        </Col>
      </Row>

      <Row gutter={16} style={{ marginTop: 16 }}>
        <Col span={6}>
          <Card title="Pending Invoices" loading={isLoading}>
            <Statistic 
              value={metrics?.data?.pending_invoices || 0} 
              suffix="invoices"
            />
          </Card>
        </Col>
        <Col span={6}>
          <Card title="Overdue Invoices" loading={isLoading}>
            <Statistic 
              value={metrics?.data?.overdue_invoices || 0} 
              suffix="invoices"
              valueStyle={{ color: '#cf1322' }}
            />
          </Card>
        </Col>
        <Col span={6}>
          <Card title="Paid This Month" loading={isLoading}>
            <Statistic 
              value={metrics?.data?.paid_this_month || 0} 
              prefix="$"
              precision={2}
              valueStyle={{ color: '#3f8600' }}
            />
          </Card>
        </Col>
        <Col span={6}>
          <Card title="Completed Loads" loading={isLoading}>
            <Statistic 
              value={metrics?.data?.completed_loads || 0} 
              suffix="loads"
            />
          </Card>
        </Col>
      </Row>
    </div>
  )
}

export default Dashboard 