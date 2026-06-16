import { useState, useEffect } from 'react'
import { 
  Upload, 
  Package, 
  CheckCircle2, 
  Trash2, 
  LayoutDashboard, 
  PlusCircle, 
  Loader2, 
  Pencil,
  BarChart3,
  Database,
  Search,
  Grid3X3,
  List,
  X,
  ChevronDown
} from 'lucide-react'

interface InventoryItem {
  id: string
  payload: {
    name: string
    sku?: string
    imageUrl?: string
    color?: string
    material?: string
  }
}

interface GroupedInventoryItem {
  sku: string
  name: string
  color?: string
  material?: string
  imageUrls: string[]
  ids: string[]
  category: string
}

const getApiUrl = () => {
  const envUrl = import.meta.env.VITE_API_URL;
  if (envUrl) {
    return envUrl;
  }
  const hostname = window.location.hostname;
  if (hostname === 'localhost' || hostname === '127.0.0.1' || hostname.startsWith('192.168.') || hostname.startsWith('10.')) {
    return `http://${hostname}:8080`;
  }
  if (import.meta.env.PROD) {
    const baseHost = hostname.replace(/^admin-/, '');
    return `https://api-${baseHost}`;
  }
  return `http://${hostname}:8080`;
};
const apiBase = getApiUrl();

const categoryLabels: Record<string, string> = {
  All: 'Tutti',
  Watch: 'Orologio',
  Ring: 'Anello',
  Necklace: 'Collana',
  Bracelet: 'Bracciale',
  Earring: 'Orecchino',
  Other: 'Altro',
  watch: 'Orologio',
  ring: 'Anello',
  necklace: 'Collana',
  bracelet: 'Bracciale',
  earring: 'Orecchino',
  other: 'Altro'
}

