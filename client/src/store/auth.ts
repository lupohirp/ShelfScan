import { create } from 'zustand'
import type { User } from '../types'
import { getApiUrl } from '../lib/api'

interface AuthState {
  user: User | null
  isAuthenticated: boolean
  login: (email: string, password: string) => Promise<void>
  logout: () => void
}

export const useAuth = create<AuthState>((set) => ({
  user: null,
  isAuthenticated: false,
  login: async (email: string, password: string) => {
    const apiBase = getApiUrl()
    const res = await fetch(`${apiBase}/login`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({ email, password }),
    })

    if (!res.ok) {
      const errorText = await res.text()
      throw new Error(errorText || 'Credenziali non valide')
    }

    const data = await res.json()
    
    // Split full name into firstName and lastName
    const names = (data.agente || 'Agente').split(' ')
    const firstName = names[0]
    const lastName = names.slice(1).join(' ')

    set({
      isAuthenticated: true,
      user: {
        id: data.id,
        email: data.email,
        firstName: firstName,
        lastName: lastName,
        role: 'rep',
      },
    })
  },
  logout: () => set({ user: null, isAuthenticated: false }),
}))
