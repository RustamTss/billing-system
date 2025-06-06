import { ArrowLeftOutlined, DollarOutlined, EditOutlined, FileTextOutlined } from '@ant-design/icons'
import { Button, Card, Col, Descriptions, Progress, Row, Space, Tag, Typography } from 'antd'
import dayjs from 'dayjs'
import React from 'react'
import { useQuery } from 'react-query'
import { useNavigate, useParams } from 'react-router-dom'
import { invoicesApi } from '../services/api'

const { Title, Text } = Typography

const InvoiceDetail = () => {
  const { id } = useParams()
  const navigate = useNavigate()

  const { data: invoice, isLoading, error } = useQuery(
    ['invoice', id],
    () => invoicesApi.getById(id),
    {
      enabled: !!id,
    }
  )

  const renderStatus = (status) => {
    const statusConfig = {
      pending: { color: 'orange', text: 'Pending' },
      paid: { color: 'green', text: 'Paid' },
      partial: { color: 'blue', text: 'Partial' },
      overdue: { color: 'red', text: 'Overdue' },
      cancelled: { color: 'default', text: 'Cancelled' },
    }
    
    const config = statusConfig[status] || { color: 'default', text: status }
    return <Tag color={config.color}>{config.text}</Tag>
  }

  const formatCurrency = (amount, currency) => {
    const symbol = currency === 'USD' ? '$' : currency === 'EUR' ? '€' : '₽'
    return `${symbol}${amount?.toLocaleString() || 0}`
  }

  const calculateProgress = (paid, total) => {
    if (!total || total === 0) return 0
    return Math.round((paid / total) * 100)
  }

  if (isLoading) {
    return <div>Loading invoice details...</div>
  }

  if (error || !invoice?.data) {
    return <div>Invoice not found</div>
  }

  const invoiceData = invoice.data
  const progress = calculateProgress(invoiceData.paid_amount || 0, invoiceData.amount)
  const remainingAmount = invoiceData.amount - (invoiceData.paid_amount || 0)

  return (
    <div>
      <div style={{ marginBottom: 24 }}>
        <Space>
          <Button 
            icon={<ArrowLeftOutlined />} 
            onClick={() => navigate('/invoices')}
          >
            Back to Invoices
          </Button>
          <Button 
            type="primary" 
            icon={<EditOutlined />}
            onClick={() => navigate(`/invoices/${id}/edit`)}
          >
            Edit Invoice
          </Button>
        </Space>
      </div>

      <Row gutter={24}>
        <Col span={16}>
          <Card>
            <div style={{ marginBottom: 24 }}>
              <Space align="center">
                <FileTextOutlined style={{ fontSize: 24, color: '#1890ff' }} />
                <div>
                  <Title level={2} style={{ margin: 0 }}>
                    {invoiceData.invoice_number}
                  </Title>
                  <Text type="secondary">Invoice Details</Text>
                </div>
                <div style={{ marginLeft: 'auto' }}>
                  {renderStatus(invoiceData.status)}
                </div>
              </Space>
            </div>

            <Descriptions bordered column={2}>
              <Descriptions.Item label="Invoice Number" span={2}>
                <Text strong>{invoiceData.invoice_number}</Text>
              </Descriptions.Item>
              
              <Descriptions.Item label="Broker">
                {invoiceData.broker_name || 'Unknown Broker'}
              </Descriptions.Item>
              
              <Descriptions.Item label="Status">
                {renderStatus(invoiceData.status)}
              </Descriptions.Item>
              
              <Descriptions.Item label="Amount">
                <Text strong style={{ fontSize: 16 }}>
                  {formatCurrency(invoiceData.amount, invoiceData.currency)}
                </Text>
              </Descriptions.Item>
              
              <Descriptions.Item label="Paid Amount">
                <Text strong style={{ color: '#52c41a' }}>
                  {formatCurrency(invoiceData.paid_amount || 0, invoiceData.currency)}
                </Text>
              </Descriptions.Item>
              
              <Descriptions.Item label="Remaining">
                <Text strong style={{ color: remainingAmount > 0 ? '#ff4d4f' : '#52c41a' }}>
                  {formatCurrency(remainingAmount, invoiceData.currency)}
                </Text>
              </Descriptions.Item>
              
              <Descriptions.Item label="Currency">
                {invoiceData.currency}
              </Descriptions.Item>
              
              <Descriptions.Item label="Issue Date">
                {dayjs(invoiceData.created_at).format('MMM DD, YYYY')}
              </Descriptions.Item>
              
              <Descriptions.Item label="Due Date">
                <Text type={dayjs(invoiceData.due_date).isBefore(dayjs()) ? 'danger' : 'default'}>
                  {dayjs(invoiceData.due_date).format('MMM DD, YYYY')}
                </Text>
              </Descriptions.Item>
              
              {invoiceData.description && (
                <Descriptions.Item label="Description" span={2}>
                  {invoiceData.description}
                </Descriptions.Item>
              )}
              
              {invoiceData.terms && (
                <Descriptions.Item label="Terms" span={2}>
                  {invoiceData.terms}
                </Descriptions.Item>
              )}
            </Descriptions>
          </Card>
        </Col>

        <Col span={8}>
          <Space direction="vertical" style={{ width: '100%' }} size="large">
            <Card title="Payment Progress" size="small">
              <div style={{ marginBottom: 16 }}>
                <Progress 
                  percent={progress} 
                  status={progress === 100 ? 'success' : 'active'}
                  strokeColor={progress === 100 ? '#52c41a' : '#1890ff'}
                />
              </div>
              <Descriptions column={1} size="small">
                <Descriptions.Item label="Total Amount">
                  {formatCurrency(invoiceData.amount, invoiceData.currency)}
                </Descriptions.Item>
                <Descriptions.Item label="Paid">
                  {formatCurrency(invoiceData.paid_amount || 0, invoiceData.currency)}
                </Descriptions.Item>
                <Descriptions.Item label="Remaining">
                  {formatCurrency(remainingAmount, invoiceData.currency)}
                </Descriptions.Item>
              </Descriptions>
            </Card>

            <Card title="Timeline" size="small">
              <div style={{ marginBottom: 8 }}>
                <Text strong>Created:</Text><br />
                <Text type="secondary">
                  {dayjs(invoiceData.created_at).format('MMM DD, YYYY HH:mm')}
                </Text>
              </div>
              
              <div style={{ marginBottom: 8 }}>
                <Text strong>Due Date:</Text><br />
                <Text type={dayjs(invoiceData.due_date).isBefore(dayjs()) ? 'danger' : 'default'}>
                  {dayjs(invoiceData.due_date).format('MMM DD, YYYY')}
                </Text>
              </div>
              
              {invoiceData.updated_at !== invoiceData.created_at && (
                <div>
                  <Text strong>Last Updated:</Text><br />
                  <Text type="secondary">
                    {dayjs(invoiceData.updated_at).format('MMM DD, YYYY HH:mm')}
                  </Text>
                </div>
              )}
            </Card>

            <Card title="Actions" size="small">
              <Space direction="vertical" style={{ width: '100%' }}>
                <Button block icon={<DollarOutlined />}>
                  Record Payment
                </Button>
                <Button block>
                  Download PDF
                </Button>
                <Button block>
                  Send Reminder
                </Button>
              </Space>
            </Card>
          </Space>
        </Col>
      </Row>
    </div>
  )
}

export default InvoiceDetail 