function App() {
  const [items, setItems] = useState<InventoryItem[]>([])
  const [view, setView] = useState<'list' | 'add' | 'edit'>('list')
  const [layoutMode, setLayoutMode] = useState<'grid' | 'table'>('grid')
  
  // File upload state
  const [files, setFiles] = useState<FileList | null>(null)
  const [name, setName] = useState('')
  const [sku, setSku] = useState('')
  const [category, setCategory] = useState<string>('watch')

  // Search & Filter State
  const [searchTerm, setSearchTerm] = useState('')
  const [selectedFilterCategory, setSelectedFilterCategory] = useState('All')

  // Category specific fields
  const [watchStrapColor, setWatchStrapColor] = useState('')
  const [watchStrapMaterial, setWatchStrapMaterial] = useState('')
  const [watchDialColor, setWatchDialColor] = useState('')
  const [watchDialShape, setWatchDialShape] = useState('')

  const [ringMetal, setRingMetal] = useState('')
  const [ringStone, setRingStone] = useState('')
  const [ringStoneColor, setRingStoneColor] = useState('')

  const [jewelryColor, setJewelryColor] = useState('')
  const [jewelryMaterial, setJewelryMaterial] = useState('')
  const [loading, setLoading] = useState(false)
  const [success, setSuccess] = useState(false)
  const [editIds, setEditIds] = useState<string[]>([])
  const [selectedImageMap, setSelectedImageMap] = useState<Record<string, string>>({})

  const fetchInventory = async () => {
    try {
      const response = await fetch(`${apiBase}/inventory`)
      if (response.ok) {
        const data = await response.json()
        setItems(data || [])
      }
    } catch (err) {
      console.error('Failed to fetch inventory:', err)
    }
  }

  useEffect(() => {
    fetchInventory()
  }, [])

  const resetForm = () => {
    setName('')
    setSku('')
    setJewelryColor('')
    setJewelryMaterial('')
    setWatchStrapColor('')
    setWatchStrapMaterial('')
    setWatchDialColor('')
    setWatchDialShape('')
    setRingMetal('')
    setRingStone('')
    setRingStoneColor('')
    setFiles(null)
    setEditIds([])
  }

  const handleStartEdit = (item: GroupedInventoryItem) => {
    setEditIds(item.ids)
    setName(item.name)
    setSku(item.sku)
    setJewelryColor(item.color || '')
    setJewelryMaterial(item.material || '')
    setView('edit')
  }

  const handleUpload = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!files || !name || !sku) return

    setLoading(true)
    const formData = new FormData()
    for (let i = 0; i < files.length; i++) {
      formData.append('images', files[i])
    }
    formData.append('name', name)
    formData.append('sku', sku)

    let finalColor = ''
    let finalMaterial = ''

    if (category === 'watch') {
      finalColor = jewelryColor || 'Gold'
      if (watchStrapColor) finalColor += `, strap: ${watchStrapColor}`
      if (watchDialColor) finalColor += `, dial: ${watchDialColor}`

      finalMaterial = watchStrapMaterial || 'Digital'
      if (watchDialShape) finalMaterial += ` (${watchDialShape} case)`
    } else if (category === 'ring') {
      finalColor = ringMetal || jewelryColor || 'Gold'
      if (ringStoneColor) finalColor += `, stone: ${ringStoneColor}`

      finalMaterial = ringStone || 'Metal'
      if (jewelryMaterial) finalMaterial += `, ${jewelryMaterial}`
    } else {
      finalColor = jewelryColor
      finalMaterial = jewelryMaterial
    }

    formData.append('color', finalColor)
    formData.append('material', finalMaterial)

    try {
      const response = await fetch(`${apiBase}/upload`, {
        method: 'POST',
        body: formData,
      })
      if (response.ok) {
        setSuccess(true)
        resetForm()
        fetchInventory()
        setTimeout(() => {
          setSuccess(false)
          setView('list')
        }, 1500)
      }
    } catch (err) {
      console.error('Upload failed:', err)
    } finally {
      setLoading(false)
    }
  }

  const handleEdit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (editIds.length === 0 || !name || !sku) return

    setLoading(true)
    try {
      for (const id of editIds) {
        const formData = new FormData()
        formData.append('name', name)
        formData.append('sku', sku)
        formData.append('color', jewelryColor)
        formData.append('material', jewelryMaterial)

        await fetch(`${apiBase}/inventory?id=${id}`, {
          method: 'PUT',
          body: formData,
        })
      }
      setSuccess(true)
      resetForm()
      fetchInventory()
      setTimeout(() => {
        setSuccess(false)
        setView('list')
      }, 1500)
    } catch (err) {
      console.error('Edit failed:', err)
    } finally {
      setLoading(false)
    }
  }

  const handleDelete = async (ids: string[]) => {
    if (!confirm(`Sei sicuro di voler eliminare questo articolo (incluse tutte le sue ${ids.length} viste)?`)) return
    try {
      for (const id of ids) {
        await fetch(`${apiBase}/inventory?id=${id}`, {
          method: 'DELETE',
        })
      }
      fetchInventory()
    } catch (err) {
      console.error('Delete failed:', err)
    }
  }

  // Helper function to dynamically guess categories from item metadata
  const getItemCategory = (item: InventoryItem) => {
    const pName = item.payload.name.toLowerCase()
    const pMaterial = (item.payload.material || '').toLowerCase()
    if (pName.includes('anello') || pName.includes('ring')) return 'Ring'
    if (pName.includes('collana') || pName.includes('necklace')) return 'Necklace'
    if (pName.includes('bracciale') || pName.includes('bracelet')) return 'Bracelet'
    if (pName.includes('orecchin') || pName.includes('earring')) return 'Earring'
    if (pName.includes('orologio') || pName.includes('watch') || pMaterial.includes('case') || pMaterial.includes('strap')) return 'Watch'
    return 'Other'
  }

  // Grouping logic
  const getGroupedItems = () => {
    const groupedMap: Record<string, GroupedInventoryItem> = {}
    items.forEach(item => {
      const sku = item.payload.sku || 'N/A'
      if (!groupedMap[sku]) {
        groupedMap[sku] = {
          sku: sku,
          name: item.payload.name,
          color: item.payload.color || '',
          material: item.payload.material || '',
          imageUrls: [],
          ids: [],
          category: getItemCategory(item)
        }
      }
      if (item.payload.imageUrl && !groupedMap[sku].imageUrls.includes(item.payload.imageUrl)) {
        groupedMap[sku].imageUrls.push(item.payload.imageUrl)
      }
      if (!groupedMap[sku].ids.includes(item.id)) {
        groupedMap[sku].ids.push(item.id)
      }
    })
    return Object.values(groupedMap)
  }

  const groupedItems = getGroupedItems()

  // Dashboard Stats Calculations
  const totalViews = items.length
  const uniqueProductsCount = groupedItems.length
  
  const categoryCounts = groupedItems.reduce((acc, item) => {
    const cat = item.category
    acc[cat] = (acc[cat] || 0) + 1
    return acc
  }, {} as Record<string, number>)

  const activeCategoriesCount = Object.keys(categoryCounts).length

  // Filter & Search Logic
  const filteredItems = groupedItems.filter(item => {
    const matchesSearch = 
      item.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
      item.sku.toLowerCase().includes(searchTerm.toLowerCase()) ||
      (item.color || '').toLowerCase().includes(searchTerm.toLowerCase()) ||
      (item.material || '').toLowerCase().includes(searchTerm.toLowerCase())
    
    const matchesCategory = 
      selectedFilterCategory === 'All' || 
      item.category === selectedFilterCategory

    return matchesSearch && matchesCategory
  })

  return (
    <div className="dashboard-wrapper">
      
      {/* Sidebar Navigation */}
      <aside className="sidebar">
        <div className="sidebar-header">
          <div className="sidebar-brand">
            <Package size={20} strokeWidth={2.5} />
            ShelfScan <span>Admin</span>
          </div>
        </div>

        <nav className="sidebar-nav">
          <button 
            onClick={() => setView('list')}
            className={`nav-item ${view === 'list' ? 'active' : ''}`}
          >
            <LayoutDashboard size={18} />
            Dashboard
          </button>
          
          <button 
            onClick={() => { resetForm(); setView('add'); }}
            className={`nav-item ${view === 'add' ? 'active' : ''}`}
          >
            <PlusCircle size={18} />
            Indicizza Nuovo Articolo
          </button>
        </nav>

        <div className="sidebar-footer">
          <div className="connection-pill">
            <span className="connection-dot"></span>
            Qdrant Connesso
          </div>
        </div>
      </aside>

      {/* Main Panel Content */}
      <main className="main-content">
        
        {/* Header */}
        <header className="page-header">
          <div className="page-title">
            <h1>Panoramica Inventario</h1>
            <p>Pannello di Controllo Amministratore • Indicizzatore Embeddings</p>
          </div>
          
          <div>
            <button 
              onClick={() => { resetForm(); setView('add'); }} 
              className="lj-btn lj-btn-black"
            >
              <PlusCircle size={15} />
              Indicizza Nuovo Articolo
            </button>
          </div>
        </header>

        {/* Bento Metrics Cards */}
        <section className="bento-grid">
          <div className="kpi-card">
            <div className="kpi-content">
              <span className="kpi-label">Viste Vettori Indicizzate</span>
              <span className="kpi-value">{totalViews}</span>
            </div>
            <Upload size={28} className="kpi-icon" />
          </div>

          <div className="kpi-card">
            <div className="kpi-content">
              <span className="kpi-label">Prodotti Distinti</span>
              <span className="kpi-value">{uniqueProductsCount}</span>
            </div>
            <Package size={28} className="kpi-icon" />
          </div>

          <div className="kpi-card">
            <div className="kpi-content">
              <span className="kpi-label">Categorie Attive</span>
              <span className="kpi-value">{activeCategoriesCount}</span>
            </div>
            <BarChart3 size={28} className="kpi-icon" />
          </div>

          <div className="kpi-card">
            <div className="kpi-content">
              <span className="kpi-label">Motore Database</span>
              <span className="kpi-value" style={{ fontSize: '1.25rem', fontWeight: 800, marginTop: '8px', color: 'var(--color-success)', letterSpacing: '0.05em' }}>
                QDRANT LOCALE
              </span>
            </div>
            <Database size={28} className="kpi-icon" style={{ color: 'var(--color-success)', opacity: 0.2 }} />
          </div>
        </section>

        {/* Layout Body Grid */}
        <div className="dashboard-layout">
          
          {/* Left Column: Inventory List/Grid */}
          <section className="inventory-section">
            
            {/* Unified Toolbar */}
            <div className="toolbar">
              <div className="search-container">
                <Search size={16} className="search-icon" />
                <input 
                  type="text"
                  placeholder="Cerca per Nome, SKU, Colore o Materiale..."
                  value={searchTerm}
                  onChange={(e) => setSearchTerm(e.target.value)}
                  className="search-input"
                />
              </div>

              <div className="toolbar-actions">
                {/* Layout Toggles */}
                <div className="layout-toggle">
                  <button 
                    onClick={() => setLayoutMode('grid')}
                    className={`toggle-btn ${layoutMode === 'grid' ? 'active' : ''}`}
                    title="Vista Griglia"
                  >
                    <Grid3X3 size={16} />
                  </button>
                  <button 
                    onClick={() => setLayoutMode('table')}
                    className={`toggle-btn ${layoutMode === 'table' ? 'active' : ''}`}
                    title="Vista Tabella"
                  >
                    <List size={16} />
                  </button>
                </div>
              </div>
            </div>

            {/* Category selection bar */}
            <div className="category-filters">
              {['All', 'Watch', 'Ring', 'Necklace', 'Bracelet', 'Earring', 'Other'].map((cat) => (
                <button
                  key={cat}
                  onClick={() => setSelectedFilterCategory(cat)}
                  className={`filter-tab ${selectedFilterCategory === cat ? 'active' : ''}`}
                >
                  {categoryLabels[cat] || cat}
                </button>
              ))}
            </div>

            {/* Grid layout */}
            {layoutMode === 'grid' && filteredItems.length > 0 && (
              <div className="inventory-grid">
                {filteredItems.map((item) => {
                  const cat = item.category
                  const mainImage = selectedImageMap[item.sku] || item.imageUrls[0] || ''
                  return (
                    <article key={item.sku} className="product-card">
                      <div>
                        {/* Image Container */}
                        <div className="product-img-box">
                          {mainImage ? (
                            <img src={mainImage} alt={item.name} className="product-img" loading="lazy" />
                          ) : (
                            <Package size={32} style={{ color: 'var(--color-gray-300)' }} />
                          )}
                          <span className="category-tag">{categoryLabels[cat] || cat}</span>
                        </div>

                        {/* Details */}
                        <div className="product-info">
                          <h3 className="product-title">{item.name}</h3>
                          <p className="product-sku">SKU: {item.sku}</p>

                          {/* Multi-view Thumbnail Row */}
                          {item.imageUrls.length > 1 && (
                            <div className="product-views-row" style={{ display: 'flex', gap: '4px', marginTop: '8px', marginBottom: '8px', overflowX: 'auto', paddingBottom: '4px' }}>
                              {item.imageUrls.map((url, idx) => (
                                <img
                                  key={idx}
                                  src={url}
                                  alt={`View ${idx}`}
                                  loading="lazy"
                                  onMouseEnter={() => setSelectedImageMap(prev => ({ ...prev, [item.sku]: url }))}
                                  style={{
                                    width: '28px',
                                    height: '28px',
                                    objectFit: 'cover',
                                    border: mainImage === url ? '2px solid black' : '1px solid #ddd',
                                    cursor: 'pointer',
                                    borderRadius: '4px',
                                    transition: 'all 0.1s ease'
                                  }}
                                />
                              ))}
                            </div>
                          )}

                          <div className="product-meta-list">
                            {item.color && (
                              <div className="meta-row">
                                <span className="meta-label">Colore:</span>
                                <span className="meta-value">{item.color}</span>
                              </div>
                            )}
                            {item.material && (
                              <div className="meta-row">
                                <span className="meta-label">Materiale:</span>
                                <span className="meta-value">{item.material}</span>
                              </div>
                            )}
                          </div>
                          <p className="qdrant-id" style={{ fontSize: '10px', color: 'var(--color-gray-400)', marginTop: '4px' }}>
                            {item.ids.length} viste Qdrant
                          </p>
                        </div>
                      </div>

                      {/* Card Actions */}
                      <div className="product-actions">
                        <button 
                          onClick={() => handleStartEdit(item)}
                          className="lj-btn-icon"
                          title="Modifica Prodotto"
                        >
                          <Pencil size={14} />
                        </button>
                        <button 
                          onClick={() => handleDelete(item.ids)}
                          className="lj-btn-icon lj-btn-icon-danger"
                          title="Elimina Prodotto"
                        >
                          <Trash2 size={14} />
                        </button>
                      </div>
                    </article>
                  )
                })}
              </div>
            )}

            {/* Table layout */}
            {layoutMode === 'table' && filteredItems.length > 0 && (
              <div className="table-responsive">
                <table className="lj-table">
                  <thead>
                    <tr>
                      <th style={{ width: '80px' }}>Anteprima</th>
                      <th>Nome Prodotto</th>
                      <th>SKU</th>
                      <th>Categoria</th>
                      <th>Dettagli Colore</th>
                      <th>Dettagli Materiale</th>
                      <th style={{ width: '100px', textAlign: 'right' }}>Azioni</th>
                    </tr>
                  </thead>
                  <tbody>
                    {filteredItems.map((item) => {
                      const cat = item.category
                      const mainImage = selectedImageMap[item.sku] || item.imageUrls[0] || ''
                      return (
                        <tr key={item.sku}>
                          <td>
                            <div style={{ display: 'flex', flexDirection: 'column', gap: '4px' }}>
                              <div className="table-thumb" style={{ width: '40px', height: '40px', overflow: 'hidden', borderRadius: '4px', border: '1px solid #eee' }}>
                                {mainImage ? (
                                  <img src={mainImage} alt={item.name} style={{ width: '100%', height: '100%', objectFit: 'cover' }} loading="lazy" />
                                ) : (
                                  <Package size={16} style={{ color: 'var(--color-gray-400)' }} />
                                )}
                              </div>
                              {item.imageUrls.length > 1 && (
                                <span style={{ fontSize: '9px', color: '#888', textAlign: 'center' }}>
                                  {item.imageUrls.length} viste
                                </span>
                              )}
                            </div>
                          </td>
                          <td style={{ fontWeight: 800, textTransform: 'uppercase', letterSpacing: '0.02em' }}>
                            {item.name}
                          </td>
                          <td style={{ fontFamily: 'monospace', fontWeight: 600 }}>
                            {item.sku}
                          </td>
                          <td>
                            <span style={{ fontSize: '11px', fontWeight: 900, textTransform: 'uppercase', letterSpacing: '0.05em', border: '1px solid #000', padding: '2px 6px' }}>
                              {categoryLabels[cat] || cat}
                            </span>
                          </td>
                          <td style={{ color: 'var(--color-secondary)' }}>{item.color || '—'}</td>
                          <td style={{ color: 'var(--color-secondary)' }}>{item.material || '—'}</td>
                          <td>
                            <div style={{ display: 'flex', justifyContent: 'flex-end', gap: '6px' }}>
                              <button 
                                onClick={() => handleStartEdit(item)}
                                className="lj-btn-icon"
                                title="Modifica Prodotto"
                              >
                                <Pencil size={13} />
                              </button>
                              <button 
                                onClick={() => handleDelete(item.ids)}
                                className="lj-btn-icon lj-btn-icon-danger"
                                title="Elimina Prodotto"
                              >
                                <Trash2 size={13} />
                              </button>
                            </div>
                          </td>
                        </tr>
                      )
                    })}
                  </tbody>
                </table>
              </div>
            )}

            {/* Empty State */}
            {filteredItems.length === 0 && (
              <div className="empty-state">
                <Package size={48} style={{ color: 'var(--color-gray-300)' }} />
                <h3 className="empty-state-title">Nessun Gioiello Trovato</h3>
                <p className="empty-state-subtitle">Modifica la query di ricerca o seleziona un'altra categoria</p>
                <button 
                  onClick={() => { setSearchTerm(''); setSelectedFilterCategory('All'); }}
                  style={{ marginTop: '16px', textDecoration: 'underline', color: 'black', border: 'none', background: 'none', fontWeight: 800, fontSize: '0.75rem', textTransform: 'uppercase', letterSpacing: '0.08em', cursor: 'pointer' }}
                >
                  Azzera tutti i filtri
                </button>
              </div>
            )}

          </section>

          {/* Right Column: Sidebar stats */}
          <aside className="distribution-sidebar">
            
            {/* Category distribution */}
            <div className="sidebar-card">
              <h3 className="sidebar-card-title">Distribuzione Inventario</h3>
              <div className="distribution-list">
                {['Watch', 'Ring', 'Necklace', 'Bracelet', 'Earring', 'Other'].map((cat) => {
                  const count = categoryCounts[cat] || 0
                  const pct = totalViews > 0 ? Math.round((count / totalViews) * 100) : 0
                  return (
                    <div key={cat} className="dist-item">
                      <div className="dist-info">
                        <span className="dist-label">{(categoryLabels[cat] || cat)}</span>
                        <span className="dist-count">{count} ({pct}%)</span>
                      </div>
                      <div className="progress-track">
                        <div className="progress-fill" style={{ width: `${pct}%` }}></div>
                      </div>
                    </div>
                  )
                })}
              </div>
            </div>

            {/* Database Technical Details */}
            <div className="sidebar-card">
              <h3 className="sidebar-card-title">Proprietà di Sistema</h3>
              <div className="db-info-list">
                <div className="db-info-row">
                  <span className="db-info-key">Motore Vettoriale:</span>
                  <span className="db-info-val">QDRANT DB</span>
                </div>
                <div className="db-info-row">
                  <span className="db-info-key">Collezione Target:</span>
                  <span className="db-info-val">jewelry_inventory</span>
                </div>
                <div className="db-info-row">
                  <span className="db-info-key">Metrica Distanza:</span>
                  <span className="db-info-val">Coseno</span>
                </div>
                <div className="db-info-row">
                  <span className="db-info-key">Modello di Embedding:</span>
                  <span className="db-info-val">Gemini Multimodal</span>
                </div>
                <div className="db-info-row">
                  <span className="db-info-key">Dimensione Vettoriale:</span>
                  <span className="db-info-val">768</span>
                </div>
              </div>
            </div>

          </aside>

        </div>
      </main>

      {/* Drawers Layout: slide-over overlays */}
      
      {/* 1. Add New Item Drawer */}
      {view === 'add' && (
        <div className="drawer-backdrop" onClick={() => { resetForm(); setView('list'); }}>
          <div className="drawer-content" onClick={(e) => e.stopPropagation()}>
            <div className="drawer-header">
              <h2 className="drawer-title">Indicizza Nuovo Articolo</h2>
              <button onClick={() => { resetForm(); setView('list'); }} className="drawer-close-btn">
                <X size={20} />
              </button>
            </div>

            <div className="drawer-body">
              {success && (
                <div className="lj-alert">
                  <CheckCircle2 size={16} /> Articolo Indicizzato con Successo!
                </div>
              )}

              <form onSubmit={handleUpload} style={{ display: 'flex', flexDirection: 'column', gap: '24px' }}>
                <div className="form-group">
                  <label className="form-label">Nome Prodotto</label>
                  <input 
                    type="text" 
                    value={name} 
                    onChange={(e) => setName(e.target.value)} 
                    placeholder="es. OROLOGIO LIU JO ELEGANTE ORO"
                    className="form-input"
                    required
                  />
                </div>

                <div className="form-group">
                  <label className="form-label">SKU (Codice Modello)</label>
                  <input 
                    type="text" 
                    value={sku} 
                    onChange={(e) => setSku(e.target.value)} 
                    placeholder="es. TLJ2642"
                    className="form-input"
                    required
                  />
                </div>

                <div className="form-group">
                  <label className="form-label">Categoria</label>
                  <div className="form-select-wrapper">
                    <select 
                      value={category} 
                      onChange={(e) => setCategory(e.target.value)} 
                      className="form-select"
                    >
                      <option value="watch">Orologio</option>
                      <option value="ring">Anello</option>
                      <option value="necklace">Collana</option>
                      <option value="bracelet">Bracciale</option>
                      <option value="earring">Orecchino</option>
                      <option value="other">Altro</option>
                    </select>
                    <ChevronDown size={16} className="form-select-arrow" />
                  </div>
                </div>

                {/* Category specific parameter selectors */}
                {category === 'watch' && (
                  <div className="category-properties">
                    <h4 className="category-properties-title">Proprietà Personalizzate Orologio</h4>
                    <h4 className="category-properties-title">ATTENZIONE: QUESTE PROPRIETÀ NON SONO OBBLIGATORIE, MA AIUTERANNO SHELFSCAN A RILEVARE CORRISPONDENZE MIGLIORI</h4>
                    <div className="form-row">
                      <div className="form-group" style={{ marginBottom: 0 }}>
                        <label className="form-label" style={{ fontSize: '10px' }}>Colore Cassa/Metallo</label>
                        <input 
                          type="text" 
                          value={jewelryColor} 
                          onChange={(e) => setJewelryColor(e.target.value)} 
                          placeholder="es. Oro"
                          className="form-input"
                          required
                        />
                      </div>
                      <div className="form-group" style={{ marginBottom: 0 }}>
                        <label className="form-label" style={{ fontSize: '10px' }}>Forma Cassa/Quadrante</label>
                        <div className="form-select-wrapper">
                          <select 
                            value={watchDialShape} 
                            onChange={(e) => setWatchDialShape(e.target.value)} 
                            className="form-select"
                          >
                            <option value="">Seleziona Forma...</option>
                            <option value="round">Rotonda</option>
                            <option value="square">Quadrata</option>
                            <option value="rectangular">Rettangolare</option>
                          </select>
                          <ChevronDown size={14} className="form-select-arrow" />
                        </div>
                      </div>
                    </div>

                    <div className="form-row">
                      <div className="form-group" style={{ marginBottom: 0 }}>
                        <label className="form-label" style={{ fontSize: '10px' }}>Colore Cinturino</label>
                        <input 
                          type="text" 
                          value={watchStrapColor} 
                          onChange={(e) => setWatchStrapColor(e.target.value)} 
                          placeholder="es. Rosa"
                          className="form-input"
                        />
                      </div>
                      <div className="form-group" style={{ marginBottom: 0 }}>
                        <label className="form-label" style={{ fontSize: '10px' }}>Materiale Cinturino</label>
                        <input 
                          type="text" 
                          value={watchStrapMaterial} 
                          onChange={(e) => setWatchStrapMaterial(e.target.value)} 
                          placeholder="es. Silicone"
                          className="form-input"
                        />
                      </div>
                    </div>

                    <div className="form-group" style={{ marginBottom: 0 }}>
                      <label className="form-label" style={{ fontSize: '10px' }}>Colore Quadrante</label>
                      <input 
                        type="text" 
                        value={watchDialColor} 
                        onChange={(e) => setWatchDialColor(e.target.value)} 
                        placeholder="es. Bianco"
                        className="form-input"
                      />
                    </div>
                  </div>
                )}

                {category === 'ring' && (
                  <div className="category-properties">
                    <h4 className="category-properties-title">Proprietà Personalizzate Anello</h4>
                    <div className="form-row">
                      <div className="form-group" style={{ marginBottom: 0 }}>
                        <label className="form-label" style={{ fontSize: '10px' }}>Colore Metallo</label>
                        <input 
                          type="text" 
                          value={ringMetal} 
                          onChange={(e) => setRingMetal(e.target.value)} 
                          placeholder="es. Oro Giallo"
                          className="form-input"
                          required
                        />
                      </div>
                      <div className="form-group" style={{ marginBottom: 0 }}>
                        <label className="form-label" style={{ fontSize: '10px' }}>Tipo di Pietra</label>
                        <input 
                          type="text" 
                          value={ringStone} 
                          onChange={(e) => setRingStone(e.target.value)} 
                          placeholder="es. Diamante"
                          className="form-input"
                        />
                      </div>
                    </div>

                    <div className="form-group" style={{ marginBottom: 0 }}>
                      <label className="form-label" style={{ fontSize: '10px' }}>Colore Pietra</label>
                      <input 
                        type="text" 
                        value={ringStoneColor} 
                        onChange={(e) => setRingStoneColor(e.target.value)} 
                        placeholder="es. Bianco"
                        className="form-input"
                      />
                    </div>
                  </div>
                )}

                {(category !== 'watch' && category !== 'ring') && (
                  <div className="category-properties">
                    <h4 className="category-properties-title">Specifiche del Gioiello</h4>
                    <div className="form-row">
                      <div className="form-group" style={{ marginBottom: 0 }}>
                        <label className="form-label" style={{ fontSize: '10px' }}>Colore Metallo</label>
                        <input 
                          type="text" 
                          value={jewelryColor} 
                          onChange={(e) => setJewelryColor(e.target.value)} 
                          placeholder="es. Argento"
                          className="form-input"
                          required
                        />
                      </div>
                      <div className="form-group" style={{ marginBottom: 0 }}>
                        <label className="form-label" style={{ fontSize: '10px' }}>Dettagli Stile</label>
                        <input 
                          type="text" 
                          value={jewelryMaterial} 
                          onChange={(e) => setJewelryMaterial(e.target.value)} 
                          placeholder="es. Catena, Tennis"
                          className="form-input"
                        />
                      </div>
                    </div>
                  </div>
                )}

                <div className="form-group">
                  <label className="form-label">Foto del Gioiello (Carica più angolazioni)</label>
                  <div className="dropzone" onClick={() => document.getElementById('drawer-file-input')?.click()}>
                    <Upload size={24} style={{ color: 'var(--color-gray-400)', margin: '0 auto' }} />
                    <p className="dropzone-title">{files ? `${files.length} Foto Selezionate` : 'Clicca per Caricare File'}</p>
                    <p className="dropzone-subtitle">Formati supportati: JPG, PNG, WebP</p>
                    <input 
                      id="drawer-file-input"
                      type="file" 
                      accept="image/*" 
                      multiple
                      onChange={(e) => setFiles(e.target.files)} 
                      style={{ display: 'none' }}
                    />
                  </div>
                </div>

                <div style={{ display: 'flex', flexDirection: 'column', gap: '12px', marginTop: '12px' }}>
                  <button 
                    type="submit" 
                    disabled={loading || !files || !name || !sku}
                    className="lj-btn lj-btn-black"
                    style={{ width: '100%' }}
                  >
                    {loading ? <Loader2 className="animate-spin" size={16} /> : <PlusCircle size={16} />}
                    {loading ? 'Generazione Embeddings...' : 'Indicizza Tutte le Viste'}
                  </button>
                  
                  <button 
                    type="button"
                    onClick={() => { resetForm(); setView('list'); }}
                    className="lj-btn lj-btn-outline"
                    style={{ width: '100%' }}
                  >
                    Annulla
                  </button>
                </div>
              </form>
            </div>
          </div>
        </div>
      )}

      {/* 2. Edit Item Drawer */}
      {view === 'edit' && (
        <div className="drawer-backdrop" onClick={() => { resetForm(); setView('list'); }}>
          <div className="drawer-content" onClick={(e) => e.stopPropagation()}>
            <div className="drawer-header">
              <h2 className="drawer-title">Modifica Articolo</h2>
              <button onClick={() => { resetForm(); setView('list'); }} className="drawer-close-btn">
                <X size={20} />
              </button>
            </div>

            <div className="drawer-body">
              {success && (
                <div className="lj-alert">
                  <CheckCircle2 size={16} /> Articolo aggiornato con successo!
                </div>
              )}

              <form onSubmit={handleEdit} style={{ display: 'flex', flexDirection: 'column', gap: '24px' }}>
                <div className="form-group">
                  <label className="form-label">Nome Prodotto</label>
                  <input 
                    type="text" 
                    value={name} 
                    onChange={(e) => setName(e.target.value)} 
                    placeholder="es. Anello con Diamante"
                    className="form-input"
                    required
                  />
                </div>

                <div className="form-group">
                  <label className="form-label">SKU (Codice Modello)</label>
                  <input 
                    type="text" 
                    value={sku} 
                    onChange={(e) => setSku(e.target.value)} 
                    placeholder="es. TLJ2642"
                    className="form-input"
                    required
                  />
                </div>

                <div className="category-properties">
                  <h4 className="category-properties-title">Specifiche</h4>
                  <div className="form-group">
                    <label className="form-label">Dettagli Colore</label>
                    <input 
                      type="text" 
                      value={jewelryColor} 
                      onChange={(e) => setJewelryColor(e.target.value)} 
                      placeholder="es. Oro, cinturino: Rosa"
                      className="form-input"
                      required
                    />
                  </div>
                  <div className="form-group" style={{ marginBottom: 0 }}>
                    <label className="form-label">Dettagli Materiale & Stile</label>
                    <input 
                      type="text" 
                      value={jewelryMaterial} 
                      onChange={(e) => setJewelryMaterial(e.target.value)} 
                      placeholder="es. Silicone (cassa rotonda)"
                      className="form-input"
                    />
                  </div>
                </div>

                <div style={{ display: 'flex', flexDirection: 'column', gap: '12px', marginTop: '12px' }}>
                  <button 
                    type="submit" 
                    disabled={loading || !name || !sku}
                    className="lj-btn lj-btn-black"
                    style={{ width: '100%' }}
                  >
                    {loading ? <Loader2 className="animate-spin" size={16} /> : <Pencil size={16} />}
                    {loading ? 'Salvataggio in corso...' : 'Salva Modifiche'}
                  </button>
                  
                  <button 
                    type="button"
                    onClick={() => { resetForm(); setView('list'); }}
                    className="lj-btn lj-btn-outline"
                    style={{ width: '100%' }}
                  >
                    Annulla
                  </button>
                </div>
              </form>
            </div>
          </div>
        </div>
      )}

    </div>
  )
}

export default App
