import {
    CarOutlined,
    DeleteOutlined,
    EditOutlined,
    EnvironmentOutlined,
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
    Divider,
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
import { brokersApi, loadsApi } from '../services/api'

const { Title, Text } = Typography
const { Search } = Input
const { Option } = Select

const Loads = () => {
  const { message } = App.useApp()
  const [currentPage, setCurrentPage] = useState(1)
  const [pageSize, setPageSize] = useState(20)
  const [searchQuery, setSearchQuery] = useState('')
  const [statusFilter, setStatusFilter] = useState('')
  const [isModalVisible, setIsModalVisible] = useState(false)
  const [modalMode, setModalMode] = useState('create') // 'create', 'edit', 'view'
  const [selectedLoad, setSelectedLoad] = useState(null)
  const [form] = Form.useForm()
  
  const queryClient = useQueryClient()

  // Load data
  const { data: loadsData, isLoading, error } = useQuery(
    ['loads', currentPage, pageSize, searchQuery, statusFilter],
    () => {
      const params = { page: currentPage, limit: pageSize }
      if (statusFilter) params.status = statusFilter
      return loadsApi.getAll(params)
    },
    {
      keepPreviousData: true,
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
  const createMutation = useMutation(loadsApi.create, {
    onSuccess: () => {
      message.success('Load created successfully')
      queryClient.invalidateQueries('loads')
      setIsModalVisible(false)
      form.resetFields()
    },
    onError: (error) => {
      message.error(error.response?.data?.error || 'Error creating load')
    },
  })

  const updateMutation = useMutation(
    ({ id, data }) => loadsApi.update(id, data),
    {
      onSuccess: () => {
        message.success('Load updated successfully')
        queryClient.invalidateQueries('loads')
        setIsModalVisible(false)
        form.resetFields()
      },
      onError: (error) => {
        message.error(error.response?.data?.error || 'Error updating load')
      },
    }
  )

  const deleteMutation = useMutation(loadsApi.delete, {
    onSuccess: () => {
      message.success('Load deleted successfully')
      queryClient.invalidateQueries('loads')
    },
    onError: (error) => {
      message.error(error.response?.data?.error || 'Error deleting load')
    },
  })

  // Handlers
  const handleCreate = () => {
    setModalMode('create')
    setSelectedLoad(null)
    setIsModalVisible(true)
    form.resetFields()
    form.setFieldsValue({
      currency: 'USD',
      status: 'planned',
      pickup_date: dayjs().add(1, 'day'),
      delivery_date: dayjs().add(3, 'day'),
    })
  }

  const handleEdit = (load) => {
    setModalMode('edit')
    setSelectedLoad(load)
    setIsModalVisible(true)
    form.setFieldsValue({
      ...load,
      pickup_date: load.pickup_date ? dayjs(load.pickup_date) : null,
      delivery_date: load.delivery_date ? dayjs(load.delivery_date) : null,
      // Flatten route for form
      origin_address: load.route?.origin?.address,
      origin_city: load.route?.origin?.city,
      origin_state: load.route?.origin?.state,
      destination_address: load.route?.destination?.address,
      destination_city: load.route?.destination?.city,
      destination_state: load.route?.destination?.state,
    })
  }

  const handleView = (load) => {
    setModalMode('view')
    setSelectedLoad(load)
    setIsModalVisible(true)
  }

  const handleDelete = (loadId) => {
    deleteMutation.mutate(loadId)
  }

  const handleModalOk = () => {
    if (modalMode === 'view') {
      setIsModalVisible(false)
      return
    }

    form.validateFields().then((values) => {
      const loadData = {
        ...values,
        pickup_date: values.pickup_date ? values.pickup_date.toISOString() : null,
        delivery_date: values.delivery_date ? values.delivery_date.toISOString() : null,
        route: {
          origin: {
            address: values.origin_address || '',
            city: values.origin_city || '',
            state: values.origin_state || '',
          },
          destination: {
            address: values.destination_address || '',
            city: values.destination_city || '',
            state: values.destination_state || '',
          },
        },
      }

      // Remove flat route fields
      delete loadData.origin_address
      delete loadData.origin_city
      delete loadData.origin_state
      delete loadData.destination_address
      delete loadData.destination_city
      delete loadData.destination_state

      if (modalMode === 'create') {
        createMutation.mutate(loadData)
      } else if (modalMode === 'edit') {
        updateMutation.mutate({ id: selectedLoad.id, data: loadData })
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

  // Render status
  const renderStatus = (status) => {
    const statusConfig = {
      planned: { color: 'blue', text: 'Planned' },
      in_transit: { color: 'orange', text: 'In Transit' },
      delivered: { color: 'green', text: 'Delivered' },
      cancelled: { color: 'red', text: 'Cancelled' },
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
      title: 'Load #',
      dataIndex: 'load_number',
      key: 'load_number',
      render: (text, record) => (
        <Space>
          <CarOutlined style={{ color: '#1890ff' }} />
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
      title: 'Route',
      key: 'route',
      render: (record) => {
        const origin = record.route?.origin
        const destination = record.route?.destination
        return (
          <div>
            <div>
              <EnvironmentOutlined style={{ color: '#52c41a' }} />
              <Text style={{ marginLeft: 4 }}>
                {origin?.city}, {origin?.state}
              </Text>
            </div>
            <div style={{ marginTop: 4 }}>
              <EnvironmentOutlined style={{ color: '#ff4d4f' }} />
              <Text style={{ marginLeft: 4 }}>
                {destination?.city}, {destination?.state}
              </Text>
            </div>
          </div>
        )
      },
    },
    {
      title: 'Broker',
      dataIndex: 'broker_name',
      key: 'broker_name',
      render: (text) => text || 'Unknown Broker',
    },
    {
      title: 'Cost',
      dataIndex: 'cost',
      key: 'cost',
      align: 'right',
      render: (value, record) => formatCurrency(value, record.currency),
    },
    {
      title: 'Pickup Date',
      dataIndex: 'pickup_date',
      key: 'pickup_date',
      render: (date) => {
        if (!date) return '-'
        return dayjs(date).format('MMM DD, YYYY')
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
      title: 'Driver',
      dataIndex: 'driver_name',
      key: 'driver_name',
      render: (text) => text || 'Not assigned',
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
              title="Delete load?"
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
                Loads
              </Title>
              <Text type="secondary">
                Manage shipment loads
              </Text>
            </Col>
            <Col>
              <Button
                type="primary"
                icon={<PlusOutlined />}
                onClick={handleCreate}
                size="large"
              >
                Create Load
              </Button>
            </Col>
          </Row>
        </div>

        <div style={{ marginBottom: 16 }}>
          <Row gutter={16}>
            <Col span={12}>
              <Search
                placeholder="Search by load number or driver..."
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
                <Option value="planned">Planned</Option>
                <Option value="in_transit">In Transit</Option>
                <Option value="delivered">Delivered</Option>
                <Option value="cancelled">Cancelled</Option>
              </Select>
            </Col>
          </Row>
        </div>

        <Table
          columns={columns}
          dataSource={loadsData?.data || []}
          rowKey="id"
          loading={isLoading}
          pagination={{
            current: currentPage,
            pageSize: pageSize,
            total: loadsData?.pagination?.total || 0,
            showSizeChanger: true,
            showQuickJumper: true,
            showTotal: (total, range) =>
              `${range[0]}-${range[1]} of ${total} records`,
            onChange: (page, size) => {
              setCurrentPage(page)
              setPageSize(size)
            },
          }}
          scroll={{ x: 1200 }}
        />
      </Card>

      {/* Modal */}
      <Modal
        title={
          modalMode === 'create'
            ? 'Create Load'
            : modalMode === 'edit'
            ? 'Edit Load'
            : 'Load Details'
        }
        open={isModalVisible}
        onOk={handleModalOk}
        onCancel={() => setIsModalVisible(false)}
        width={800}
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
                label="Load Number"
                name="load_number"
                rules={[{ required: true, message: 'Please enter load number' }]}
              >
                <Input placeholder="LD-20241201-001" />
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

          <Divider>Route Information</Divider>

          <Row gutter={16}>
            <Col span={12}>
              <Title level={5}>Origin</Title>
              <Form.Item
                label="Address"
                name="origin_address"
                rules={[{ required: true, message: 'Please enter origin address' }]}
              >
                <Input placeholder="123 Main Street" />
              </Form.Item>
              <Row gutter={8}>
                <Col span={12}>
                  <Form.Item
                    label="City"
                    name="origin_city"
                    rules={[{ required: true, message: 'Please enter city' }]}
                  >
                    <Input placeholder="New York" />
                  </Form.Item>
                </Col>
                <Col span={12}>
                  <Form.Item
                    label="State"
                    name="origin_state"
                    rules={[{ required: true, message: 'Please enter state' }]}
                  >
                    <Input placeholder="NY" />
                  </Form.Item>
                </Col>
              </Row>
            </Col>
            
            <Col span={12}>
              <Title level={5}>Destination</Title>
              <Form.Item
                label="Address"
                name="destination_address"
                rules={[{ required: true, message: 'Please enter destination address' }]}
              >
                <Input placeholder="456 Oak Avenue" />
              </Form.Item>
              <Row gutter={8}>
                <Col span={12}>
                  <Form.Item
                    label="City"
                    name="destination_city"
                    rules={[{ required: true, message: 'Please enter city' }]}
                  >
                    <Input placeholder="Los Angeles" />
                  </Form.Item>
                </Col>
                <Col span={12}>
                  <Form.Item
                    label="State"
                    name="destination_state"
                    rules={[{ required: true, message: 'Please enter state' }]}
                  >
                    <Input placeholder="CA" />
                  </Form.Item>
                </Col>
              </Row>
            </Col>
          </Row>

          <Divider>Load Details</Divider>

          <Row gutter={16}>
            <Col span={8}>
              <Form.Item
                label="Cost"
                name="cost"
                rules={[
                  { required: true, message: 'Please enter cost' },
                  { type: 'number', min: 0.01, message: 'Cost must be greater than 0' }
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
                  <Option value="planned">Planned</Option>
                  <Option value="in_transit">In Transit</Option>
                  <Option value="delivered">Delivered</Option>
                  <Option value="cancelled">Cancelled</Option>
                </Select>
              </Form.Item>
            </Col>
          </Row>

          <Row gutter={16}>
            <Col span={12}>
              <Form.Item
                label="Pickup Date"
                name="pickup_date"
                rules={[{ required: true, message: 'Please select pickup date' }]}
              >
                <DatePicker 
                  style={{ width: '100%' }} 
                  format="YYYY-MM-DD"
                  placeholder="Select pickup date"
                />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item
                label="Delivery Date"
                name="delivery_date"
              >
                <DatePicker 
                  style={{ width: '100%' }} 
                  format="YYYY-MM-DD"
                  placeholder="Select delivery date"
                />
              </Form.Item>
            </Col>
          </Row>

          <Divider>Driver & Equipment</Divider>

          <Row gutter={16}>
            <Col span={12}>
              <Form.Item
                label="Driver Name"
                name="driver_name"
              >
                <Input placeholder="John Doe" />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item
                label="Driver Phone"
                name="driver_phone"
              >
                <Input placeholder="+1 (555) 123-4567" />
              </Form.Item>
            </Col>
          </Row>

          <Row gutter={16}>
            <Col span={8}>
              <Form.Item
                label="Truck"
                name="truck"
              >
                <Input placeholder="Truck details" />
              </Form.Item>
            </Col>
            <Col span={8}>
              <Form.Item
                label="Trailer"
                name="trailer"
              >
                <Input placeholder="Trailer details" />
              </Form.Item>
            </Col>
            <Col span={8}>
              <Form.Item
                label="Equipment Type"
                name="equipment_type"
              >
                <Select placeholder="Select equipment">
                  <Option value="dry_van">Dry Van</Option>
                  <Option value="reefer">Reefer</Option>
                  <Option value="flatbed">Flatbed</Option>
                  <Option value="step_deck">Step Deck</Option>
                  <Option value="lowboy">Lowboy</Option>
                </Select>
              </Form.Item>
            </Col>
          </Row>

          <Form.Item
            label="Notes"
            name="notes"
          >
            <Input.TextArea
              rows={3}
              placeholder="Additional notes about the load..."
            />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}

export default Loads 