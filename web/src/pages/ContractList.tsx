import { useState, useEffect, useCallback } from 'react'
import { Table, Tag, Button, message, Space, Typography, Popconfirm } from 'antd'
import { ReloadOutlined, SwapOutlined, InfoCircleOutlined } from '@ant-design/icons'
import { useNavigate } from 'react-router-dom'
import { getContractList, switchContract } from '../api'
import type { ContractEntry } from '../types'

const { Title } = Typography

const ContractList = () => {
  const navigate = useNavigate()
  const [contracts, setContracts] = useState<ContractEntry[]>([])
  const [currentAddress, setCurrentAddress] = useState('')
  const [loading, setLoading] = useState(false)
  const [switching, setSwitching] = useState<string | null>(null)

  const fetchContracts = useCallback(async () => {
    setLoading(true)
    try {
      const res = await getContractList()
      setContracts(res.data.contracts || [])
      setCurrentAddress(res.data.currentAddress || '')
    } catch {
      message.error('获取合约列表失败')
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    fetchContracts()
  }, [fetchContracts])

  const handleSwitch = async (address: string) => {
    setSwitching(address)
    try {
      await switchContract({ address })
      message.success(`已切换到合约: ${address}`)
      fetchContracts()
    } catch {
      message.error('切换合约失败')
    } finally {
      setSwitching(null)
    }
  }

  const columns = [
    {
      title: '名称',
      dataIndex: 'name',
      key: 'name',
      render: (text: string) => text || '-',
    },
    {
      title: '符号',
      dataIndex: 'symbol',
      key: 'symbol',
      render: (text: string) => text || '-',
    },
    {
      title: '地址',
      dataIndex: 'address',
      key: 'address',
      ellipsis: true,
      width: 280,
      render: (text: string) => (
        <Typography.Text copyable style={{ fontSize: 12 }}>
          {text}
        </Typography.Text>
      ),
    },
    {
      title: '网络',
      dataIndex: 'network',
      key: 'network',
      width: 100,
    },
    {
      title: '状态',
      dataIndex: 'isActive',
      key: 'isActive',
      width: 100,
      render: (isActive: boolean, record: ContractEntry) =>
        record.address === currentAddress ? (
          <Tag color="green">监听中</Tag>
        ) : isActive ? (
          <Tag color="blue">活跃</Tag>
        ) : (
          <Tag>非活跃</Tag>
        ),
    },
    {
      title: '部署时间',
      dataIndex: 'createdAt',
      key: 'createdAt',
      width: 180,
      render: (text: string) => new Date(text).toLocaleString(),
    },
    {
      title: '操作',
      key: 'action',
      width: 200,
      render: (_: unknown, record: ContractEntry) => (
        <Space>
          <Button
            type="link"
            size="small"
            icon={<InfoCircleOutlined />}
            onClick={() => navigate(`/manage/contracts/${record.address}`)}
          >
            详情
          </Button>
          {record.address !== currentAddress && (
            <Popconfirm
              title="确认切换监听合约？"
              description={`将切换到合约: ${record.address}`}
              onConfirm={() => handleSwitch(record.address)}
              okText="确认"
              cancelText="取消"
            >
              <Button
                type="primary"
                size="small"
                icon={<SwapOutlined />}
                loading={switching === record.address}
                danger
              >
                切换监听
              </Button>
            </Popconfirm>
          )}
        </Space>
      ),
    },
  ]

  return (
    <div>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 16 }}>
        <Title level={3} style={{ margin: 0 }}>
          合约列表
        </Title>
        <Button icon={<ReloadOutlined />} onClick={fetchContracts} loading={loading}>
          刷新
        </Button>
      </div>
      <Table
        columns={columns}
        dataSource={contracts}
        rowKey="address"
        loading={loading}
        pagination={{ pageSize: 10 }}
        scroll={{ x: 1000 }}
      />
    </div>
  )
}

export default ContractList
