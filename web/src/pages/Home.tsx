import { useState, useEffect, useCallback } from 'react'
import { useNavigate } from 'react-router-dom'
import {
  Card, Button, Modal, Form, Input, InputNumber, Select, message, Typography, Row, Col, Statistic, Tag, Skeleton,
  Table, Empty, Tooltip, Divider,
} from 'antd'
import {
  RocketOutlined,
  UnorderedListOutlined,
  ApiOutlined,
  CheckCircleOutlined,
  DatabaseOutlined,
  SendOutlined,
  SwapOutlined,
  CopyOutlined,
  InfoCircleOutlined,
} from '@ant-design/icons'
import { deployContract, getContractList, getConfig, getTxHistory, sendTransaction, getGasFeeSuggestion, estimateGas, tokenTransfer } from '../api'
import type { ContractEntry, TxHistoryEntry, GasFeeSuggestion } from '../types'

const { Title, Paragraph, Text } = Typography
const { Option } = Select

// 单位转换工具函数
const parseToWei = (value: string, unit: string): bigint => {
  if (!value || value === '0') return BigInt(0)
  
  const decimals = unit === 'eth' ? 18 : unit === 'gwei' ? 9 : 0
  
  // 处理小数
  const parts = value.split('.')
  const integerPart = parts[0] || '0'
  const decimalPart = parts[1] || ''
  
  // 限制小数位数并补齐
  const trimmedDecimal = decimalPart.slice(0, decimals).padEnd(decimals, '0')
  
  // 直接拼接整数和小数部分，得到 wei 值
  return BigInt(integerPart + trimmedDecimal)
}

