import { useState } from 'react'
import './App.css'
import { apiFetch } from './api'

function App() {
  // Order state
  const [orderUserId, setOrderUserId] = useState('')
  const [orderDesc, setOrderDesc] = useState('')
  const [orderAmount, setOrderAmount] = useState('')
  const [orderResult, setOrderResult] = useState('')

  const [listUserId, setListUserId] = useState('')
  const [listResult, setListResult] = useState('')

  const [statusOrderId, setStatusOrderId] = useState('')
  const [statusResult, setStatusResult] = useState('')

  // Payment state
  const [payUserId, setPayUserId] = useState('')
  const [payResult, setPayResult] = useState('')

  const [depositUserId, setDepositUserId] = useState('')
  const [depositAmount, setDepositAmount] = useState('')
  const [depositResult, setDepositResult] = useState('')

  const [balanceUserId, setBalanceUserId] = useState('')
  const [balanceResult, setBalanceResult] = useState('')

  // Order handlers
  async function createOrder() {
    setOrderResult('...')
    const res = await apiFetch('/api/v1/orders', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ user_id: orderUserId, description: orderDesc, amount: Number(orderAmount) })
    })
    setOrderResult(await res.text())
  }
  async function listOrders() {
    setListResult('...')
    const res = await apiFetch(`/api/v1/orders?user_id=${encodeURIComponent(listUserId)}`)
    setListResult(await res.text())
  }
  async function getOrderStatus() {
    setStatusResult('...')
    const res = await apiFetch(`/api/v1/orders/${encodeURIComponent(statusOrderId)}/status`)
    setStatusResult(await res.text())
  }

  // Payment handlers
  async function createAccount() {
    setPayResult('...')
    const res = await apiFetch('/api/v1/payment/accounts', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ user_id: payUserId })
    })
    setPayResult(await res.text())
  }
  async function deposit() {
    setDepositResult('...')
    const res = await apiFetch(`/api/v1/payment/accounts/${encodeURIComponent(depositUserId)}/deposit`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ user_id: depositUserId, amount: Number(depositAmount) })
    })
    setDepositResult(await res.text())
  }
  async function getBalance() {
    setBalanceResult('...')
    const res = await apiFetch(`/api/v1/payment/accounts/${encodeURIComponent(balanceUserId)}/balance`)
    setBalanceResult(await res.text())
  }

  return (
    <div style={{ maxWidth: 600, margin: '2em auto', fontFamily: 'sans-serif' }}>
      <h1>Goshop Frontend</h1>
      <h2>Order Service</h2>
      <div style={{ border: '1px solid #ccc', padding: 16, marginBottom: 24 }}>
        <h3>Создать заказ</h3>
        <input placeholder="user_id" value={orderUserId} onChange={e => setOrderUserId(e.target.value)} />
        <input placeholder="description" value={orderDesc} onChange={e => setOrderDesc(e.target.value)} />
        <input placeholder="amount" type="number" value={orderAmount} onChange={e => setOrderAmount(e.target.value)} />
        <button onClick={createOrder}>Создать</button>
        <pre>{orderResult}</pre>
      </div>
      <div style={{ border: '1px solid #ccc', padding: 16, marginBottom: 24 }}>
        <h3>Список заказов пользователя</h3>
        <input placeholder="user_id" value={listUserId} onChange={e => setListUserId(e.target.value)} />
        <button onClick={listOrders}>Показать</button>
        <pre>{listResult}</pre>
      </div>
      <div style={{ border: '1px solid #ccc', padding: 16, marginBottom: 24 }}>
        <h3>Статус заказа</h3>
        <input placeholder="order_id" value={statusOrderId} onChange={e => setStatusOrderId(e.target.value)} />
        <button onClick={getOrderStatus}>Показать</button>
        <pre>{statusResult}</pre>
      </div>
      <h2>Payment Service</h2>
      <div style={{ border: '1px solid #ccc', padding: 16, marginBottom: 24 }}>
        <h3>Создать счет</h3>
        <input placeholder="user_id" value={payUserId} onChange={e => setPayUserId(e.target.value)} />
        <button onClick={createAccount}>Создать</button>
        <pre>{payResult}</pre>
      </div>
      <div style={{ border: '1px solid #ccc', padding: 16, marginBottom: 24 }}>
        <h3>Пополнить счет</h3>
        <input placeholder="user_id" value={depositUserId} onChange={e => setDepositUserId(e.target.value)} />
        <input placeholder="amount" type="number" value={depositAmount} onChange={e => setDepositAmount(e.target.value)} />
        <button onClick={deposit}>Пополнить</button>
        <pre>{depositResult}</pre>
      </div>
      <div style={{ border: '1px solid #ccc', padding: 16, marginBottom: 24 }}>
        <h3>Получить баланс</h3>
        <input placeholder="user_id" value={balanceUserId} onChange={e => setBalanceUserId(e.target.value)} />
        <button onClick={getBalance}>Показать</button>
        <pre>{balanceResult}</pre>
      </div>
    </div>
  )
}

export default App
