import { Routes, Route, Navigate } from 'react-router-dom'
import { ConfigProvider } from 'antd'
import zhCN from 'antd/locale/zh_CN'
import MainLayout from './layouts/MainLayout'
import Home from './pages/Home'
import ContractList from './pages/ContractList'
import ContractDetail from './pages/ContractDetail'

function App() {
  return (
    <ConfigProvider locale={zhCN}>
      <Routes>
        <Route path="/manage" element={<MainLayout />}>
          <Route index element={<Navigate to="/manage/index" replace />} />
          <Route path="index" element={<Home />} />
          <Route path="contracts" element={<ContractList />} />
          <Route path="contracts/:address" element={<ContractDetail />} />
        </Route>
        <Route path="*" element={<Navigate to="/manage/index" replace />} />
      </Routes>
    </ConfigProvider>
  )
}

export default App
