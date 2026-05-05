import type { Store, Product, CheckSession } from '../types'

export const mockStores: Store[] = [
  { id: '1', name: 'Gioielleria Bianchi', city: 'Milano', address: 'Via Montenapoleone 12' },
  { id: '2', name: 'Oro & Argento', city: 'Roma', address: 'Via Condotti 45' },
  { id: '3', name: 'Preziosi da Maria', city: 'Firenze', address: 'Ponte Vecchio 8' },
  { id: '4', name: 'Luxury Corner', city: 'Torino', address: 'Via Roma 102' },
  { id: '5', name: 'Gioielli Del Corso', city: 'Napoli', address: 'Via Toledo 78' },
  { id: '6', name: 'Bijoux Elegante', city: 'Bologna', address: 'Via Indipendenza 33' },
]

export const mockProducts: Product[] = [
  { id: '1', sku: 'LJ-RING-001', name: 'Anello Precious Heart', category: 'Anelli', imageUrl: '', status: 'active' },
  { id: '2', sku: 'LJ-RING-002', name: 'Anello Infinity Gold', category: 'Anelli', imageUrl: '', status: 'active' },
  { id: '3', sku: 'LJ-NECK-001', name: 'Collana Diamond Chain', category: 'Collane', imageUrl: '', status: 'active' },
  { id: '4', sku: 'LJ-NECK-002', name: 'Collana Pearl Drop', category: 'Collane', imageUrl: '', status: 'active' },
  { id: '5', sku: 'LJ-NECK-003', name: 'Collana Minimal Bar', category: 'Collane', imageUrl: '', status: 'active' },
  { id: '6', sku: 'LJ-BRAC-001', name: 'Bracciale Tennis Classic', category: 'Bracciali', imageUrl: '', status: 'active' },
  { id: '7', sku: 'LJ-BRAC-002', name: 'Bracciale Chain Link', category: 'Bracciali', imageUrl: '', status: 'active' },
  { id: '8', sku: 'LJ-EARR-001', name: 'Orecchini Stud Diamond', category: 'Orecchini', imageUrl: '', status: 'active' },
  { id: '9', sku: 'LJ-EARR-002', name: 'Orecchini Hoop Gold', category: 'Orecchini', imageUrl: '', status: 'active' },
  { id: '10', sku: 'LJ-EARR-003', name: 'Orecchini Drop Pearl', category: 'Orecchini', imageUrl: '', status: 'active' },
  { id: '11', sku: 'LJ-RING-003', name: 'Anello Solitaire', category: 'Anelli', imageUrl: '', status: 'active' },
  { id: '12', sku: 'LJ-BRAC-003', name: 'Bracciale Charm Collection', category: 'Bracciali', imageUrl: '', status: 'active' },
]

export const mockHistory: CheckSession[] = [
  {
    id: '1',
    store: mockStores[0],
    status: 'finalized',
    scans: [],
    foundProducts: mockProducts.slice(0, 8),
    missingProducts: mockProducts.slice(8),
    coverage: 67,
    createdAt: '2026-03-15T10:30:00Z',
    finalizedAt: '2026-03-15T10:45:00Z',
  },
  {
    id: '2',
    store: mockStores[1],
    status: 'finalized',
    scans: [],
    foundProducts: mockProducts.slice(0, 5),
    missingProducts: mockProducts.slice(5),
    coverage: 42,
    createdAt: '2026-03-14T14:00:00Z',
    finalizedAt: '2026-03-14T14:20:00Z',
  },
  {
    id: '3',
    store: mockStores[2],
    status: 'finalized',
    scans: [],
    foundProducts: mockProducts.slice(0, 10),
    missingProducts: mockProducts.slice(10),
    coverage: 83,
    createdAt: '2026-03-12T09:15:00Z',
    finalizedAt: '2026-03-12T09:30:00Z',
  },
  {
    id: '4',
    store: mockStores[3],
    status: 'draft',
    scans: [],
    foundProducts: mockProducts.slice(0, 3),
    missingProducts: mockProducts.slice(3),
    coverage: 25,
    createdAt: '2026-03-11T16:00:00Z',
  },
]
