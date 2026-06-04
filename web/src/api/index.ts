import axios from 'axios'
import type {
  ContractListResponse,
  DeployRequest,
  DeployResponse,
  SwitchContractRequest,
  SwitchContractResponse,
  TokenInfo,
  TokenBalanceResponse,
  EventsResponse,
  TxHistoryEntry,
  SendTxRequest,
  SendTxResponse,
  GasFeeSuggestion,
} from '../types'

export interface ConfigResponse {
  maxTokenAmount: string
}

const api = axios.create({
  baseURL: '/api',
  timeout: 15000,
  headers: {
    'Content-Type': 'application/json',
  },
})

// ============ 合约相关 API ============

/** 获取合约列表 */
export const getContractList = () =>
  api.get<ContractListResponse>('/contract/list')

/** 获取当前活跃合约 */
export const getCurrentContract = () =>
  api.get<{ address: string; name: string; symbol: string; network: string; isActive: boolean }>('/contract/current')

/** 切换合约 */
export const switchContract = (data: SwitchContractRequest) =>
  api.post<SwitchContractResponse>('/contract/switch', data)

// ============ 代币相关 API ============

/** 获取代币信息（可指定合约地址，不传则查当前活跃合约） */
export const getTokenInfo = (contractAddr?: string) =>
  api.get<TokenInfo>('/token/info', { params: contractAddr ? { address: contractAddr } : undefined })

/** 获取代币余额 */
export const getTokenBalance = (holder: string) =>
  api.get<TokenBalanceResponse>('/token/balance', { params: { holder } })

/** 部署合约 */
export const deployContract = (data: DeployRequest) =>
  api.post<DeployResponse>('/token/deploy', data)

/** 代币转账 */
export const tokenTransfer = (data: { to: string; amount: string }) =>
  api.post<{ txHash: string; status: string }>('/token/transfer', data)

// ============ 事件相关 API ============

/** 获取事件列表 */
export const getEvents = (params?: { address?: string; limit?: number; offset?: number; contractOnly?: boolean }) =>
  api.get<EventsResponse>('/events', { params })

/** 获取配置信息 */
export const getConfig = () =>
  api.get<ConfigResponse>('/config')

// ============ 交易相关 API ============

/** 获取交易历史 */
export const getTxHistory = (params?: { limit?: number; offset?: number }) =>
  api.get<TxHistoryEntry[]>('/tx/history', { params })

/** 发送 ETH 交易 */
export const sendTransaction = (data: SendTxRequest) =>
  api.post<SendTxResponse>('/tx/send', data)

/** 获取 Gas 费用建议 */
export const getGasFeeSuggestion = () =>
  api.get<GasFeeSuggestion>('/tx/gas-fee')

/** 估算 Gas */
export const estimateGas = (to: string, value: string) =>
  api.post<{ gas: number }>('/tx/estimate-gas', { to, value })

export default api
