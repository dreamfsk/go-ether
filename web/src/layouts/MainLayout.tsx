import { Outlet, useNavigate, useLocation } from 'react-router-dom'
import { Layout, Menu, theme } from 'antd'
import {
  HomeOutlined,
  UnorderedListOutlined,
  ApiOutlined,
} from '@ant-design/icons'

const { Header, Content, Footer } = Layout

const MainLayout = () => {
  const navigate = useNavigate()
  const location = useLocation()
  const {
    token: { colorBgContainer, borderRadiusLG },
  } = theme.useToken()

  const getSelectedKey = () => {
    if (location.pathname === '/manage/index') return 'home'
    if (location.pathname.startsWith('/manage/contracts')) return 'contracts'
    return 'home'
  }

  const menuItems = [
    {
      key: 'home',
      icon: <HomeOutlined />,
      label: '首页',
      onClick: () => navigate('/manage/index'),
    },
    {
      key: 'contracts',
      icon: <UnorderedListOutlined />,
      label: '合约列表',
      onClick: () => navigate('/manage/contracts'),
    },
  ]

  return (
    <Layout style={{ minHeight: '100vh' }}>
      <Header
        style={{
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'space-between',
          padding: '0 24px',
          background: '#001529',
        }}
      >
        <div style={{ display: 'flex', alignItems: 'center', gap: 16 }}>
          <ApiOutlined style={{ fontSize: 24, color: '#fff' }} />
          <span style={{ color: '#fff', fontSize: 18, fontWeight: 600 }}>
            Go-Ether 管理面板
          </span>
        </div>
        <Menu
          theme="dark"
          mode="horizontal"
          selectedKeys={[getSelectedKey()]}
          items={menuItems}
          style={{ flex: 1, justifyContent: 'flex-end', minWidth: 0 }}
        />
      </Header>
      <Content style={{ padding: '24px' }}>
        <div
          style={{
            background: colorBgContainer,
            borderRadius: borderRadiusLG,
            padding: 24,
            minHeight: 'calc(100vh - 160px)',
          }}
        >
          <Outlet />
        </div>
      </Content>
      <Footer style={{ textAlign: 'center' }}>
        Go-Ether ©{new Date().getFullYear()} - 迷你区块浏览器与 ERC-20 监听服务
      </Footer>
    </Layout>
  )
}

export default MainLayout
