import { LockOutlined, MailOutlined, TeamOutlined, UserOutlined } from '@ant-design/icons'
import { Alert, Button, Card, Form, Input, Select, Space, Typography } from 'antd'
import React, { useState } from 'react'
import { useMutation } from 'react-query'
import { useNavigate } from 'react-router-dom'
import { useAuth } from '../contexts/AuthContext'
import { authApi } from '../services/api'

const { Title } = Typography
const { Option } = Select

const Register = () => {
  const [form] = Form.useForm()
  const navigate = useNavigate()
  const { login } = useAuth()
  const [error, setError] = useState('')

  const registerMutation = useMutation(authApi.register, {
    onSuccess: (data) => {
      login(data.data.token, data.data.user)
      navigate('/')
    },
    onError: (error) => {
      setError(error.response?.data?.error || 'Ошибка регистрации')
    }
  })

  const onFinish = (values) => {
    setError('')
    registerMutation.mutate(values)
  }

  return (
    <div style={{
      display: 'flex',
      justifyContent: 'center',
      alignItems: 'center',
      minHeight: '100vh',
      backgroundColor: '#f0f2f5'
    }}>
      <Card style={{ width: 450, boxShadow: '0 4px 8px rgba(0,0,0,0.1)' }}>
        <Space direction="vertical" size="large" style={{ width: '100%' }}>
          <div style={{ textAlign: 'center' }}>
            <Title level={2}>Регистрация</Title>
            <Typography.Text type="secondary">
              Создание нового аккаунта
            </Typography.Text>
          </div>

          {error && (
            <Alert
              message={error}
              type="error"
              showIcon
              closable
              onClose={() => setError('')}
            />
          )}

          <Form
            form={form}
            name="register"
            onFinish={onFinish}
            autoComplete="off"
            layout="vertical"
          >
            <Form.Item
              name="username"
              label="Имя пользователя"
              rules={[
                { required: true, message: 'Введите имя пользователя!' },
                { min: 3, message: 'Минимум 3 символа!' },
                { max: 50, message: 'Максимум 50 символов!' }
              ]}
            >
              <Input
                prefix={<UserOutlined />}
                placeholder="Имя пользователя"
                size="large"
              />
            </Form.Item>

            <Form.Item
              name="email"
              label="Email"
              rules={[
                { required: true, message: 'Введите email!' },
                { type: 'email', message: 'Некорректный email!' }
              ]}
            >
              <Input
                prefix={<MailOutlined />}
                placeholder="Email"
                size="large"
              />
            </Form.Item>

            <Form.Item
              name="password"
              label="Пароль"
              rules={[
                { required: true, message: 'Введите пароль!' },
                { min: 6, message: 'Минимум 6 символов!' }
              ]}
            >
              <Input.Password
                prefix={<LockOutlined />}
                placeholder="Пароль"
                size="large"
              />
            </Form.Item>

            <Form.Item
              name="confirmPassword"
              label="Подтверждение пароля"
              dependencies={['password']}
              rules={[
                { required: true, message: 'Подтвердите пароль!' },
                ({ getFieldValue }) => ({
                  validator(_, value) {
                    if (!value || getFieldValue('password') === value) {
                      return Promise.resolve()
                    }
                    return Promise.reject(new Error('Пароли не совпадают!'))
                  },
                }),
              ]}
            >
              <Input.Password
                prefix={<LockOutlined />}
                placeholder="Подтверждение пароля"
                size="large"
              />
            </Form.Item>

            <Form.Item
              name="role"
              label="Роль"
              rules={[
                { required: true, message: 'Выберите роль!' }
              ]}
            >
              <Select
                placeholder="Выберите роль"
                size="large"
                prefix={<TeamOutlined />}
              >
                <Option value="user">Пользователь</Option>
                <Option value="admin">Администратор</Option>
              </Select>
            </Form.Item>

            <Form.Item>
              <Button
                type="primary"
                htmlType="submit"
                size="large"
                loading={registerMutation.isLoading}
                style={{ width: '100%' }}
              >
                Зарегистрироваться
              </Button>
            </Form.Item>
          </Form>

          <div style={{ textAlign: 'center' }}>
            <Typography.Text type="secondary">
              Уже есть аккаунт?{' '}
              <Button 
                type="link" 
                onClick={() => navigate('/login')}
                style={{ padding: 0 }}
              >
                Войти
              </Button>
            </Typography.Text>
          </div>
        </Space>
      </Card>
    </div>
  )
}

export default Register 