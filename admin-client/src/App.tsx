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

const apiBase = import.meta.env.PROD ? '/api' : `http://${window.location.hostname}:8080`;

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
  const [editId, setEditId] = useState<string | null>(null)

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
    setEditId(null)
  }

  const handleStartEdit = (item: InventoryItem) => {
    setEditId(item.id)
    setName(item.payload.name)
    setSku(item.payload.sku || '')
    setJewelryColor(item.payload.color || '')
    setJewelryMaterial(item.payload.material || '')
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
    if (!editId || !name || !sku) return

    setLoading(true)
    const formData = new FormData()
    formData.append('name', name)
    formData.append('sku', sku)
    formData.append('color', jewelryColor)
    formData.append('material', jewelryMaterial)

    try {
      const response = await fetch(`${apiBase}/inventory?id=${editId}`, {
        method: 'PUT',
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
      console.error('Edit failed:', err)
    } finally {
      setLoading(false)
    }
  }

  const handleDelete = async (id: string) => {
    if (!confirm('Are you sure you want to delete this item?')) return
    try {
      const response = await fetch(`${apiBase}/inventory?id=${id}`, {
        method: 'DELETE',
      })
      if (response.ok) {
        fetchInventory()
      }
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

  // Dashboard Stats Calculations
  const totalViews = items.length
  const uniqueProductsCount = new Set(items.map(item => item.payload.sku || item.payload.name)).size
  
  const categoryCounts = items.reduce((acc, item) => {
    const cat = getItemCategory(item)
    acc[cat] = (acc[cat] || 0) + 1
    return acc
  }, {} as Record<string, number>)

  const activeCategoriesCount = Object.keys(categoryCounts).length

  // Filter & Search Logic
  const filteredItems = items.filter(item => {
    const matchesSearch = 
      item.payload.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
      (item.payload.sku || '').toLowerCase().includes(searchTerm.toLowerCase()) ||
      (item.payload.color || '').toLowerCase().includes(searchTerm.toLowerCase()) ||
      (item.payload.material || '').toLowerCase().includes(searchTerm.toLowerCase())
    
    const matchesCategory = 
      selectedFilterCategory === 'All' || 
      getItemCategory(item) === selectedFilterCategory

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
            Index New Item
          </button>
        </nav>

        <div className="sidebar-footer">
          <div className="connection-pill">
            <span className="connection-dot"></span>
            Qdrant Online
          </div>
        </div>
      </aside>

      {/* Main Panel Content */}
      <main className="main-content">
        
        {/* Header */}
        <header className="page-header">
          <div className="page-title">
            <h1>Inventory Overview</h1>
            <p>Admin Control Panel • Embeddings Indexer</p>
          </div>
          
          <div>
            <button 
              onClick={() => { resetForm(); setView('add'); }} 
              className="lj-btn lj-btn-black"
            >
              <PlusCircle size={15} />
              Index New Item
            </button>
          </div>
        </header>

        {/* Bento Metrics Cards */}
        <section className="bento-grid">
          <div className="kpi-card">
            <div className="kpi-content">
              <span className="kpi-label">Indexed Vector Views</span>
              <span className="kpi-value">{totalViews}</span>
            </div>
            <Upload size={28} className="kpi-icon" />
          </div>

          <div className="kpi-card">
            <div className="kpi-content">
              <span className="kpi-label">Distinct Products</span>
              <span className="kpi-value">{uniqueProductsCount}</span>
            </div>
            <Package size={28} className="kpi-icon" />
          </div>

          <div className="kpi-card">
            <div className="kpi-content">
              <span className="kpi-label">Active Categories</span>
              <span className="kpi-value">{activeCategoriesCount}</span>
            </div>
            <BarChart3 size={28} className="kpi-icon" />
          </div>

          <div className="kpi-card">
            <div className="kpi-content">
              <span className="kpi-label">Database Engine</span>
              <span className="kpi-value" style={{ fontSize: '1.25rem', fontWeight: 800, marginTop: '8px', color: 'var(--color-success)', letterSpacing: '0.05em' }}>
                QDRANT LOCAL
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
                  placeholder="Search by Name, SKU, Color or Material..."
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
                    title="Grid View"
                  >
                    <Grid3X3 size={16} />
                  </button>
                  <button 
                    onClick={() => setLayoutMode('table')}
                    className={`toggle-btn ${layoutMode === 'table' ? 'active' : ''}`}
                    title="Table View"
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
                  {cat}
                </button>
              ))}
            </div>

            {/* Grid layout */}
            {layoutMode === 'grid' && filteredItems.length > 0 && (
              <div className="inventory-grid">
                {filteredItems.map((item) => {
                  const cat = getItemCategory(item)
                  return (
                    <article key={item.id} className="product-card">
                      <div>
                        {/* Image Container */}
                        <div className="product-img-box">
                          {item.payload.imageUrl ? (
                            <img src={item.payload.imageUrl} alt={item.payload.name} className="product-img" />
                          ) : (
                            <Package size={32} style={{ color: 'var(--color-gray-300)' }} />
                          )}
                          <span className="category-tag">{cat}</span>
                        </div>

                        {/* Details */}
                        <div className="product-info">
                          <h3 className="product-title">{item.payload.name}</h3>
                          <p className="product-sku">SKU: {item.payload.sku || 'N/A'}</p>

                          <div className="product-meta-list">
                            {item.payload.color && (
                              <div className="meta-row">
                                <span className="meta-label">Color:</span>
                                <span className="meta-value">{item.payload.color}</span>
                              </div>
                            )}
                            {item.payload.material && (
                              <div className="meta-row">
                                <span className="meta-label">Material:</span>
                                <span className="meta-value">{item.payload.material}</span>
                              </div>
                            )}
                          </div>
                          <p className="qdrant-id">ID: {item.id}</p>
                        </div>
                      </div>

                      {/* Card Actions */}
                      <div className="product-actions">
                        <button 
                          onClick={() => handleStartEdit(item)}
                          className="lj-btn-icon"
                          title="Edit Item"
                        >
                          <Pencil size={14} />
                        </button>
                        <button 
                          onClick={() => handleDelete(item.id)}
                          className="lj-btn-icon lj-btn-icon-danger"
                          title="Delete Item"
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
                      <th style={{ width: '60px' }}>Preview</th>
                      <th>Product Name</th>
                      <th>SKU</th>
                      <th>Category</th>
                      <th>Color Details</th>
                      <th>Material Details</th>
                      <th style={{ width: '100px', textAlign: 'right' }}>Actions</th>
                    </tr>
                  </thead>
                  <tbody>
                    {filteredItems.map((item) => {
                      const cat = getItemCategory(item)
                      return (
                        <tr key={item.id}>
                          <td>
                            <div className="table-thumb">
                              {item.payload.imageUrl ? (
                                <img src={item.payload.imageUrl} alt={item.payload.name} />
                              ) : (
                                <Package size={16} style={{ color: 'var(--color-gray-400)' }} />
                              )}
                            </div>
                          </td>
                          <td style={{ fontWeight: 800, textTransform: 'uppercase', letterSpacing: '0.02em' }}>
                            {item.payload.name}
                          </td>
                          <td style={{ fontFamily: 'monospace', fontWeight: 600 }}>
                            {item.payload.sku || 'N/A'}
                          </td>
                          <td>
                            <span style={{ fontSize: '11px', fontWeight: 900, textTransform: 'uppercase', letterSpacing: '0.05em', border: '1px solid #000', padding: '2px 6px' }}>
                              {cat}
                            </span>
                          </td>
                          <td style={{ color: 'var(--color-secondary)' }}>{item.payload.color || '—'}</td>
                          <td style={{ color: 'var(--color-secondary)' }}>{item.payload.material || '—'}</td>
                          <td>
                            <div style={{ display: 'flex', justifyContent: 'flex-end', gap: '6px' }}>
                              <button 
                                onClick={() => handleStartEdit(item)}
                                className="lj-btn-icon"
                                title="Edit Item"
                              >
                                <Pencil size={13} />
                              </button>
                              <button 
                                onClick={() => handleDelete(item.id)}
                                className="lj-btn-icon lj-btn-icon-danger"
                                title="Delete Item"
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
                <h3 className="empty-state-title">No Jewelry Found</h3>
                <p className="empty-state-subtitle">Adjust your search query or select another category</p>
                <button 
                  onClick={() => { setSearchTerm(''); setSelectedFilterCategory('All'); }}
                  style={{ marginTop: '16px', textDecoration: 'underline', color: 'black', border: 'none', background: 'none', fontWeight: 800, fontSize: '0.75rem', textTransform: 'uppercase', letterSpacing: '0.08em', cursor: 'pointer' }}
                >
                  Clear all filters
                </button>
              </div>
            )}

          </section>

          {/* Right Column: Sidebar stats */}
          <aside className="distribution-sidebar">
            
            {/* Category distribution */}
            <div className="sidebar-card">
              <h3 className="sidebar-card-title">Inventory Distribution</h3>
              <div className="distribution-list">
                {['Watch', 'Ring', 'Necklace', 'Bracelet', 'Earring', 'Other'].map((cat) => {
                  const count = categoryCounts[cat] || 0
                  const pct = totalViews > 0 ? Math.round((count / totalViews) * 100) : 0
                  return (
                    <div key={cat} className="dist-item">
                      <div className="dist-info">
                        <span className="dist-label">{cat}s</span>
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
              <h3 className="sidebar-card-title">System Properties</h3>
              <div className="db-info-list">
                <div className="db-info-row">
                  <span className="db-info-key">Vector Engine:</span>
                  <span className="db-info-val">QDRANT DB</span>
                </div>
                <div className="db-info-row">
                  <span className="db-info-key">Target Collection:</span>
                  <span className="db-info-val">jewelry_inventory</span>
                </div>
                <div className="db-info-row">
                  <span className="db-info-key">Distance Metric:</span>
                  <span className="db-info-val">Cosine</span>
                </div>
                <div className="db-info-row">
                  <span className="db-info-key">Embedding Model:</span>
                  <span className="db-info-val">Gemini Multimodal</span>
                </div>
                <div className="db-info-row">
                  <span className="db-info-key">Vector Dimension:</span>
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
              <h2 className="drawer-title">Index New Item</h2>
              <button onClick={() => { resetForm(); setView('list'); }} className="drawer-close-btn">
                <X size={20} />
              </button>
            </div>

            <div className="drawer-body">
              {success && (
                <div className="lj-alert">
                  <CheckCircle2 size={16} /> Item Indexed Successfully!
                </div>
              )}

              <form onSubmit={handleUpload} style={{ display: 'flex', flexDirection: 'column', gap: '24px' }}>
                <div className="form-group">
                  <label className="form-label">Product Name</label>
                  <input 
                    type="text" 
                    value={name} 
                    onChange={(e) => setName(e.target.value)} 
                    placeholder="e.g. LIU JO ELEGANT WATCH GOLD"
                    className="form-input"
                    required
                  />
                </div>

                <div className="form-group">
                  <label className="form-label">SKU (Stock Keeping Unit)</label>
                  <input 
                    type="text" 
                    value={sku} 
                    onChange={(e) => setSku(e.target.value)} 
                    placeholder="e.g. TLJ2642"
                    className="form-input"
                    required
                  />
                </div>

                <div className="form-group">
                  <label className="form-label">Category</label>
                  <div className="form-select-wrapper">
                    <select 
                      value={category} 
                      onChange={(e) => setCategory(e.target.value)} 
                      className="form-select"
                    >
                      <option value="watch">Watch (Orologio)</option>
                      <option value="ring">Ring (Anello)</option>
                      <option value="necklace">Necklace (Collana)</option>
                      <option value="bracelet">Bracelet (Bracciale)</option>
                      <option value="earring">Earring (Orecchino)</option>
                      <option value="other">Other (Altro)</option>
                    </select>
                    <ChevronDown size={16} className="form-select-arrow" />
                  </div>
                </div>

                {/* Category specific parameter selectors */}
                {category === 'watch' && (
                  <div className="category-properties">
                    <h4 className="category-properties-title">Watch Custom Properties</h4>
                    <h4 className="category-properties-title">WARNING: THESE PROPERTIES ARE NOT REQUIRED, BUT IT WILL HELP SHELFSCAN TO DETECT BETTER MATCHES</h4>
                    <div className="form-row">
                      <div className="form-group" style={{ marginBottom: 0 }}>
                        <label className="form-label" style={{ fontSize: '10px' }}>Case/Metal Color</label>
                        <input 
                          type="text" 
                          value={jewelryColor} 
                          onChange={(e) => setJewelryColor(e.target.value)} 
                          placeholder="e.g. Gold"
                          className="form-input"
                          required
                        />
                      </div>
                      <div className="form-group" style={{ marginBottom: 0 }}>
                        <label className="form-label" style={{ fontSize: '10px' }}>Case/Dial Shape</label>
                        <div className="form-select-wrapper">
                          <select 
                            value={watchDialShape} 
                            onChange={(e) => setWatchDialShape(e.target.value)} 
                            className="form-select"
                          >
                            <option value="">Select Shape...</option>
                            <option value="round">Round</option>
                            <option value="square">Square</option>
                            <option value="rectangular">Rectangular</option>
                          </select>
                          <ChevronDown size={14} className="form-select-arrow" />
                        </div>
                      </div>
                    </div>

                    <div className="form-row">
                      <div className="form-group" style={{ marginBottom: 0 }}>
                        <label className="form-label" style={{ fontSize: '10px' }}>Strap Color</label>
                        <input 
                          type="text" 
                          value={watchStrapColor} 
                          onChange={(e) => setWatchStrapColor(e.target.value)} 
                          placeholder="e.g. Pink"
                          className="form-input"
                        />
                      </div>
                      <div className="form-group" style={{ marginBottom: 0 }}>
                        <label className="form-label" style={{ fontSize: '10px' }}>Strap Material</label>
                        <input 
                          type="text" 
                          value={watchStrapMaterial} 
                          onChange={(e) => setWatchStrapMaterial(e.target.value)} 
                          placeholder="e.g. Silicone"
                          className="form-input"
                        />
                      </div>
                    </div>

                    <div className="form-group" style={{ marginBottom: 0 }}>
                      <label className="form-label" style={{ fontSize: '10px' }}>Dial Color</label>
                      <input 
                        type="text" 
                        value={watchDialColor} 
                        onChange={(e) => setWatchDialColor(e.target.value)} 
                        placeholder="e.g. White"
                        className="form-input"
                      />
                    </div>
                  </div>
                )}

                {category === 'ring' && (
                  <div className="category-properties">
                    <h4 className="category-properties-title">Ring Custom Properties</h4>
                    <div className="form-row">
                      <div className="form-group" style={{ marginBottom: 0 }}>
                        <label className="form-label" style={{ fontSize: '10px' }}>Metal Color</label>
                        <input 
                          type="text" 
                          value={ringMetal} 
                          onChange={(e) => setRingMetal(e.target.value)} 
                          placeholder="e.g. Yellow Gold"
                          className="form-input"
                          required
                        />
                      </div>
                      <div className="form-group" style={{ marginBottom: 0 }}>
                        <label className="form-label" style={{ fontSize: '10px' }}>Stone Type</label>
                        <input 
                          type="text" 
                          value={ringStone} 
                          onChange={(e) => setRingStone(e.target.value)} 
                          placeholder="e.g. Diamond"
                          className="form-input"
                        />
                      </div>
                    </div>

                    <div className="form-group" style={{ marginBottom: 0 }}>
                      <label className="form-label" style={{ fontSize: '10px' }}>Stone Color</label>
                      <input 
                        type="text" 
                        value={ringStoneColor} 
                        onChange={(e) => setRingStoneColor(e.target.value)} 
                        placeholder="e.g. White"
                        className="form-input"
                      />
                    </div>
                  </div>
                )}

                {(category !== 'watch' && category !== 'ring') && (
                  <div className="category-properties">
                    <h4 className="category-properties-title">Jewelry Specifications</h4>
                    <div className="form-row">
                      <div className="form-group" style={{ marginBottom: 0 }}>
                        <label className="form-label" style={{ fontSize: '10px' }}>Metal Color</label>
                        <input 
                          type="text" 
                          value={jewelryColor} 
                          onChange={(e) => setJewelryColor(e.target.value)} 
                          placeholder="e.g. Silver"
                          className="form-input"
                          required
                        />
                      </div>
                      <div className="form-group" style={{ marginBottom: 0 }}>
                        <label className="form-label" style={{ fontSize: '10px' }}>Style Details</label>
                        <input 
                          type="text" 
                          value={jewelryMaterial} 
                          onChange={(e) => setJewelryMaterial(e.target.value)} 
                          placeholder="e.g. Chain, Tennis"
                          className="form-input"
                        />
                      </div>
                    </div>
                  </div>
                )}

                <div className="form-group">
                  <label className="form-label">Jewelry Photos (Upload Multiple Angles)</label>
                  <div className="dropzone" onClick={() => document.getElementById('drawer-file-input')?.click()}>
                    <Upload size={24} style={{ color: 'var(--color-gray-400)', margin: '0 auto' }} />
                    <p className="dropzone-title">{files ? `${files.length} Photos Selected` : 'Click to Upload Files'}</p>
                    <p className="dropzone-subtitle">Supported formats: JPG, PNG, WebP</p>
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
                    {loading ? 'Generating Embeddings...' : 'Index All Views'}
                  </button>
                  
                  <button 
                    type="button"
                    onClick={() => { resetForm(); setView('list'); }}
                    className="lj-btn lj-btn-outline"
                    style={{ width: '100%' }}
                  >
                    Cancel
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
              <h2 className="drawer-title">Edit Item</h2>
              <button onClick={() => { resetForm(); setView('list'); }} className="drawer-close-btn">
                <X size={20} />
              </button>
            </div>

            <div className="drawer-body">
              {success && (
                <div className="lj-alert">
                  <CheckCircle2 size={16} /> Item updated successfully!
                </div>
              )}

              <form onSubmit={handleEdit} style={{ display: 'flex', flexDirection: 'column', gap: '24px' }}>
                <div className="form-group">
                  <label className="form-label">Product Name</label>
                  <input 
                    type="text" 
                    value={name} 
                    onChange={(e) => setName(e.target.value)} 
                    placeholder="e.g. Diamond Ring"
                    className="form-input"
                    required
                  />
                </div>

                <div className="form-group">
                  <label className="form-label">SKU (Stock Keeping Unit)</label>
                  <input 
                    type="text" 
                    value={sku} 
                    onChange={(e) => setSku(e.target.value)} 
                    placeholder="e.g. TLJ2642"
                    className="form-input"
                    required
                  />
                </div>

                <div className="category-properties">
                  <h4 className="category-properties-title">Specifications</h4>
                  <div className="form-group">
                    <label className="form-label">Color Details</label>
                    <input 
                      type="text" 
                      value={jewelryColor} 
                      onChange={(e) => setJewelryColor(e.target.value)} 
                      placeholder="e.g. Gold, strap: Pink"
                      className="form-input"
                      required
                    />
                  </div>
                  <div className="form-group" style={{ marginBottom: 0 }}>
                    <label className="form-label">Material & Style Details</label>
                    <input 
                      type="text" 
                      value={jewelryMaterial} 
                      onChange={(e) => setJewelryMaterial(e.target.value)} 
                      placeholder="e.g. Silicone (round case)"
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
                    {loading ? 'Saving Changes...' : 'Save Changes'}
                  </button>
                  
                  <button 
                    type="button"
                    onClick={() => { resetForm(); setView('list'); }}
                    className="lj-btn lj-btn-outline"
                    style={{ width: '100%' }}
                  >
                    Cancel
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
