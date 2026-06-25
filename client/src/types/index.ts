export type UserRole = 'admin' | 'rep' | 'retailer'

export interface User {
  id: string
  email: string
  firstName: string
  lastName: string
  role: UserRole
}

export interface Store {
  id: string
  name: string
  city: string
  address: string
}

export interface Product {
  id: string
  sku: string
  name: string
  category: string
  subcategory?: string
  imageUrl: string
  status: 'active' | 'discontinued'
  confidence?: number
  count?: number
}

export interface DetectedProduct {
  id: string
  product: Product
  confidence: number
  manualOverride: boolean
}

export interface Scan {
  id: string
  imageUrl: string
  detectedProducts: DetectedProduct[]
  createdAt: string
}

export interface AnalyzedImage {
  capturedImage: string
  detections: { desc: string, box?: number[], box_2d?: number[] }[]
}

export interface CheckSession {
  id: string
  store: Store
  status: 'draft' | 'finalized'
  scans: Scan[]
  foundProducts: Product[]
  missingProducts: Product[]
  coverage: number
  createdAt: string
  finalizedAt?: string
  analyzedImages?: AnalyzedImage[]
}
