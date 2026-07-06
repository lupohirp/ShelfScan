import { create } from 'zustand'

interface PwaState {
  deferredPrompt: any | null
  setDeferredPrompt: (prompt: any | null) => void
  showInstallBanner: boolean
  setShowInstallBanner: (show: boolean) => void
}

export const usePwa = create<PwaState>((set) => ({
  deferredPrompt: null,
  setDeferredPrompt: (prompt) => set({ deferredPrompt: prompt }),
  showInstallBanner: false,
  setShowInstallBanner: (show) => set({ showInstallBanner: show }),
}))
