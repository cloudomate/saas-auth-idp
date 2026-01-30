import React from 'react'
import ReactDOM from 'react-dom/client'
import { BrowserRouter } from 'react-router-dom'
import { SaasAuthProvider } from '@saas-starter/react'
import App from './App'
import './styles/index.css'

const API_URL = import.meta.env.VITE_API_URL || 'http://localhost:4455'

ReactDOM.createRoot(document.getElementById('root')!).render(
  <React.StrictMode>
    <BrowserRouter>
      <SaasAuthProvider
        apiUrl={API_URL}
        storagePrefix="example-app"
        onAuthStateChange={(user) => {
          console.log('Auth state changed:', user?.email || 'logged out')
        }}
      >
        <App />
      </SaasAuthProvider>
    </BrowserRouter>
  </React.StrictMode>,
)
