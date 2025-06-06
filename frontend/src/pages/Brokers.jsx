import {
  DeleteOutlined,
  EditOutlined,
  EyeOutlined,
  PlusOutlined,
  SearchOutlined,
  UserOutlined,
} from '@ant-design/icons'
import {
  Button,
  Card,
  Col,
  Divider,
  Form,
  Input,
  InputNumber,
  message,
  Modal,
  Popconfirm,
  Rate,
  Row,
  Select,
  Space,
  Table,
  Tag,
  Tooltip,
  Typography,
} from 'antd'
import React, { useState } from 'react'
import { useMutation, useQuery, useQueryClient } from 'react-query'
import { brokersApi } from '../services/api'

const { Title, Text } = Typography
const { Search } = Input
const { Option } = Select

const Brokers = () => {
  const [currentPage, setCurrentPage] = useState(1)
  const [pageSize, setPageSize] = useState(20)
  const [searchQuery, setSearchQuery] = useState('')
  const [isModalVisible, setIsModalVisible] = useState(false)
  const [modalMode, setModalMode] = useState('create') // 'create', 'edit', 'view'
  const [selectedBroker, setSelectedBroker] = useState(null)
  const [form] = Form.useForm()
  
  const queryClient = useQueryClient()

  // Load data
  const { data: brokersData, isLoading, error } = useQuery(
    ['brokers', currentPage, pageSize, searchQuery],
    () => {
      if (searchQuery) {
        return brokersApi.search(searchQuery, { page: currentPage, limit: pageSize })
      }
      return brokersApi.getAll({ page: currentPage, limit: pageSize })
    },
    {
      keepPreviousData: true,
    }
  )

  // Mutations
  const createMutation = useMutation(brokersApi.create, {
    onSuccess: () => {
      message.success('Broker created successfully')
      queryClient.invalidateQueries('brokers')
      setIsModalVisible(false)
      form.resetFields()
    },
    onError: (error) => {
      message.error(error.response?.data?.error || 'Error creating broker')
    },
  })

  const updateMutation = useMutation(
    ({ id, data }) => brokersApi.update(id, data),
    {
      onSuccess: () => {
        message.success('Broker updated successfully')
        queryClient.invalidateQueries('brokers')
        setIsModalVisible(false)
        form.resetFields()
      },
      onError: (error) => {
        message.error(error.response?.data?.error || 'Error updating broker')
      },
    }
  )

  const deleteMutation = useMutation(brokersApi.delete, {
    onSuccess: () => {
      message.success('Broker deleted successfully')
      queryClient.invalidateQueries('brokers')
    },
    onError: (error) => {
      message.error(error.response?.data?.error || 'Error deleting broker')
    },
  })

  // Handlers
  const handleCreate = () => {
    setModalMode('create')
    setSelectedBroker(null)
    setIsModalVisible(true)
    form.resetFields()
  }

  const handleEdit = (broker) => {
    setModalMode('edit')
    setSelectedBroker(broker)
    setIsModalVisible(true)
    form.setFieldsValue({
      ...broker,
      // Expand address for form
      street: broker.address?.street,
      city: broker.address?.city,
      state: broker.address?.state,
      country: broker.address?.country,
      zip_code: broker.address?.zip_code,
    })
  }

  const handleView = (broker) => {
    setModalMode('view')
    setSelectedBroker(broker)
    setIsModalVisible(true)
  }

  const handleDelete = (brokerId) => {
    deleteMutation.mutate(brokerId)
  }

  const handleModalOk = () => {
    if (modalMode === 'view') {
      setIsModalVisible(false)
      return
    }

    form.validateFields().then((values) => {
      // Form address
      const brokerData = {
        ...values,
        address: {
          street: values.street || '',
          city: values.city || '',
          state: values.state || '',
          country: values.country || '',
          zip_code: values.zip_code || '',
        },
      }
      
      // Remove separate address fields
      delete brokerData.street
      delete brokerData.city
      delete brokerData.state
      delete brokerData.country
      delete brokerData.zip_code

      if (modalMode === 'create') {
        createMutation.mutate(brokerData)
      } else if (modalMode === 'edit') {
        updateMutation.mutate({ id: selectedBroker.id, data: brokerData })
      }
    })
  }

  const handleSearch = (value) => {
    setSearchQuery(value)
    setCurrentPage(1)
  }

  // Render status
  const renderStatus = (status) => {
    const statusConfig = {
      active: { color: 'green', text: 'Active' },
      inactive: { color: 'orange', text: 'Inactive' },
      suspended: { color: 'red', text: 'Suspended' },
    }
    
    const config = statusConfig[status] || { color: 'default', text: status }
    return <Tag color={config.color}>{config.text}</Tag>
  }

  // Render reliability score
  const renderReliabilityScore = (score) => {
    const color = score >= 8 ? '#52c41a' : score >= 6 ? '#faad14' : '#ff4d4f'
    return (
      <Space>
        <Rate disabled value={score / 2} style={{ fontSize: 14 }} />
        <Text style={{ color }}>{score}/10</Text>
      </Space>
    )
  }

  // Table columns
  const columns = [
    {
      title: 'Company',
      dataIndex: 'company_name',
      key: 'company_name',
      render: (text, record) => (
        <Space>
          <UserOutlined style={{ color: '#1890ff' }} />
          <div>
            <Text strong>{text}</Text>
            {record.contact_person && (
              <div><Text type="secondary">{record.contact_person}</Text></div>
            )}
          </div>
        </Space>
      ),
    },
    {
      title: 'Contact',
      key: 'contacts',
      render: (record) => (
        <div>
          <div>{record.email}</div>
          {record.phone && <Text type="secondary">{record.phone}</Text>}
        </div>
      ),
    },
    {
      title: 'Address',
      key: 'address',
      render: (record) => {
        const addr = record.address
        if (!addr?.city && !addr?.state) return '-'
        return `${addr.city || ''}, ${addr.state || ''}`
      },
    },
    {
      title: 'Credit Limit',
      dataIndex: 'credit_limit',
      key: 'credit_limit',
      align: 'right',
      render: (value) => value ? `$${value.toLocaleString()}` : '-',
    },
    {
      title: 'Rating',
      dataIndex: 'reliability_score',
      key: 'reliability_score',
      align: 'center',
      render: renderReliabilityScore,
    },
    {
      title: 'Status',
      dataIndex: 'status',
      key: 'status',
      align: 'center',
      render: renderStatus,
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
              title="Delete broker?"
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
                Brokers
              </Title>
              <Text type="secondary">
                Manage partner companies
              </Text>
            </Col>
            <Col>
              <Button
                type="primary"
                icon={<PlusOutlined />}
                onClick={handleCreate}
                size="large"
              >
                Add Broker
              </Button>
            </Col>
          </Row>
        </div>

        <div style={{ marginBottom: 16 }}>
          <Search
            placeholder="Search by company name, email or phone..."
            allowClear
            enterButton={<SearchOutlined />}
            size="large"
            onSearch={handleSearch}
            style={{ maxWidth: 400 }}
          />
        </div>

        <Table
          columns={columns}
          dataSource={brokersData?.data || []}
          rowKey="id"
          loading={isLoading}
          pagination={{
            current: currentPage,
            pageSize: pageSize,
            total: brokersData?.pagination?.total || 0,
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
            ? 'Add Broker'
            : modalMode === 'edit'
            ? 'Edit Broker'
            : 'Broker Information'
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
            <Col span={24}>
              <Form.Item
                label="Company Name"
                name="company_name"
                rules={[{ required: true, message: 'Please enter company name' }]}
              >
                <Input placeholder="ABC Transport Services" />
              </Form.Item>
            </Col>
          </Row>

          <Row gutter={16}>
            <Col span={12}>
              <Form.Item
                label="Contact Person"
                name="contact_person"
              >
                <Input placeholder="John Smith" />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item
                label="Email"
                name="email"
                rules={[
                  { required: true, message: 'Please enter email' },
                  { type: 'email', message: 'Invalid email format' }
                ]}
              >
                <Input placeholder="contact@company.com" />
              </Form.Item>
            </Col>
          </Row>

          <Row gutter={16}>
            <Col span={12}>
              <Form.Item
                label="Phone"
                name="phone"
              >
                <Input placeholder="+1 (555) 123-4567" />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item
                label="Status"
                name="status"
                initialValue="active"
              >
                <Select>
                  <Option value="active">Active</Option>
                  <Option value="inactive">Inactive</Option>
                  <Option value="suspended">Suspended</Option>
                </Select>
              </Form.Item>
            </Col>
          </Row>

          <Divider>Address</Divider>

          <Row gutter={16}>
            <Col span={24}>
              <Form.Item
                label="Street"
                name="street"
              >
                <Input placeholder="123 Main Street" />
              </Form.Item>
            </Col>
          </Row>

          <Row gutter={16}>
            <Col span={8}>
              <Form.Item
                label="City"
                name="city"
              >
                <Input placeholder="New York" />
              </Form.Item>
            </Col>
            <Col span={8}>
              <Form.Item
                label="State/Province"
                name="state"
              >
                <Input placeholder="NY" />
              </Form.Item>
            </Col>
            <Col span={8}>
              <Form.Item
                label="Country"
                name="country"
                initialValue="USA"
              >
                <Input placeholder="USA" />
              </Form.Item>
            </Col>
          </Row>

          <Row gutter={16}>
            <Col span={8}>
              <Form.Item
                label="ZIP Code"
                name="zip_code"
              >
                <Input placeholder="10001" />
              </Form.Item>
            </Col>
          </Row>

          <Divider>Financial Parameters</Divider>

          <Row gutter={16}>
            <Col span={12}>
              <Form.Item
                label="Credit Limit ($)"
                name="credit_limit"
                initialValue={0}
              >
                <InputNumber
                  style={{ width: '100%' }}
                  min={0}
                  step={1000}
                  formatter={value => `$ ${value}`.replace(/\B(?=(\d{3})+(?!\d))/g, ',')}
                  parser={value => value.replace(/\$\s?|(,*)/g, '')}
                />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item
                label="Reliability Score"
                name="reliability_score"
                initialValue={5}
              >
                <InputNumber
                  style={{ width: '100%' }}
                  min={0}
                  max={10}
                  step={1}
                />
              </Form.Item>
            </Col>
          </Row>

          <Form.Item
            label="Notes"
            name="notes"
          >
            <Input.TextArea
              rows={3}
              placeholder="Additional information about the broker..."
            />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}

export default Brokers 