const Home = () => {
  const navigate = useNavigate()
  const [deployOpen, setDeployOpen] = useState(false)
  const [deploying, setDeploying] = useState(false)
  const [statsLoading, setStatsLoading] = useState(true)
  const [contracts, setContracts] = useState<ContractEntry[]>([])
  const [currentAddress, setCurrentAddress] = useState('')
  const [maxSupply, setMaxSupply] = useState<string>('')
  const [form] = Form.useForm()

  // ============ 交易模块状态 ============
  const [txHistory, setTxHistory] = useState<TxHistoryEntry[]>([])
  const [txLoading, setTxLoading] = useState(true)
  const [sendOpen, setSendOpen] = useState(false)
  const [sending, setSending] = useState(false)
  const [sendForm] = Form.useForm()
  const [displayWei, setDisplayWei] = useState<string>('-')
  const [gasFeeSuggestion, setGasFeeSuggestion] = useState<GasFeeSuggestion | null>(null)
  const [gasFeeLoading, setGasFeeLoading] = useState(false)
  const [gasPriceWei, setGasPriceWei] = useState<string>('0')
  const [gasLimit, setGasLimit] = useState<string>('21000')
  const [estimatingGas, setEstimatingGas] = useState(false)
  const [txType, setTxType] = useState<'eth' | 'erc20'>('eth')

  // 计算最终交易额（金额 + gas 费用）
  const calculateTotalCost = () => {
    if (displayWei === '-' || displayWei === '无效输入') return null
    try {
      const valueWei = BigInt(displayWei)
      const gasPrice = BigInt(gasPriceWei || '0')
      const gas = BigInt(gasLimit || '21000')
      const gasCost = gasPrice * gas
      const total = valueWei + gasCost
      return {
        valueWei: valueWei.toString(),
        gasCost: gasCost.toString(),
        total: total.toString(),
        totalETH: formatETH(total.toString())
      }
    } catch {
      return null
    }
  }

  // 格式化 wei 为 ETH
  const formatETH = (wei: string): string => {
    try {
      const weiBig = BigInt(wei)
      const ethDivisor = BigInt(10) ** BigInt(18)
      
      // 整数部分
      const intPart = weiBig / ethDivisor
      // 小数部分（余数）
      const remainder = weiBig % ethDivisor
      
      // 格式化小数部分为 18 位字符串
      let decimalStr = remainder.toString().padStart(18, '0')
      // 移除末尾的 0
      decimalStr = decimalStr.replace(/0+$/, '')
      
      if (decimalStr === '') {
        return intPart.toString()
      }
      return `${intPart}.${decimalStr}`
    } catch {
      return wei
    }
  }

  // 监听表单字段变化，实时计算 wei 值
  const handleValuesChange = (_: any, allValues: { 
    value?: number | string
    unit?: string
    gasPrice?: number | string
    gasLimit?: number | string
    to?: string
  }) => {
    // 计算金额 wei
    const { value, unit = 'eth', gasPrice, to } = allValues
    let calculatedWei = '-'
    
    if (value === undefined || value === null || value === '') {
      setDisplayWei('-')
    } else {
      try {
        const wei = parseToWei(String(value), unit)
        calculatedWei = wei.toString()
        setDisplayWei(calculatedWei)
      } catch (e) {
        console.error('parseToWei error:', e)
        setDisplayWei('无效输入')
      }
    }

    // 计算 gas price wei（固定使用 Gwei 单位）
    if (gasPrice === undefined || gasPrice === null || gasPrice === '') {
      setGasPriceWei('0')
    } else {
      try {
        const gpWei = parseToWei(String(gasPrice), 'gwei')
        setGasPriceWei(gpWei.toString())
      } catch (e) {
        console.error('parseToWei gasPrice error:', e)
        setGasPriceWei('0')
      }
    }

    // 更新 gas limit
    if (allValues.gasLimit !== undefined) {
      setGasLimit(String(allValues.gasLimit || '21000'))
    }

    // 自动估算 gas（当有接收地址和金额时）
    if (to && calculatedWei !== '-' && calculatedWei !== '无效输入') {
      estimateGasAuto(to, calculatedWei)
    }
  }

  // 自动估算 gas
  const estimateGasAuto = async (to: string, valueWei: string) => {
    console.log('estimateGasAuto called:', { to, valueWei, toLength: to.length })
    if (!to || !valueWei || to.length !== 42) {
      console.log('estimateGasAuto skipped: invalid params')
      return
    }
    setEstimatingGas(true)
    try {
      console.log('Calling estimateGas API...')
      const res = await estimateGas(to, valueWei)
      console.log('estimateGas result:', res.data)
      const estimatedGas = res.data.gas
      setGasLimit(String(estimatedGas))
      sendForm.setFieldValue('gasLimit', estimatedGas)
    } catch (e) {
      console.error('estimateGas error:', e)
    } finally {
      setEstimatingGas(false)
    }
  }

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

  const fetchTxHistory = useCallback(async () => {
    setTxLoading(true)
    try {
      const res = await getTxHistory({ limit: 5 })
      setTxHistory(res.data || [])
    } catch {
      // silent fail
    } finally {
      setTxLoading(false)
    }
  }, [])

    // 打开发送交易弹窗时获取 gas fee 建议
  const fetchGasFeeSuggestion = useCallback(async () => {
    setGasFeeLoading(true)
    try {
      const res = await getGasFeeSuggestion()
      setGasFeeSuggestion(res.data)
    } catch (err) {
      console.error('Failed to fetch gas fee suggestion:', err)
      message.error('获取 Gas 费用建议失败')
    } finally {
      setGasFeeLoading(false)
    }
  }, [])

  useEffect(() => {
    fetchStats()
  }, [fetchStats])

  useEffect(() => {
    fetchTxHistory()
  }, [fetchTxHistory])

  // 打开部署弹窗时获取配置
  useEffect(() => {
    if (deployOpen && !maxSupply) {
      getConfig().then(res => setMaxSupply(res.data.maxTokenAmount))
    }
  }, [deployOpen, maxSupply])

  // 打开发送交易弹窗时获取 gas fee 建议
  useEffect(() => {
    if (sendOpen) {
      fetchGasFeeSuggestion()
    }
  }, [sendOpen, fetchGasFeeSuggestion])

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
      fetchTxHistory() // 部署成功后刷新交易列表
    } catch (err: unknown) {
      if (err && typeof err === 'object' && 'errorFields' in err) {
        return
      }
      message.error('部署合约失败')
    } finally {
      setDeploying(false)
    }
  }

  const handleSendTx = async () => {
    try {
      const values = await sendForm.validateFields()
      setSending(true)
      
      if (txType === 'erc20') {
        // ERC-20 代币转账
        const unit = values.unit || 'token'
        const tokenAmount = parseToWei(String(values.value), unit).toString()
        
        const res = await tokenTransfer({
          to: values.to,
          amount: tokenAmount,
        })
        
        message.success(`代币转账交易已发送！哈希: ${res.data.txHash.slice(0, 10)}...`)
      } else {
        // ETH 转账
        const unit = values.unit || 'eth'
        const weiValue = parseToWei(String(values.value), unit).toString()
        
        // 转换 gasPrice 为 wei（如果用户提供了）
        let gasPriceWei = ''
        if (values.gasPrice) {
          gasPriceWei = parseToWei(String(values.gasPrice), 'gwei').toString()
        }
        
        // 获取 gasLimit（默认为 21000）
        const gasLimit = values.gasLimit || '21000'
        
        await sendTransaction({
          to: values.to,
          value: weiValue,
          gas: gasLimit,
          gasPrice: gasPriceWei,
        })
        message.success('ETH 转账交易已发送！')
      }
      
      setSendOpen(false)
      sendForm.resetFields()
      setTxType('eth')
      fetchTxHistory()
    } catch (err: unknown) {
      if (err && typeof err === 'object' && 'errorFields' in err) {
        return
      }
      message.error('发送交易失败')
    } finally {
      setSending(false)
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
        <Col xs={24} md={8}>
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

        <Col xs={24} md={8}>
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

        <Col xs={24} md={8}>
          <Card
            hoverable
            style={{ textAlign: 'center' }}
            onClick={() => setSendOpen(true)}
          >
            <SendOutlined style={{ fontSize: 40, color: '#faad14', marginBottom: 16 }} />
            <Title level={4}>发送 ETH</Title>
            <Paragraph type="secondary">
              快捷发送 ETH 转账交易
            </Paragraph>
            <Button type="primary" icon={<SendOutlined />} ghost>
              发送交易
            </Button>
          </Card>
        </Col>
      </Row>

      {/* 最近交易 */}
      <Card
        title={<><SwapOutlined /> 最近交易</>}
        style={{ maxWidth: 900, margin: '24px auto 0', overflowX: 'auto' }}
        loading={txLoading}
        extra={
          <Button 
            size="small" 
            icon={<SwapOutlined />}
            onClick={fetchTxHistory}
            loading={txLoading}
          >
            刷新
          </Button>
        }
      >
        {!txLoading && txHistory.length === 0 ? (
          <Empty description="暂无交易记录" />
        ) : (
          <Table<TxHistoryEntry>
            dataSource={txHistory}
            rowKey="id"
            size="small"
            pagination={false}
            scroll={{ x: 'max-content' }}
            columns={[
              {
                title: '交易哈希',
                dataIndex: 'txHash',
                key: 'txHash',
                ellipsis: true,
                width: 200,
                render: (hash: string) => (
                  <Tooltip title={hash}>
                    <Text copyable={{ text: hash, icon: [<CopyOutlined key="copy" style={{ marginLeft: 4, color: '#999' }} />, <CopyOutlined key="copied" style={{ marginLeft: 4, color: '#1677ff' }} />] }} style={{ fontSize: 12 }}>
                      {hash.slice(0, 8)}...{hash.slice(-6)}
                    </Text>
                  </Tooltip>
                ),
              },
              {
                title: '类型',
                dataIndex: 'txType',
                key: 'txType',
                width: 100,
                render: (txType: string) => (
                  <Tag color={txType === 'eth_transfer' ? 'blue' : 'green'}>
                    {txType === 'eth_transfer' ? 'ETH 转账' : 'ERC-20 转账'}
                  </Tag>
                ),
              },
              {
                title: '发送方',
                dataIndex: 'fromAddr',
                key: 'fromAddr',
                ellipsis: true,
                width: 150,
                render: (addr: string) => (
                  <Text style={{ fontSize: 12 }}>{addr.slice(0, 6)}...{addr.slice(-4)}</Text>
                ),
              },
              {
                title: '接收方',
                dataIndex: 'toAddr',
                key: 'toAddr',
                ellipsis: true,
                width: 150,
                render: (addr: string) => (
                  <Text style={{ fontSize: 12 }}>{addr.slice(0, 6)}...{addr.slice(-4)}</Text>
                ),
              },
              {
                title: '金额 (wei)',
                dataIndex: 'value',
                key: 'value',
                width: 120,
                render: (val: string) => (
                  <Text style={{ fontSize: 12 }}>{val}</Text>
                ),
              },
              {
                title: '状态',
                dataIndex: 'status',
                key: 'status',
                width: 80,
                render: (status: number, record: TxHistoryEntry) => {
                  if (status === 0) return <Tag color="processing">待确认</Tag>
                  if (status === 1) return <Tag color="success">成功</Tag>
                  return (
                    <Tooltip title={record.failureReason || '未知原因'}>
                      <Tag color="error">失败</Tag>
                    </Tooltip>
                  )
                },
              },
              {
                title: '失败原因',
                dataIndex: 'failureReason',
                key: 'failureReason',
                width: 150,
                ellipsis: true,
                render: (reason: string) => (
                  <Tooltip title={reason}>
                    <Text style={{ fontSize: 12, color: '#ff4d4f' }}>{reason || '-'}</Text>
                  </Tooltip>
                ),
              },
            ]}
          />
        )}
      </Card>

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

      <Modal
        title={txType === 'erc20' ? '发送 ERC-20 代币' : '发送 ETH'}
        open={sendOpen}
        onCancel={() => { setSendOpen(false); sendForm.resetFields(); setTxType('eth') }}
        onOk={handleSendTx}
        confirmLoading={sending}
        okText="发送"
        cancelText="取消"
        destroyOnClose
      >
        <Form form={sendForm} layout="vertical" style={{ marginTop: 16 }} onValuesChange={handleValuesChange}>
          <Form.Item label="交易类型">
            <Select
              value={txType}
              onChange={(value: 'eth' | 'erc20') => {
                setTxType(value)
                if (value === 'erc20') {
                  sendForm.setFieldValue('to', currentAddress)
                }
              }}
              style={{ width: 200 }}
            >
              <Option value="eth">ETH 转账</Option>
              <Option value="erc20">ERC-20 代币转账</Option>
            </Select>
          </Form.Item>
          
          {txType === 'erc20' && (
            <Form.Item label="合约地址">
              <Input
                value={currentAddress}
                disabled
                placeholder="当前监听合约地址"
              />
            </Form.Item>
          )}
          
          <Form.Item
            label="接收地址"
            name="to"
            rules={[{ required: true, message: '请输入接收地址' }]}
          >
            <Input placeholder="0x..." />
          </Form.Item>
          <Form.Item
            label="金额"
            required
            style={{ marginBottom: 0 }}
          >
            <Input.Group compact style={{ display: 'flex' }}>
              <Form.Item
                name="value"
                noStyle
                rules={[{ required: true, message: '请输入金额' }]}
              >
                <InputNumber
                  style={{ flex: 1, minWidth: 120 }}
                  placeholder="例如: 1.5"
                  min={0}
                  step={0.0001}
                  stringMode
                  precision={18}
                />
              </Form.Item>
              <Form.Item
                name="unit"
                noStyle
                initialValue="eth"
              >
                <Select
                  style={{ width: 100 }}
                  defaultValue="eth"
                >
                  <Option value="eth">ETH</Option>
                  <Option value="gwei">Gwei</Option>
                  <Option value="wei">Wei</Option>
                </Select>
              </Form.Item>
            </Input.Group>
          </Form.Item>
          <Form.Item name="convertedWei" style={{ marginBottom: 16 }}>
            <div style={{ color: 'rgba(0,0,0,0.45)', fontSize: 12 }}>
              实际发送 (wei): {displayWei}
            </div>
          </Form.Item>
          
          <Divider style={{ margin: '16px 0' }} />
          
          <Form.Item label="Gas 费用设置">
            <div style={{ fontSize: 12, color: 'rgba(0,0,0,0.45)', marginBottom: 12 }}>
              {gasFeeLoading ? (
                <span>获取 Gas 费用建议中...</span>
              ) : gasFeeSuggestion ? (
                <div>
                  <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 4 }}>
                    <span>建议 Gas Price:</span>
                    <span>{formatGwei(gasFeeSuggestion.gasPrice)} Gwei</span>
                  </div>
                  {gasFeeSuggestion.supportsEIP1559 && (
                    <>
                      <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 4 }}>
                        <span>Base Fee:</span>
                        <span>{formatGwei(gasFeeSuggestion.baseFee)} Gwei</span>
                      </div>
                      <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 4 }}>
                        <span>Priority Fee:</span>
                        <span>{formatGwei(gasFeeSuggestion.gasTipCap)} Gwei</span>
                      </div>
                    </>
                  )}
                  <div style={{ display: 'flex', justifyContent: 'space-between', marginTop: 8, paddingTop: 8, borderTop: '1px dashed #d9d9d9' }}>
                    <span><strong>预估 Gas 费:</strong></span>
                    <span><strong>{gasFeeSuggestion.estimatedCostETH} ETH</strong></span>
                  </div>
                </div>
              ) : (
                <span>无法获取 Gas 费用建议</span>
              )}
            </div>
            
            <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
              <Form.Item
                name="gasPrice"
                noStyle
              >
                <InputNumber
                  style={{ flex: 1 }}
                  placeholder="Gas Price"
                  min={0}
                  step={1}
                  stringMode
                  precision={9}
                  disabled={gasFeeLoading}
                  addonAfter="Gwei"
                />
              </Form.Item>
            </div>
            
            <div style={{ fontSize: 12, color: '#1890ff', marginTop: 8, cursor: 'pointer' }} onClick={() => {
              if (gasFeeSuggestion) {
                sendForm.setFieldValue('gasPrice', formatGwei(gasFeeSuggestion.gasPrice))
                // 触发计算
                handleValuesChange({}, { ...sendForm.getFieldsValue(), gasPrice: formatGwei(gasFeeSuggestion.gasPrice) })
              }
            }}>
              <InfoCircleOutlined style={{ marginRight: 4 }} />
              点击使用推荐 Gas Price
            </div>
          </Form.Item>
          
          <Form.Item 
            label={estimatingGas ? 'Gas Limit (估算中...)' : 'Gas Limit'}
            name="gasLimit"
            initialValue="21000"
          >
            <InputNumber
              style={{ width: '100%' }}
              placeholder="例如: 21000"
              min={21000}
              max={1000000}
              stringMode
              disabled={estimatingGas}
            />
            {estimatingGas && <Text type="secondary" style={{ fontSize: 12 }}>正在根据接收地址自动估算...</Text>}
          </Form.Item>
          
          <Divider style={{ margin: '16px 0' }} />
          
          {/* 最终交易额显示 */}
          <Form.Item label={<><strong>交易汇总</strong></>}>
            {(() => {
              const totalCost = calculateTotalCost()
              if (!totalCost) {
                return <Text type="secondary">请输入金额</Text>
              }
              return (
                <div style={{ fontSize: 13 }}>
                  <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 4 }}>
                    <span>发送金额:</span>
                    <span>{formatETH(totalCost.valueWei)} ETH</span>
                  </div>
                  <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 4 }}>
                    <span>Gas 费用:</span>
                    <span>{totalCost.gasCost === '0' ? '自动计算' : `${formatETH(totalCost.gasCost)} ETH`}</span>
                  </div>
                  <div style={{ display: 'flex', justifyContent: 'space-between', marginTop: 8, paddingTop: 8, borderTop: '1px dashed #d9d9d9' }}>
                    <span><strong>总计:</strong></span>
                    <span style={{ color: '#1890ff', fontWeight: 'bold' }}>{totalCost.totalETH} ETH</span>
                  </div>
                </div>
              )
            })()}
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}

function formatGwei(wei: string): string {
  try {
    const weiBig = BigInt(wei)
    const gweiDivisor = BigInt(10) ** BigInt(9)
    const gwei = Number(weiBig / gweiDivisor) + (Number(weiBig % gweiDivisor) / 1e9)
    return gwei.toFixed(9).replace(/\.?0+$/, '')
  } catch {
    return wei
  }
}

export default Home
