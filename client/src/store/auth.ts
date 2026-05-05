import { create } from 'zustand'
import type { User } from '../types'

interface AuthState {
  user: User | null
  isAuthenticated: boolean
  login: (email: string, password: string) => Promise<void>
  logout: () => void
}

export const useAuth = create<AuthState>((set) => ({
  user: null,
  isAuthenticated: false,
  login: async (_email: string, _password: string) => {
    // Mock login - replace with real API
    await new Promise(r => setTimeout(r, 800))
    set({
      isAuthenticated: true,
      user: {
        id: '1',
        email: _email,
        firstName: 'Marco',
        lastName: 'Rossi',
        role: 'rep',
      },
    })
  },
  logout: () => set({ user: null, isAuthenticated: false }),
}))
