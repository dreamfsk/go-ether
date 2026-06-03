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
