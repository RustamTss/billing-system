import {
    CreditCardOutlined,
    DeleteOutlined,
    DollarOutlined,
    EditOutlined,
    EyeOutlined,
    PlusOutlined,
    SearchOutlined,
} from '@ant-design/icons'
import {
    App,
    Button,
    Card,
    Col,
    DatePicker,
    Form,
    Input,
    InputNumber,
    Modal,
    Popconfirm,
    Row,
    Select,
    Space,
    Table,
    Tag,
    Tooltip,
    Typography
} from 'antd'
import dayjs from 'dayjs'
import React, { useState } from 'react'
import { useMutation, useQuery, useQueryClient } from 'react-query'
import { brokersApi, invoicesApi, paymentsApi } from '../services/api'

const { Title, Text } = Typography
const { Search } = Input
const { Option } = Select

const Payments = () => {
  const { message } = App.useApp()
  const [currentPage, setCurrentPage] = useState(1)
  const [pageSize, setPageSize] = useState(20)
  const [searchQuery, setSearchQuery] = useState('')
  const [methodFilter, setMethodFilter] = useState('')
  const [isModalVisible, setIsModalVisible] = useState(false)
  const [modalMode, setModalMode] = useState('create') // 'create', 'edit', 'view'
  const [selectedPayment, setSelectedPayment] = useState(null)
  const [selectedBrokerId, setSelectedBrokerId] = useState(null) // Для фильтрации инвойсов
  const [form] = Form.useForm()
  
  const queryClient = useQueryClient()

  // Load data
  const { data: paymentsData, isLoading, error } = useQuery(
    ['payments', currentPage, pageSize, searchQuery, methodFilter],
    () => {
      const params = { page: currentPage, limit: pageSize }
      if (methodFilter) params.method = methodFilter
      return paymentsApi.getAll(params)
    },
    {
      keepPreviousData: true,
    }
  )

  // Load invoices for selection (filtered by broker if selected)
  const { data: invoicesData, isLoading: invoicesLoading } = useQuery(
    ['invoices-select', selectedBrokerId], 
    () => {
      const params = { limit: 1000, status: 'pending,partial' }
      if (selectedBrokerId) {
        params.broker_id = selectedBrokerId
      }
      return invoicesApi.getAll(params)
    },
    {
      staleTime: 5 * 60 * 1000, // Cache for 5 minutes
      onSuccess: (data) => {
        console.log('Invoices loaded for broker:', selectedBrokerId, data)
      }
    }
  )

  // Load brokers for selection
  const { data: brokersData, isLoading: brokersLoading } = useQuery('brokers-select', () => 
    brokersApi.getAll({ limit: 1000 }), // Get all brokers for selection
    {
      staleTime: 5 * 60 * 1000, // Cache for 5 minutes
      onSuccess: (data) => {
        console.log('Brokers loaded:', data)
      }
    }
  )

  // Mutations
  const createMutation = useMutation(paymentsApi.create, {
    onSuccess: () => {
      message.success('Payment created successfully')
      queryClient.invalidateQueries('payments')
      setIsModalVisible(false)
      form.resetFields()
    },
    onError: (error) => {
      message.error(error.response?.data?.error || 'Error creating payment')
    },
  })

  const updateMutation = useMutation(
    ({ id, data }) => paymentsApi.update(id, data),
    {
      onSuccess: () => {
        message.success('Payment updated successfully')
        queryClient.invalidateQueries('payments')
        setIsModalVisible(false)
        form.resetFields()
      },
      onError: (error) => {
        message.error(error.response?.data?.error || 'Error updating payment')
      },
    }
  )

  const deleteMutation = useMutation(paymentsApi.delete, {
    onSuccess: () => {
      message.success('Payment deleted successfully')
      queryClient.invalidateQueries('payments')
    },
    onError: (error) => {
      message.error(error.response?.data?.error || 'Error deleting payment')
    },
  })

  // Handlers
  const handleCreate = () => {
    setModalMode('create')
    setSelectedPayment(null)
    setSelectedBrokerId(null) // Сбрасываем фильтр брокера
    setIsModalVisible(true)
    form.resetFields()
    form.setFieldsValue({
      currency: 'USD',
      payment_method: 'wire_transfer',
      payment_date: dayjs(),
    })
  }

  const handleEdit = (payment) => {
    setModalMode('edit')
    setSelectedPayment(payment)
    setIsModalVisible(true)
    form.setFieldsValue({
      ...payment,
      payment_date: payment.payment_date ? dayjs(payment.payment_date) : null,
    })
  }

  const handleView = (payment) => {
    setModalMode('view')
    setSelectedPayment(payment)
    setIsModalVisible(true)
  }

  const handleDelete = (paymentId) => {
    deleteMutation.mutate(paymentId)
  }

  const handleModalOk = () => {
    if (modalMode === 'view') {
      setIsModalVisible(false)
      return
    }

    form.validateFields().then((values) => {
      const paymentData = {
        ...values,
        payment_date: values.payment_date ? values.payment_date.toISOString() : null,
      }

      if (modalMode === 'create') {
        createMutation.mutate(paymentData)
      } else if (modalMode === 'edit') {
        updateMutation.mutate({ id: selectedPayment.id, data: paymentData })
      }
    })
  }

  const handleSearch = (value) => {
    setSearchQuery(value)
    setCurrentPage(1)
  }

  const handleMethodFilter = (value) => {
    setMethodFilter(value)
    setCurrentPage(1)
  }

  // Handler для изменения брокера в форме
  const handleBrokerChange = (brokerId) => {
    setSelectedBrokerId(brokerId)
    // Сбрасываем выбранный инвойс при смене брокера
    form.setFieldsValue({ invoice_id: undefined })
  }

  // Render payment method
  const renderPaymentMethod = (method) => {
    const methodConfig = {
      wire_transfer: { color: 'blue', text: 'Wire Transfer' },
      check: { color: 'orange', text: 'Check' },
      cash: { color: 'green', text: 'Cash' },
      card: { color: 'purple', text: 'Card' },
      crypto: { color: 'gold', text: 'Crypto' },
    }
    
    const config = methodConfig[method] || { color: 'default', text: method }
    return <Tag color={config.color}>{config.text}</Tag>
  }

  // Format currency
  const formatCurrency = (amount, currency) => {
    const symbol = currency === 'USD' ? '$' : currency === 'EUR' ? '€' : '₽'
    return `${symbol}${amount?.toLocaleString() || 0}`
  }

  // Table columns
  const columns = [
    {
      title: 'Transaction',
      key: 'transaction',
      render: (record) => (
        <Space>
          <CreditCardOutlined style={{ color: '#1890ff' }} />
          <div>
            <Text strong>{record.transaction_id || `PAY-${record.id?.slice(-8)}`}</Text>
            <div><Text type="secondary" style={{ fontSize: '12px' }}>
              {dayjs(record.payment_date).format('MMM DD, YYYY')}
            </Text></div>
          </div>
        </Space>
      ),
    },
    {
      title: 'Invoice #',
      dataIndex: 'invoice_number',
      key: 'invoice_number',
      render: (text) => text || 'N/A',
    },
    {
      title: 'Broker',
      dataIndex: 'broker_name',
      key: 'broker_name',
      render: (text) => text || 'Unknown Broker',
    },
    {
      title: 'Amount',
      dataIndex: 'amount',
      key: 'amount',
      align: 'right',
      render: (value, record) => (
        <Space>
          <DollarOutlined />
          <Text strong style={{ color: '#52c41a' }}>
            {formatCurrency(value, record.currency)}
          </Text>
        </Space>
      ),
    },
    {
      title: 'Method',
      dataIndex: 'payment_method',
      key: 'payment_method',
      align: 'center',
      render: renderPaymentMethod,
    },
    {
      title: 'Reference',
      dataIndex: 'reference_number',
      key: 'reference_number',
      render: (text) => text || '-',
    },
    {
      title: 'Actions',
      key: 'actions',
      align: 'center',
      width: 150,
      render: (record) => (
        <Space>
          <Tooltip title="View">
            <Button
              icon={<EyeOutlined />}
              size="small"
              onClick={() => handleView(record)}
            />
          </Tooltip>
          <Tooltip title="Edit">
            <Button
              type="primary"
              icon={<EditOutlined />}
              size="small"
              onClick={() => handleEdit(record)}
            />
          </Tooltip>
          <Tooltip title="Delete">
            <Popconfirm
              title="Delete payment?"
              description="This action cannot be undone"
              onConfirm={() => handleDelete(record.id)}
              okText="Yes"
              cancelText="No"
            >
              <Button
                danger
                icon={<DeleteOutlined />}
                size="small"
              />
            </Popconfirm>
          </Tooltip>
        </Space>
      ),
    },
  ]

  return (
    <div>
      <Card>
        <div style={{ marginBottom: 16 }}>
          <Row justify="space-between" align="middle">
            <Col>
              <Title level={2} style={{ margin: 0 }}>
                Payments
              </Title>
              <Text type="secondary">
                Manage payment records
              </Text>
            </Col>
            <Col>
              <Button
                type="primary"
                icon={<PlusOutlined />}
                onClick={handleCreate}
                size="large"
              >
                Record Payment
              </Button>
            </Col>
          </Row>
        </div>

        <div style={{ marginBottom: 16 }}>
          <Row gutter={16}>
            <Col span={12}>
              <Search
                placeholder="Search by transaction ID or reference..."
                allowClear
                enterButton={<SearchOutlined />}
                size="large"
                onSearch={handleSearch}
              />
            </Col>
            <Col span={6}>
              <Select
                placeholder="Filter by method"
                allowClear
                size="large"
                style={{ width: '100%' }}
                onChange={handleMethodFilter}
              >
                <Option value="wire_transfer">Wire Transfer</Option>
                <Option value="check">Check</Option>
                <Option value="cash">Cash</Option>
                <Option value="card">Card</Option>
                <Option value="crypto">Crypto</Option>
              </Select>
            </Col>
          </Row>
        </div>

        <Table
          columns={columns}
          dataSource={paymentsData?.data || []}
          rowKey="id"
          loading={isLoading}
          pagination={{
            current: currentPage,
            pageSize: pageSize,
            total: paymentsData?.pagination?.total || 0,
            showSizeChanger: true,
            showQuickJumper: true,
            showTotal: (total, range) =>
              `${range[0]}-${range[1]} of ${total} records`,
            onChange: (page, size) => {
              setCurrentPage(page)
              setPageSize(size)
            },
          }}
          scroll={{ x: 1000 }}
        />
      </Card>

      {/* Modal */}
      <Modal
        title={
          modalMode === 'create'
            ? 'Record Payment'
            : modalMode === 'edit'
            ? 'Edit Payment'
            : 'Payment Details'
        }
        open={isModalVisible}
        onOk={handleModalOk}
        onCancel={() => setIsModalVisible(false)}
        width={600}
        okText={modalMode === 'view' ? 'Close' : 'Save'}
        cancelText="Cancel"
        confirmLoading={createMutation.isLoading || updateMutation.isLoading}
        footer={modalMode === 'view' ? [
          <Button key="close" onClick={() => setIsModalVisible(false)}>
            Close
          </Button>
        ] : undefined}
      >
        <Form
          form={form}
          layout="vertical"
          disabled={modalMode === 'view'}
        >
          <Row gutter={16}>
            <Col span={12}>
              <Form.Item
                label="Broker"
                name="broker_id"
                rules={[{ required: true, message: 'Please select a broker' }]}
              >
                <Select
                  placeholder="Select a broker first"
                  loading={brokersLoading}
                  showSearch
                  onChange={handleBrokerChange}
                  filterOption={(input, option) =>
                    (option?.label ?? '').toLowerCase().includes(input.toLowerCase())
                  }
                  options={brokersData?.data?.map(broker => ({
                    value: broker.id,
                    label: `${broker.company_name} (${broker.contact_person || 'No contact'})`,
                  })) || []}
                  notFoundContent={brokersLoading ? 'Loading...' : 'No brokers found'}
                />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item
                label="Invoice"
                name="invoice_id"
                rules={[{ required: true, message: 'Please select an invoice' }]}
              >
                <Select
                  placeholder={selectedBrokerId ? "Select an invoice" : "Select a broker first"}
                  loading={invoicesLoading}
                  disabled={!selectedBrokerId}
                  showSearch
                  filterOption={(input, option) =>
                    (option?.label ?? '').toLowerCase().includes(input.toLowerCase())
                  }
                  options={invoicesData?.data?.map(invoice => ({
                    value: invoice.id,
                    label: `${invoice.invoice_number} (${invoice.currency === 'USD' ? '$' : invoice.currency === 'EUR' ? '€' : '₽'}${invoice.amount})`,
                  })) || []}
                  notFoundContent={
                    !selectedBrokerId 
                      ? 'Please select a broker first' 
                      : invoicesLoading 
                        ? 'Loading...' 
                        : 'No unpaid invoices found for this broker'
                  }
                />
              </Form.Item>
            </Col>
          </Row>

          <Row gutter={16}>
            <Col span={8}>
              <Form.Item
                label="Amount"
                name="amount"
                rules={[
                  { required: true, message: 'Please enter amount' },
                  { type: 'number', min: 0.01, message: 'Amount must be greater than 0' }
                ]}
              >
                <InputNumber
                  style={{ width: '100%' }}
                  min={0}
                  step={0.01}
                  precision={2}
                  placeholder="0.00"
                />
              </Form.Item>
            </Col>
            <Col span={8}>
              <Form.Item
                label="Currency"
                name="currency"
                rules={[{ required: true, message: 'Please select currency' }]}
              >
                <Select>
                  <Option value="USD">USD ($)</Option>
                  <Option value="EUR">EUR (€)</Option>
                  <Option value="RUB">RUB (₽)</Option>
                </Select>
              </Form.Item>
            </Col>
            <Col span={8}>
              <Form.Item
                label="Payment Date"
                name="payment_date"
                rules={[{ required: true, message: 'Please select payment date' }]}
              >
                <DatePicker 
                  style={{ width: '100%' }} 
                  format="YYYY-MM-DD"
                  placeholder="Select date"
                />
              </Form.Item>
            </Col>
          </Row>

          <Row gutter={16}>
            <Col span={12}>
              <Form.Item
                label="Payment Method"
                name="payment_method"
                rules={[{ required: true, message: 'Please select payment method' }]}
              >
                <Select>
                  <Option value="wire_transfer">Wire Transfer</Option>
                  <Option value="check">Check</Option>
                  <Option value="cash">Cash</Option>
                  <Option value="card">Card</Option>
                  <Option value="crypto">Crypto</Option>
                </Select>
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item
                label="Transaction ID"
                name="transaction_id"
              >
                <Input placeholder="Enter transaction ID" />
              </Form.Item>
            </Col>
          </Row>

          <Form.Item
            label="Reference Number"
            name="reference_number"
          >
            <Input placeholder="Enter reference number" />
          </Form.Item>

          <Form.Item
            label="Notes"
            name="notes"
          >
            <Input.TextArea
              rows={3}
              placeholder="Additional notes about the payment..."
            />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}

export default Payments 