import React from 'react'
import { useSaasAuth } from '../hooks/useSaasAuth'
import type { AuthProvider } from '../types'

export interface SocialLoginButtonsProps {
  /** Which providers to show */
  providers?: AuthProvider[]
  /** Custom redirect URI after OAuth */
  redirectUri?: string
  /** Called on successful login */
  onSuccess?: () => void
  /** Called on error */
  onError?: (error: Error) => void
  /** Custom class for the container */
  className?: string
  /** Custom class for buttons */
  buttonClassName?: string
  /** Custom button labels */
  labels?: Partial<Record<AuthProvider, string>>
  /** Whether buttons are disabled */
  disabled?: boolean
}

const defaultLabels: Record<AuthProvider, string> = {
  google: 'Continue with Google',
  github: 'Continue with GitHub',
}

/**
 * Pre-built social login buttons
 *
 * @example
 * ```tsx
 * <SocialLoginButtons
 *   providers={['google', 'github']}
 *   onSuccess={() => router.push('/dashboard')}
 *   className="flex flex-col gap-2"
 * />
 * ```
 */
export function SocialLoginButtons({
  providers = ['google', 'github'],
  redirectUri,
  onSuccess,
  onError,
  className = '',
  buttonClassName = '',
  labels = {},
  disabled = false,
}: SocialLoginButtonsProps) {
  const { loginWithGoogle, loginWithGithub, isLoading } = useSaasAuth()

  const handleLogin = async (provider: AuthProvider) => {
    try {
      if (provider === 'google') {
        await loginWithGoogle(redirectUri)
      } else if (provider === 'github') {
        await loginWithGithub(redirectUri)
      }
      onSuccess?.()
    } catch (err) {
      onError?.(err as Error)
    }
  }

  return (
    <div className={className}>
      {providers.map((provider) => (
        <button
          key={provider}
          type="button"
          onClick={() => handleLogin(provider)}
          disabled={disabled || isLoading}
          className={buttonClassName}
          aria-label={labels[provider] || defaultLabels[provider]}
        >
          {provider === 'google' && <GoogleIcon />}
          {provider === 'github' && <GitHubIcon />}
          <span>{labels[provider] || defaultLabels[provider]}</span>
        </button>
      ))}
    </div>
  )
}

// Simple SVG icons
function GoogleIcon() {
  return (
    <svg width="18" height="18" viewBox="0 0 18 18" fill="none" xmlns="http://www.w3.org/2000/svg">
      <path
        d="M17.64 9.2c0-.637-.057-1.251-.164-1.84H9v3.481h4.844a4.14 4.14 0 01-1.796 2.716v2.259h2.908c1.702-1.567 2.684-3.875 2.684-6.616z"
        fill="#4285F4"
      />
      <path
        d="M9 18c2.43 0 4.467-.806 5.956-2.18l-2.908-2.259c-.806.54-1.837.86-3.048.86-2.344 0-4.328-1.584-5.036-3.711H.957v2.332A8.997 8.997 0 009 18z"
        fill="#34A853"
      />
      <path
        d="M3.964 10.71A5.41 5.41 0 013.682 9c0-.593.102-1.17.282-1.71V4.958H.957A8.996 8.996 0 000 9c0 1.452.348 2.827.957 4.042l3.007-2.332z"
        fill="#FBBC05"
      />
      <path
        d="M9 3.58c1.321 0 2.508.454 3.44 1.345l2.582-2.58C13.463.891 11.426 0 9 0A8.997 8.997 0 00.957 4.958L3.964 7.29C4.672 5.163 6.656 3.58 9 3.58z"
        fill="#EA4335"
      />
    </svg>
  )
}

function GitHubIcon() {
  return (
    <svg width="18" height="18" viewBox="0 0 18 18" fill="currentColor" xmlns="http://www.w3.org/2000/svg">
      <path
        fillRule="evenodd"
        clipRule="evenodd"
        d="M9 0C4.037 0 0 4.037 0 9c0 3.975 2.578 7.35 6.154 8.541.45.082.616-.195.616-.434 0-.214-.008-.78-.012-1.531-2.504.544-3.032-1.207-3.032-1.207-.41-1.04-1-1.316-1-1.316-.817-.558.062-.547.062-.547.903.064 1.379.928 1.379.928.803 1.376 2.107.978 2.62.748.082-.582.314-.978.572-1.203-2-.227-4.104-1-4.104-4.452 0-.983.35-1.787.928-2.417-.094-.228-.402-1.144.088-2.384 0 0 .756-.242 2.476.923a8.62 8.62 0 012.253-.303c.765.004 1.535.103 2.253.303 1.718-1.165 2.473-.923 2.473-.923.491 1.24.183 2.156.09 2.384.578.63.926 1.434.926 2.417 0 3.461-2.107 4.223-4.113 4.445.323.278.611.828.611 1.668 0 1.204-.01 2.175-.01 2.471 0 .241.163.521.62.433C15.425 16.347 18 12.974 18 9c0-4.963-4.037-9-9-9z"
      />
    </svg>
  )
}
