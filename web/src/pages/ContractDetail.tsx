import { useState, useEffect, useCallback } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import {
  Card, Descriptions, Tag, Button, Table, Typography, message, Spin, Space, Empty,
} from 'antd'
import { ArrowLeftOutlined, ReloadOutlined } from '@ant-design/icons'
import { getContractList, getTokenInfo, getEvents } from '../api'
import type { ContractEntry, EventEntry, TokenInfo } from '../types'

const { Title } = Typography

const ContractDetail = () => {
  const { address } = useParams<{ address: string }>()
  const navigate = useNavigate()
  const [contract, setContract] = useState<ContractEntry | null>(null)
  const [tokenInfo, setTokenInfo] = useState<TokenInfo | null>(null)
  const [events, setEvents] = useState<EventEntry[]>([])
  const [loading, setLoading] = useState(false)
  const [eventsLoading, setEventsLoading] = useState(false)

  const fetchContract = useCallback(async () => {
    if (!address) return
    setLoading(true)
    try {
      const res = await getContractList()
      const list = res.data.contracts || []
      const found = list.find((c: ContractEntry) => c.address.toLowerCase() === address.toLowerCase())
      setContract(found || null)
    } catch {
      message.error('获取合约信息失败')
    } finally {
      setLoading(false)
    }
  }, [address])

  const fetchTokenInfo = useCallback(async () => {
    try {
      const res = await getTokenInfo(address)
      setTokenInfo(res.data)
    } catch {
      // token info might not be available
    }
  }, [address])

  const fetchEvents = useCallback(async () => {
    if (!address) return
    setEventsLoading(true)
    try {
      const res = await getEvents({ address, limit: 50, offset: 0 })
      setEvents(res.data.events || [])
    } catch {
      // events might not be available
    } finally {
      setEventsLoading(false)
    }
  }, [address])

  useEffect(() => {
    fetchContract()
    fetchTokenInfo()
    fetchEvents()
  }, [fetchContract, fetchTokenInfo, fetchEvents])

  if (loading) {
    return (
      <div style={{ textAlign: 'center', padding: 100 }}>
        <Spin size="large" />
      </div>
    )
  }

  if (!contract) {
    return (
      <div style={{ textAlign: 'center', padding: 100 }}>
        <Empty description="合约未找到" />
        <Button style={{ marginTop: 16 }} onClick={() => navigate('/manage/contracts')}>
          返回列表
        </Button>
      </div>
    )
  }

  const eventColumns = [
    {
      title: '交易哈希',
      dataIndex: 'txHash',
      key: 'txHash',
      ellipsis: true,
      width: 220,
      render: (text: string) => (
        <Typography.Text copyable style={{ fontSize: 12 }}>
          {`${text.slice(0, 10)}...${text.slice(-8)}`}
        </Typography.Text>
      ),
    },
    {
      title: '发送方',
      dataIndex: 'fromAddr',
      key: 'fromAddr',
      ellipsis: true,
      width: 200,
      render: (text: string) => (
        <Typography.Text style={{ fontSize: 12 }}>
          {`${text.slice(0, 8)}...${text.slice(-6)}`}
        </Typography.Text>
      ),
    },
    {
      title: '接收方',
      dataIndex: 'toAddr',
      key: 'toAddr',
      ellipsis: true,
      width: 200,
      render: (text: string) => (
        <Typography.Text style={{ fontSize: 12 }}>
          {`${text.slice(0, 8)}...${text.slice(-6)}`}
        </Typography.Text>
      ),
    },
    {
      title: '数量',
      dataIndex: 'value',
      key: 'value',
      width: 160,
    },
    {
      title: '区块号',
      dataIndex: 'blockNumber',
      key: 'blockNumber',
      width: 100,
    },
    {
      title: '时间',
      dataIndex: 'createdAt',
      key: 'createdAt',
      width: 180,
      render: (text: string) => new Date(text).toLocaleString(),
    },
  ]

  return (
    <div>
      <Space style={{ marginBottom: 16 }}>
        <Button icon={<ArrowLeftOutlined />} onClick={() => navigate('/manage/contracts')}>
          返回列表
        </Button>
      </Space>

      <Title level={3} style={{ marginTop: 8 }}>
        {contract.name || '未命名'} ({contract.symbol || '-'})
      </Title>

      <Card title="基本信息" style={{ marginBottom: 16 }}>
        <Descriptions column={{ xs: 1, sm: 2 }} bordered size="small">
          <Descriptions.Item label="合约地址">
            <Typography.Text copyable>{contract.address}</Typography.Text>
          </Descriptions.Item>
          <Descriptions.Item label="网络">{contract.network}</Descriptions.Item>
          <Descriptions.Item label="名称">{contract.name || '-'}</Descriptions.Item>
          <Descriptions.Item label="符号">{contract.symbol || '-'}</Descriptions.Item>
          <Descriptions.Item label="状态">
            {contract.isActive ? <Tag color="green">监听中</Tag> : <Tag>非活跃</Tag>}
          </Descriptions.Item>
          <Descriptions.Item label="部署者">
            {contract.deployer ? (
              <Typography.Text copyable>{contract.deployer}</Typography.Text>
            ) : '-'}
          </Descriptions.Item>
          <Descriptions.Item label="部署交易哈希" span={2}>
            {contract.txHash ? (
              <Typography.Text copyable style={{ fontSize: 12 }}>{contract.txHash}</Typography.Text>
            ) : '-'}
          </Descriptions.Item>
          <Descriptions.Item label="创建时间" span={2}>
            {new Date(contract.createdAt).toLocaleString()}
          </Descriptions.Item>
        </Descriptions>
      </Card>

      {tokenInfo && (
        <Card title="代币信息" style={{ marginBottom: 16 }}>
          <Descriptions column={{ xs: 1, sm: 2 }} bordered size="small">
            <Descriptions.Item label="名称">{tokenInfo.name}</Descriptions.Item>
            <Descriptions.Item label="符号">{tokenInfo.symbol}</Descriptions.Item>
            <Descriptions.Item label="精度">{tokenInfo.decimals}</Descriptions.Item>
            <Descriptions.Item label="总供应量">{tokenInfo.totalSupply}</Descriptions.Item>
          </Descriptions>
        </Card>
      )}

      <Card
        title="事件记录"
        extra={
          <Button icon={<ReloadOutlined />} onClick={fetchEvents} loading={eventsLoading} size="small">
            刷新
          </Button>
        }
      >
        <Table
          columns={eventColumns}
          dataSource={events}
          rowKey="id"
          loading={eventsLoading}
          pagination={{ pageSize: 10 }}
          scroll={{ x: 1000 }}
          locale={{ emptyText: '暂无事件记录' }}
        />
      </Card>
    </div>
  )
}

export default ContractDetail
