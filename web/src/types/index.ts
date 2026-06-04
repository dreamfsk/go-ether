/** 合约信息 */
export interface ContractEntry {
  id: number
  address: string
  name: string
  symbol: string
  network: string
  deployer: string
  txHash: string
  isActive: boolean
  createdAt: string
}

/** 合约列表响应 */
export interface ContractListResponse {
  contracts: ContractEntry[]
  currentAddress: string
}

/** 部署合约请求 */
export interface DeployRequest {
  name: string
  symbol: string
  initialSupply: string
  recipient: string
}

/** 部署合约响应 */
export interface DeployResponse {
  address: string
  txHash: string
}

/** 切换合约请求 */
export interface SwitchContractRequest {
  address: string
}

/** 切换合约响应 */
export interface SwitchContractResponse {
  success: boolean
  address: string
  message: string
}

/** 代币信息 */
export interface TokenInfo {
  name: string
  symbol: string
  decimals: number
  totalSupply: string
  contractAddr: string
}

/** 代币余额响应 */
export interface TokenBalanceResponse {
  holder: string
  balance: string
  contractAddr: string
}

/** 事件条目 */
export interface EventEntry {
  id: number
  txHash: string
  fromAddr: string
  toAddr: string
  value: string
  status: number
  blockNumber: number
  network: string
  txType: string
  createdAt: string
}

/** 分页事件响应 */
export interface EventsResponse {
  events: EventEntry[]
  total: number
  limit: number
  offset: number
}

// ============ 交易相关类型 ============

/** 交易历史条目 */
export interface TxHistoryEntry {
  id: number
  txHash: string
  fromAddr: string
  toAddr: string
  value: string
  gasLimit: number
  gasPrice: string
  nonce: number
  data: string
  status: number // 0=pending, 1=success, 2=failed
  blockNumber: number
  network: string
  txType: string // "eth_transfer" | "erc20_transfer"
  contractAddr?: string
  failureReason?: string
  createdAt: string
}

/** 发送交易请求 */
export interface SendTxRequest {
  to: string
  value: string
  gas?: number
  gasPrice?: string
  gasTip?: string
  data?: string
}

/** 发送交易响应 */
export interface SendTxResponse {
  txHash: string
  from: string
  to: string
  value: string
  nonce: number
  gasLimit: number
  gasPrice: string
  status: string
}

/** Gas 费用建议 */
export interface GasFeeSuggestion {
  gasPrice: string       // Legacy 交易的 gas price (wei)
  gasTipCap: string      // EIP-1559 的优先费用 (wei)
  gasFeeCap: string      // EIP-1559 的最大费用 (wei)
  baseFee: string        // 当前区块基础费用 (wei)
  estimatedCost: string  // 预估交易费用 (wei)
  estimatedCostETH: string // 预估交易费用 (ETH)
  supportsEIP1559: boolean // 是否支持 EIP-1559
}
