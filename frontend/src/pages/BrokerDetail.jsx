import { ArrowLeftOutlined, EditOutlined, UserOutlined } from '@ant-design/icons'
import { Button, Card, Col, Descriptions, Row, Space, Tag, Typography } from 'antd'
import React from 'react'
import { useQuery } from 'react-query'
import { useNavigate, useParams } from 'react-router-dom'
import { brokersApi } from '../services/api'

const { Title, Text } = Typography

const BrokerDetail = () => {
  const { id } = useParams()
  const navigate = useNavigate()

  const { data: broker, isLoading, error } = useQuery(
    ['broker', id],
    () => brokersApi.getById(id),
    {
      enabled: !!id,
    }
  )

  const renderStatus = (status) => {
    const statusConfig = {
      active: { color: 'green', text: 'Active' },
      inactive: { color: 'orange', text: 'Inactive' },
      suspended: { color: 'red', text: 'Suspended' },
    }
    
    const config = statusConfig[status] || { color: 'default', text: status }
    return <Tag color={config.color}>{config.text}</Tag>
  }

  const formatCurrency = (amount) => {
    return amount ? `$${amount.toLocaleString()}` : '$0'
  }

  if (isLoading) {
    return <div>Loading broker details...</div>
  }

  if (error || !broker?.data) {
    return <div>Broker not found</div>
  }

  const brokerData = broker.data

  return (
    <div>
      <div style={{ marginBottom: 24 }}>
        <Space>
          <Button 
            icon={<ArrowLeftOutlined />} 
            onClick={() => navigate('/brokers')}
          >
            Back to Brokers
          </Button>
          <Button 
            type="primary" 
            icon={<EditOutlined />}
            onClick={() => navigate(`/brokers/${id}/edit`)}
          >
            Edit Broker
          </Button>
        </Space>
      </div>

      <Row gutter={24}>
        <Col span={16}>
          <Card>
            <div style={{ marginBottom: 24 }}>
              <Space align="center">
                <UserOutlined style={{ fontSize: 24, color: '#1890ff' }} />
                <div>
                  <Title level={2} style={{ margin: 0 }}>
                    {brokerData.company_name}
                  </Title>
                  <Text type="secondary">Broker Details</Text>
                </div>
              </Space>
            </div>

            <Descriptions bordered column={2}>
              <Descriptions.Item label="Company Name" span={2}>
                {brokerData.company_name}
              </Descriptions.Item>
              
              <Descriptions.Item label="Contact Person">
                {brokerData.contact_person || 'Not specified'}
              </Descriptions.Item>
              
              <Descriptions.Item label="Status">
                {renderStatus(brokerData.status)}
              </Descriptions.Item>
              
              <Descriptions.Item label="Email">
                {brokerData.email}
              </Descriptions.Item>
              
              <Descriptions.Item label="Phone">
                {brokerData.phone || 'Not specified'}
              </Descriptions.Item>
              
              <Descriptions.Item label="Credit Limit">
                {formatCurrency(brokerData.credit_limit)}
              </Descriptions.Item>
              
              <Descriptions.Item label="Reliability Score">
                <Space>
                  <Text strong>{brokerData.reliability_score}/10</Text>
                  <Tag color={brokerData.reliability_score >= 8 ? 'green' : brokerData.reliability_score >= 6 ? 'orange' : 'red'}>
                    {brokerData.reliability_score >= 8 ? 'Excellent' : brokerData.reliability_score >= 6 ? 'Good' : 'Poor'}
                  </Tag>
                </Space>
              </Descriptions.Item>
              
              {brokerData.address && (
                <Descriptions.Item label="Address" span={2}>
                  <div>
                    {brokerData.address.street && <div>{brokerData.address.street}</div>}
                    <div>
                      {brokerData.address.city && `${brokerData.address.city}, `}
                      {brokerData.address.state && `${brokerData.address.state} `}
                      {brokerData.address.zip_code}
                    </div>
                    {brokerData.address.country && <div>{brokerData.address.country}</div>}
                  </div>
                </Descriptions.Item>
              )}
              
              {brokerData.notes && (
                <Descriptions.Item label="Notes" span={2}>
                  {brokerData.notes}
                </Descriptions.Item>
              )}
              
              <Descriptions.Item label="Created">
                {new Date(brokerData.created_at).toLocaleDateString()}
              </Descriptions.Item>
              
              <Descriptions.Item label="Last Updated">
                {new Date(brokerData.updated_at).toLocaleDateString()}
              </Descriptions.Item>
            </Descriptions>
          </Card>
        </Col>

        <Col span={8}>
          <Space direction="vertical" style={{ width: '100%' }} size="large">
            <Card title="Quick Stats" size="small">
              <Descriptions column={1} size="small">
                <Descriptions.Item label="Total Invoices">
                  0
                </Descriptions.Item>
                <Descriptions.Item label="Total Payments">
                  0
                </Descriptions.Item>
                <Descriptions.Item label="Outstanding Balance">
                  $0
                </Descriptions.Item>
                <Descriptions.Item label="Active Loads">
                  0
                </Descriptions.Item>
              </Descriptions>
            </Card>

            <Card title="Recent Activity" size="small">
              <Text type="secondary">No recent activity</Text>
            </Card>
          </Space>
        </Col>
      </Row>
    </div>
  )
}

export default BrokerDetail 