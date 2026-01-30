import React from 'react'
import ReactDOM from 'react-dom/client'
import { BrowserRouter } from 'react-router-dom'
import { SaasAuthProvider } from '@saas-starter/react'
import { AuthProvider } from './contexts/AuthContext'
import App from './App'
import './styles/index.css'

const API_URL = import.meta.env.VITE_API_URL || 'http://localhost:8001'

ReactDOM.createRoot(document.getElementById('root')!).render(
  <React.StrictMode>
    <BrowserRouter>
      <AuthProvider>
        <SaasAuthProvider
          apiUrl={API_URL}
          storagePrefix="example-app"
          onAuthStateChange={(user) => {
            console.log('Auth state changed:', user?.email || 'logged out')
          }}
        >
          <App />
        </SaasAuthProvider>
      </AuthProvider>
    </BrowserRouter>
  </React.StrictMode>,
)
