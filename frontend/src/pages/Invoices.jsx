import {
  DeleteOutlined,
  DollarOutlined,
  EditOutlined,
  EyeOutlined,
  FileTextOutlined,
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
import { brokersApi, invoicesApi, loadsApi } from '../services/api'

const { Title, Text } = Typography
const { Search } = Input
const { Option } = Select

const Invoices = () => {
  const { message } = App.useApp()
  const [currentPage, setCurrentPage] = useState(1)
  const [pageSize, setPageSize] = useState(20)
  const [searchQuery, setSearchQuery] = useState('')
  const [statusFilter, setStatusFilter] = useState('')
  const [isModalVisible, setIsModalVisible] = useState(false)
  const [modalMode, setModalMode] = useState('create') // 'create', 'edit', 'view'
  const [selectedInvoice, setSelectedInvoice] = useState(null)
  const [selectedBrokerId, setSelectedBrokerId] = useState(null) // Для загрузки неоплаченных грузов
  const [selectedLoads, setSelectedLoads] = useState([]) // Выбранные грузы для инвойса
  const [form] = Form.useForm()
  
  const queryClient = useQueryClient()

  // Load data
  const { data: invoicesData, isLoading, error } = useQuery(
    ['invoices', currentPage, pageSize, searchQuery, statusFilter],
    () => {
      const params = { page: currentPage, limit: pageSize }
      if (statusFilter) params.status = statusFilter
      if (searchQuery) {
        // For simplicity, using getAll with search in invoice number
        return invoicesApi.getAll(params)
      }
      return invoicesApi.getAll(params)
    },
    {
      keepPreviousData: true,
    }
  )

  // Load brokers for selection
  const { data: brokersData, isLoading: brokersLoading, error: brokersError } = useQuery('brokers-select', () => 
    brokersApi.getAll({ limit: 1000 }), // Get all brokers for selection
    {
      staleTime: 5 * 60 * 1000, // Cache for 5 minutes
      onSuccess: (data) => {
        console.log('Brokers loaded:', data)
      },
      onError: (error) => {
        console.error('Error loading brokers:', error)
      }
    }
  )

  // Load unbilled loads for selected broker
  const { data: unbilledLoadsData, isLoading: loadsLoading } = useQuery(
    ['unbilled-loads', selectedBrokerId], 
    () => loadsApi.getUnbilledByBroker(selectedBrokerId, { limit: 100 }),
    {
      enabled: !!selectedBrokerId && modalMode === 'create',
      staleTime: 2 * 60 * 1000, // Cache for 2 minutes
      onSuccess: (data) => {
        console.log('Unbilled loads loaded:', data)
      }
    }
  )

  // Mutations
  const createMutation = useMutation(invoicesApi.create, {
    onSuccess: (data) => {
      console.log('Invoice created successfully:', data)
      message.success('Invoice created successfully')
      queryClient.invalidateQueries('invoices')
      setIsModalVisible(false)
      form.resetFields()
    },
    onError: (error) => {
      console.error('Error creating invoice:', error)
      console.error('Error response:', error.response)
      console.error('Error data:', error.response?.data)
      message.error(error.response?.data?.error || 'Error creating invoice')
    },
  })

  const updateMutation = useMutation(
    ({ id, data }) => invoicesApi.update(id, data),
    {
      onSuccess: () => {
        message.success('Invoice updated successfully')
        queryClient.invalidateQueries('invoices')
        setIsModalVisible(false)
        form.resetFields()
      },
      onError: (error) => {
        message.error(error.response?.data?.error || 'Error updating invoice')
      },
    }
  )

  const deleteMutation = useMutation(invoicesApi.delete, {
    onSuccess: () => {
      message.success('Invoice deleted successfully')
      queryClient.invalidateQueries('invoices')
    },
    onError: (error) => {
      message.error(error.response?.data?.error || 'Error deleting invoice')
    },
  })

  // Handlers
  const handleCreate = () => {
    setModalMode('create')
    setSelectedInvoice(null)
    setSelectedBrokerId(null)
    setSelectedLoads([])
    setIsModalVisible(true)
    form.resetFields()
    // Set default values
    form.setFieldsValue({
      currency: 'USD',
      status: 'pending',
      due_date: dayjs().add(30, 'day'),
    })
  }

  const handleEdit = (invoice) => {
    setModalMode('edit')
    setSelectedInvoice(invoice)
    setIsModalVisible(true)
    form.setFieldsValue({
      ...invoice,
      due_date: invoice.due_date ? dayjs(invoice.due_date) : null,
      created_at: invoice.created_at ? dayjs(invoice.created_at) : null,
    })
  }

  const handleView = (invoice) => {
    setModalMode('view')
    setSelectedInvoice(invoice)
    setIsModalVisible(true)
  }

  const handleDelete = (invoiceId) => {
    deleteMutation.mutate(invoiceId)
  }

  const handleModalOk = () => {
    if (modalMode === 'view') {
      setIsModalVisible(false)
      return
    }

    form.validateFields().then((values) => {
      console.log('Form values:', values)
      const invoiceData = {
        ...values,
        due_date: values.due_date ? values.due_date.toISOString() : null,
      }
      console.log('Invoice data to send:', invoiceData)

      if (modalMode === 'create') {
        console.log('Creating invoice...')
        createMutation.mutate(invoiceData)
      } else if (modalMode === 'edit') {
        console.log('Updating invoice...')
        updateMutation.mutate({ id: selectedInvoice.id, data: invoiceData })
      }
    })
  }

  const handleSearch = (value) => {
    setSearchQuery(value)
    setCurrentPage(1)
  }

  const handleStatusFilter = (value) => {
    setStatusFilter(value)
    setCurrentPage(1)
  }

  // Handler для изменения брокера в форме
  const handleBrokerChange = (brokerId) => {
    setSelectedBrokerId(brokerId)
    // Сбрасываем выбранные грузы при смене брокера
    setSelectedLoads([])
    form.setFieldsValue({ 
      load_ids: undefined,
      amount: undefined 
    })
  }

  // Handler для выбора грузов
  const handleLoadSelection = (loadIds) => {
    setSelectedLoads(loadIds)
    
    // Автоматически рассчитываем сумму на основе выбранных грузов
    if (loadIds.length > 0 && unbilledLoadsData?.data) {
      const selectedLoadObjects = unbilledLoadsData.data.filter(load => 
        loadIds.includes(load.id)
      )
      const totalAmount = selectedLoadObjects.reduce((sum, load) => sum + (load.cost || 0), 0)
      
      // Берем валюту из первого выбранного груза
      const currency = selectedLoadObjects[0]?.currency || 'USD'
      
      form.setFieldsValue({
        amount: totalAmount,
        currency: currency
      })
    } else {
      form.setFieldsValue({
        amount: undefined
      })
    }
  }

  // Render status
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

  // Format currency
  const formatCurrency = (amount, currency) => {
    const symbol = currency === 'USD' ? '$' : currency === 'EUR' ? '€' : '₽'
    return `${symbol}${amount?.toLocaleString() || 0}`
  }

  // Table columns
  const columns = [
    {
      title: 'Invoice #',
      dataIndex: 'invoice_number',
      key: 'invoice_number',
      render: (text, record) => (
        <Space>
          <FileTextOutlined style={{ color: '#1890ff' }} />
          <div>
            <Text strong>{text}</Text>
            <div><Text type="secondary" style={{ fontSize: '12px' }}>
              {dayjs(record.created_at).format('MMM DD, YYYY')}
            </Text></div>
          </div>
        </Space>
      ),
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
          <Text strong>{formatCurrency(value, record.currency)}</Text>
        </Space>
      ),
    },
    {
      title: 'Due Date',
      dataIndex: 'due_date',
      key: 'due_date',
      render: (date) => {
        if (!date) return '-'
        const dueDate = dayjs(date)
        const isOverdue = dueDate.isBefore(dayjs())
        return (
          <Text type={isOverdue ? 'danger' : 'default'}>
            {dueDate.format('MMM DD, YYYY')}
          </Text>
        )
      },
    },
    {
      title: 'Status',
      dataIndex: 'status',
      key: 'status',
      align: 'center',
      render: renderStatus,
    },
    {
      title: 'Paid Amount',
      dataIndex: 'paid_amount',
      key: 'paid_amount',
      align: 'right',
      render: (value, record) => formatCurrency(value || 0, record.currency),
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
              title="Delete invoice?"
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
                Invoices
              </Title>
              <Text type="secondary">
                Manage billing invoices
              </Text>
            </Col>
            <Col>
              <Button
                type="primary"
                icon={<PlusOutlined />}
                onClick={handleCreate}
                size="large"
              >
                Create Invoice
              </Button>
            </Col>
          </Row>
        </div>

        <div style={{ marginBottom: 16 }}>
          <Row gutter={16}>
            <Col span={12}>
              <Search
                placeholder="Search by invoice number or broker..."
                allowClear
                enterButton={<SearchOutlined />}
                size="large"
                onSearch={handleSearch}
              />
            </Col>
            <Col span={6}>
              <Select
                placeholder="Filter by status"
                allowClear
                size="large"
                style={{ width: '100%' }}
                onChange={handleStatusFilter}
              >
                <Option value="pending">Pending</Option>
                <Option value="paid">Paid</Option>
                <Option value="partial">Partial</Option>
                <Option value="overdue">Overdue</Option>
                <Option value="cancelled">Cancelled</Option>
              </Select>
            </Col>
          </Row>
        </div>

        <Table
          columns={columns}
          dataSource={invoicesData?.data || []}
          rowKey="id"
          loading={isLoading}
          pagination={{
            current: currentPage,
            pageSize: pageSize,
            total: invoicesData?.pagination?.total || 0,
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
            ? 'Create Invoice'
            : modalMode === 'edit'
            ? 'Edit Invoice'
            : 'Invoice Details'
        }
        open={isModalVisible}
        onOk={handleModalOk}
        onCancel={() => setIsModalVisible(false)}
        width={700}
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
                label="Invoice Number"
                name="invoice_number"
                rules={[{ required: true, message: 'Please enter invoice number' }]}
              >
                <Input placeholder="INV-2024-001" />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item
                label="Broker"
                name="broker_id"
                rules={[{ required: true, message: 'Please select a broker' }]}
              >
                <Select
                  placeholder="Select a broker"
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
                label="Status"
                name="status"
                rules={[{ required: true, message: 'Please select status' }]}
              >
                <Select>
                  <Option value="pending">Pending</Option>
                  <Option value="paid">Paid</Option>
                  <Option value="partial">Partial</Option>
                  <Option value="overdue">Overdue</Option>
                  <Option value="cancelled">Cancelled</Option>
                </Select>
              </Form.Item>
            </Col>
          </Row>

          <Row gutter={16}>
            <Col span={12}>
              <Form.Item
                label="Due Date"
                name="due_date"
                rules={[{ required: true, message: 'Please select due date' }]}
              >
                <DatePicker 
                  style={{ width: '100%' }} 
                  format="YYYY-MM-DD"
                  placeholder="Select due date"
                />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item
                label="Paid Amount"
                name="paid_amount"
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
          </Row>

          {modalMode === 'create' && selectedBrokerId && (
            <Form.Item
              label="Loads to Include"
              name="load_ids"
              tooltip="Select unbilled loads to include in this invoice. Total amount will be calculated automatically."
            >
              <Select
                mode="multiple"
                placeholder={selectedBrokerId ? "Select loads to include in invoice" : "Select a broker first"}
                loading={loadsLoading}
                disabled={!selectedBrokerId}
                onChange={handleLoadSelection}
                value={selectedLoads}
                showSearch
                filterOption={(input, option) =>
                  (option?.label ?? '').toLowerCase().includes(input.toLowerCase())
                }
                options={unbilledLoadsData?.data?.map(load => ({
                  value: load.id,
                  label: `${load.load_number} - ${load.route?.origin?.city || 'N/A'} → ${load.route?.destination?.city || 'N/A'} ($${load.cost || 0})`,
                })) || []}
                notFoundContent={
                  !selectedBrokerId 
                    ? 'Please select a broker first' 
                    : loadsLoading 
                      ? 'Loading...' 
                      : 'No unbilled loads found for this broker'
                }
              />
            </Form.Item>
          )}

          <Form.Item
            label="Description"
            name="description"
          >
            <Input.TextArea
              rows={3}
              placeholder="Invoice description or notes..."
            />
          </Form.Item>

          <Form.Item
            label="Terms"
            name="terms"
          >
            <Input.TextArea
              rows={2}
              placeholder="Payment terms and conditions..."
            />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}

export default Invoices 