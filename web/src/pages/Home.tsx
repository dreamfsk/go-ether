import { useState, useEffect, useCallback } from 'react'
import { useNavigate } from 'react-router-dom'
import {
  Card, Button, Modal, Form, Input, InputNumber, message, Typography, Row, Col, Statistic, Tag, Skeleton,
} from 'antd'
import {
  RocketOutlined,
  UnorderedListOutlined,
  ApiOutlined,
  CheckCircleOutlined,
  DatabaseOutlined,
} from '@ant-design/icons'
import { deployContract, getContractList, getConfig } from '../api'
import type { ContractEntry } from '../types'

const { Title, Paragraph, Text } = Typography

const Home = () => {
  const navigate = useNavigate()
  const [deployOpen, setDeployOpen] = useState(false)
  const [deploying, setDeploying] = useState(false)
  const [statsLoading, setStatsLoading] = useState(true)
  const [contracts, setContracts] = useState<ContractEntry[]>([])
  const [currentAddress, setCurrentAddress] = useState('')
  const [maxSupply, setMaxSupply] = useState<string>('')
  const [form] = Form.useForm()

  const fetchStats = useCallback(async () => {
    setStatsLoading(true)
    try {
      const res = await getContractList()
      setContracts(res.data.contracts || [])
      setCurrentAddress(res.data.currentAddress || '')
    } catch {
      // silent fail for stats
    } finally {
      setStatsLoading(false)
    }
  }, [])

  useEffect(() => {
    fetchStats()
  }, [fetchStats])

  // 打开部署弹窗时获取配置
  useEffect(() => {
    if (deployOpen && !maxSupply) {
      getConfig().then(res => setMaxSupply(res.data.maxTokenAmount))
    }
  }, [deployOpen, maxSupply])

  // 校验供应量是否超过最大值（转最小单位后比较）
  const validateSupply = () => {
    const decimals = form.getFieldValue('decimals') || 18
    const supply = form.getFieldValue('initialSupply')
    if (!supply || !maxSupply) return

    const supplyWei = BigInt(supply) * (BigInt(10) ** BigInt(decimals))
    const maxWei = BigInt(maxSupply)

    if (supplyWei > maxWei) {
      message.error('供应量超出最大允许值')
    }
  }

  const handleDeploy = async () => {
    try {
      const values = await form.validateFields()
      setDeploying(true)

      // 转换: 供应量 × 10^decimals = 最小单位
      const decimals = values.decimals || 18
      const supply = BigInt(values.initialSupply)
      const multiplier = BigInt(10) ** BigInt(decimals)
      const initialSupplyWei = (supply * multiplier).toString()

      const res = await deployContract({
        name: values.name,
        symbol: values.symbol,
        initialSupply: initialSupplyWei,
        recipient: values.recipient || '',
      })
      message.success(`合约部署成功！地址: ${res.data.address}`)
      setDeployOpen(false)
      form.resetFields()
      fetchStats()
    } catch (err: unknown) {
      if (err && typeof err === 'object' && 'errorFields' in err) {
        return
      }
      message.error('部署合约失败')
    } finally {
      setDeploying(false)
    }
  }

  const activeContract = contracts.find(c => c.address === currentAddress)

  return (
    <div>
      <div style={{ textAlign: 'center', marginBottom: 32 }}>
        <ApiOutlined style={{ fontSize: 56, color: '#1677ff', marginBottom: 12 }} />
        <Title level={2} style={{ marginBottom: 8 }}>Go-Ether 管理面板</Title>
        <Paragraph type="secondary" style={{ fontSize: 16 }}>
          管理 ERC-20 合约，监控 Transfer 事件，轻松部署和切换合约
        </Paragraph>
      </div>

      {statsLoading ? (
        <Skeleton active paragraph={{ rows: 2 }} style={{ maxWidth: 900, margin: '0 auto 32px' }} />
      ) : (
        <Row gutter={[16, 16]} style={{ maxWidth: 900, margin: '0 auto 32px' }}>
          <Col xs={24} sm={8}>
            <Card size="small">
              <Statistic
                title="合约总数"
                value={contracts.length}
                prefix={<DatabaseOutlined />}
              />
            </Card>
          </Col>
          <Col xs={24} sm={8}>
            <Card size="small">
              <Statistic
                title="活跃合约"
                value={contracts.filter(c => c.isActive).length}
                prefix={<CheckCircleOutlined />}
                valueStyle={{ color: '#3f8600' }}
              />
            </Card>
          </Col>
          <Col xs={24} sm={8}>
            <Card size="small">
              <Statistic
                title="当前监听"
                value={currentAddress ? '已连接' : '未设置'}
                prefix={currentAddress ? <CheckCircleOutlined /> : undefined}
                valueStyle={{ color: currentAddress ? '#3f8600' : '#cf1322' }}
              />
            </Card>
          </Col>
        </Row>
      )}

      {currentAddress && activeContract && (
        <Card
          size="small"
          style={{ maxWidth: 900, margin: '0 auto 24px', background: '#f6ffed', border: '1px solid #b7eb8f' }}
        >
          <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', flexWrap: 'wrap', gap: 8 }}>
            <div>
              <Text strong>当前监听合约：</Text>
              <Tag color="green" style={{ marginLeft: 8 }}>
                {activeContract.name || activeContract.symbol || '未命名'}
              </Tag>
              <Text type="secondary" copyable style={{ fontSize: 12 }}>
                {activeContract.address}
              </Text>
            </div>
            <Button
              size="small"
              onClick={() => navigate(`/manage/contracts/${activeContract.address}`)}
            >
              查看详情
            </Button>
          </div>
        </Card>
      )}

      <Row gutter={[24, 24]} style={{ maxWidth: 900, margin: '0 auto' }}>
        <Col xs={24} md={12}>
          <Card
            hoverable
            style={{ textAlign: 'center' }}
            onClick={() => setDeployOpen(true)}
          >
            <RocketOutlined style={{ fontSize: 40, color: '#1677ff', marginBottom: 16 }} />
            <Title level={4}>部署合约</Title>
            <Paragraph type="secondary">
              部署一个新的 MyERC20 合约，部署后自动加入监听
            </Paragraph>
            <Button type="primary" icon={<RocketOutlined />}>
              开始部署
            </Button>
          </Card>
        </Col>

        <Col xs={24} md={12}>
          <Card
            hoverable
            style={{ textAlign: 'center' }}
            onClick={() => navigate('/manage/contracts')}
          >
            <UnorderedListOutlined style={{ fontSize: 40, color: '#52c41a', marginBottom: 16 }} />
            <Title level={4}>合约列表</Title>
            <Paragraph type="secondary">
              查看所有部署的合约，管理监听状态
            </Paragraph>
            <Button icon={<UnorderedListOutlined />}>
              查看列表
            </Button>
          </Card>
        </Col>
      </Row>

      <Modal
        title="部署 MyERC20 合约"
        open={deployOpen}
        onCancel={() => { setDeployOpen(false); form.resetFields() }}
        onOk={handleDeploy}
        confirmLoading={deploying}
        okText="部署"
        cancelText="取消"
        destroyOnClose
      >
        <Form form={form} layout="vertical" style={{ marginTop: 16 }}>
          <Form.Item
            label="代币名称"
            name="name"
            rules={[{ required: true, message: '请输入代币名称' }]}
          >
            <Input placeholder="例如: MyToken" />
          </Form.Item>
          <Form.Item
            label="代币符号"
            name="symbol"
            rules={[{ required: true, message: '请输入代币符号' }]}
          >
            <Input placeholder="例如: MTK" />
          </Form.Item>
          <Form.Item
            label="初始供应量"
            required
            style={{ marginBottom: 0 }}
          >
            <Input.Group compact style={{ display: 'flex' }}>
              <Form.Item
                name="initialSupply"
                noStyle
                rules={[{ required: true, message: '请输入初始供应量' }]}
              >
                <InputNumber
                  style={{ flex: 1, minWidth: 120 }}
                  placeholder="例如: 1000000"
                  min={0}
                  stringMode
                  onBlur={validateSupply}
                />
              </Form.Item>
              <Form.Item
                name="decimals"
                noStyle
                initialValue={18}
              >
                <InputNumber
                  style={{ width: 80 }}
                  min={0}
                  max={18}
                  defaultValue={18}
                  controls={false}
                  onBlur={validateSupply}
                />
              </Form.Item>
            </Input.Group>
          </Form.Item>
          <div style={{ color: 'rgba(0,0,0,0.45)', fontSize: 12, marginBottom: 16 }}>
            实际发送数量 = 供应量 × 10^decimals，当前 decimals 默认 18
          </div>
          <Form.Item
            label="接收地址（可选，默认为部署者）"
            name="recipient"
          >
            <Input placeholder="0x...（留空则默认发送给部署者）" />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}

export default Home
