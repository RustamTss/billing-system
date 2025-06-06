import { LockOutlined, UserOutlined } from '@ant-design/icons'
import { Alert, Button, Card, Form, Input, Space, Typography } from 'antd'
import React, { useState } from 'react'
import { useMutation } from 'react-query'
import { useNavigate } from 'react-router-dom'
import { useAuth } from '../contexts/AuthContext'
import { authApi } from '../services/api'

const { Title } = Typography

const Login = () => {
  const [form] = Form.useForm()
  const navigate = useNavigate()
  const { login } = useAuth()
  const [error, setError] = useState('')

  const loginMutation = useMutation(authApi.login, {
    onSuccess: (data) => {
      login(data.data.token, data.data.user)
      navigate('/')
    },
    onError: (error) => {
      setError(error.response?.data?.error || 'Ошибка авторизации')
    }
  })

  const onFinish = (values) => {
    setError('')
    loginMutation.mutate(values)
  }

  return (
    <div style={{
      display: 'flex',
      justifyContent: 'center',
      alignItems: 'center',
      minHeight: '100vh',
      backgroundColor: '#f0f2f5'
    }}>
      <Card style={{ width: 400, boxShadow: '0 4px 8px rgba(0,0,0,0.1)' }}>
        <Space direction="vertical" size="large" style={{ width: '100%' }}>
          <div style={{ textAlign: 'center' }}>
            <Title level={2}>Вход в систему</Title>
            <Typography.Text type="secondary">
              Billing Management System
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
            name="login"
            onFinish={onFinish}
            autoComplete="off"
            layout="vertical"
          >
            <Form.Item
              name="username"
              label="Имя пользователя"
              rules={[
                { required: true, message: 'Введите имя пользователя!' }
              ]}
            >
              <Input
                prefix={<UserOutlined />}
                placeholder="Имя пользователя"
                size="large"
              />
            </Form.Item>

            <Form.Item
              name="password"
              label="Пароль"
              rules={[
                { required: true, message: 'Введите пароль!' }
              ]}
            >
              <Input.Password
                prefix={<LockOutlined />}
                placeholder="Пароль"
                size="large"
              />
            </Form.Item>

            <Form.Item>
              <Button
                type="primary"
                htmlType="submit"
                size="large"
                loading={loginMutation.isLoading}
                style={{ width: '100%' }}
              >
                Войти
              </Button>
            </Form.Item>
          </Form>

          <div style={{ textAlign: 'center' }}>
            <Typography.Text type="secondary">
              Нет аккаунта?{' '}
              <Button 
                type="link" 
                onClick={() => navigate('/register')}
                style={{ padding: 0 }}
              >
                Зарегистрироваться
              </Button>
            </Typography.Text>
          </div>
        </Space>
      </Card>
    </div>
  )
}

export default Login 