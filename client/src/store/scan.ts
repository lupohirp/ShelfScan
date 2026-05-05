import { create } from 'zustand'
import type { Store, Product, CheckSession } from '../types'

interface ScanState {
  selectedStore: Store | null
  currentSession: CheckSession | null
  setStore: (store: Store) => void
  clearSession: () => void
  setSession: (session: CheckSession) => void
  toggleProduct: (productId: string) => void
}

export const useScan = create<ScanState>((set) => ({
  selectedStore: null,
  currentSession: null,
  setStore: (store) => set({ selectedStore: store }),
  clearSession: () => set({ currentSession: null, selectedStore: null }),
  setSession: (session) => set({ currentSession: session }),
  toggleProduct: (productId) =>
    set((state) => {
      if (!state.currentSession) return state
      const found = state.currentSession.foundProducts
      const missing = state.currentSession.missingProducts
      const isFound = found.some((p) => p.id === productId)

      let newFound: Product[]
      let newMissing: Product[]

      if (isFound) {
        const product = found.find((p) => p.id === productId)!
        newFound = found.filter((p) => p.id !== productId)
        newMissing = [...missing, product]
      } else {
        const product = missing.find((p) => p.id === productId)!
        newMissing = missing.filter((p) => p.id !== productId)
        newFound = [...found, product]
      }

      const total = newFound.length + newMissing.length
      return {
        currentSession: {
          ...state.currentSession,
          foundProducts: newFound,
          missingProducts: newMissing,
          coverage: total > 0 ? Math.round((newFound.length / total) * 100) : 0,
        },
      }
    }),
}))
