import { useState, useEffect } from 'react'
import { Upload, Package, CheckCircle2, Trash2, LayoutDashboard, PlusCircle, Loader2 } from 'lucide-react'

interface InventoryItem {
  id: number
  payload: {
    name: string
  }
}

function App() {
  const [items, setItems] = useState<InventoryItem[]>([])
  const [view, setView] = useState<'list' | 'add'>('list')
  const [files, setFiles] = useState<FileList | null>(null)
  const [name, setName] = useState('')
  const [loading, setLoading] = useState(false)
  const [success, setSuccess] = useState(false)

  const fetchInventory = async () => {
    try {
      const apiHost = window.location.hostname
      const response = await fetch(`http://${apiHost}:8080/inventory`)
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

  const handleUpload = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!files || !name) return

    setLoading(true)
    const formData = new FormData()
    for (let i = 0; i < files.length; i++) {
      formData.append('images', files[i])
    }
    formData.append('name', name)

    try {
      const apiHost = window.location.hostname
      const response = await fetch(`http://${apiHost}:8080/upload`, {
        method: 'POST',
        body: formData,
      })
      if (response.ok) {
        setSuccess(true)
        setName('')
        setFiles(null)
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

  const handleDelete = async (id: number) => {
    if (!confirm('Are you sure you want to delete this item?')) return
    try {
      const apiHost = window.location.hostname
      const response = await fetch(`http://${apiHost}:8080/inventory?id=${id}`, {
        method: 'DELETE',
      })
      if (response.ok) {
        fetchInventory()
      }
    } catch (err) {
      console.error('Delete failed:', err)
    }
  }

  return (
    <div style={{ minHeight: '100vh', backgroundColor: '#f8fafc', color: '#1e293b', fontFamily: 'system-ui, sans-serif' }}>
      {/* Navbar */}
      <nav style={{ backgroundColor: 'white', borderBottom: '1px solid #e2e8f0', padding: '0 20px' }}>
        <div style={{ maxWidth: '1000px', margin: '0 auto', height: '64px', display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: '10px', fontSize: '1.25rem', fontWeight: 700, color: '#2563eb' }}>
            <Package size={24} />
            <span>ShelfScan Admin</span>
          </div>
          <div style={{ display: 'flex', gap: '20px' }}>
            <button 
              onClick={() => setView('list')}
              style={{ background: 'none', border: 'none', color: view === 'list' ? '#2563eb' : '#64748b', fontWeight: 600, cursor: 'pointer', display: 'flex', alignItems: 'center', gap: '6px' }}
            >
              <LayoutDashboard size={18} /> Dashboard
            </button>
            <button 
              onClick={() => setView('add')}
              style={{ background: 'none', border: 'none', color: view === 'add' ? '#2563eb' : '#64748b', fontWeight: 600, cursor: 'pointer', display: 'flex', alignItems: 'center', gap: '6px' }}
            >
              <PlusCircle size={18} /> New Item
            </button>
          </div>
        </div>
      </nav>

      <main style={{ maxWidth: '1000px', margin: '40px auto', padding: '0 20px' }}>
        {view === 'list' ? (
          <div>
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '24px' }}>
              <h2 style={{ fontSize: '1.5rem', fontWeight: 700 }}>Inventory Overview</h2>
              <span style={{ backgroundColor: '#e2e8f0', padding: '4px 12px', borderRadius: '20px', fontSize: '0.875rem', color: '#475569' }}>
                {items.length} Views Indexed
              </span>
            </div>

            <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fill, minmax(280px, 1fr))', gap: '20px' }}>
              {items.map((item) => (
                <div key={item.id} style={{ backgroundColor: 'white', borderRadius: '12px', padding: '20px', border: '1px solid #e2e8f0', display: 'flex', flexDirection: 'column', justifyContent: 'space-between', boxShadow: '0 1px 3px rgba(0,0,0,0.1)' }}>
                  <div>
                    <div style={{ width: '40px', height: '40px', backgroundColor: '#eff6ff', borderRadius: '8px', display: 'flex', alignItems: 'center', justifySelf: 'center', justifyContent: 'center', color: '#2563eb', marginBottom: '16px' }}>
                      <Package size={20} />
                    </div>
                    <h3 style={{ margin: 0, fontSize: '1.125rem', fontWeight: 600 }}>{item.payload.name}</h3>
                    <p style={{ color: '#64748b', fontSize: '0.875rem', marginTop: '4px' }}>ID: {item.id}</p>
                  </div>
                  <button 
                    onClick={() => handleDelete(item.id)}
                    style={{ alignSelf: 'flex-end', background: 'none', border: 'none', color: '#ef4444', cursor: 'pointer', padding: '8px', borderRadius: '6px', transition: 'background 0.2s' }}
                    onMouseOver={(e) => (e.currentTarget.style.backgroundColor = '#fee2e2')}
                    onMouseOut={(e) => (e.currentTarget.style.backgroundColor = 'transparent')}
                  >
                    <Trash2 size={18} />
                  </button>
                </div>
              ))}
              {items.length === 0 && (
                <div style={{ gridColumn: '1 / -1', textAlign: 'center', padding: '60px', backgroundColor: 'white', borderRadius: '12px', border: '1px dashed #cbd5e1' }}>
                  <Package size={48} style={{ color: '#cbd5e1', marginBottom: '16px' }} />
                  <p style={{ color: '#64748b' }}>No items in inventory yet.</p>
                  <button onClick={() => setView('add')} style={{ marginTop: '12px', color: '#2563eb', border: 'none', background: 'none', fontWeight: 600, cursor: 'pointer' }}>Add your first item</button>
                </div>
              )}
            </div>
          </div>
        ) : (
          <div style={{ maxWidth: '600px', margin: '0 auto', backgroundColor: 'white', padding: '40px', borderRadius: '16px', boxShadow: '0 4px 6px -1px rgba(0,0,0,0.1)' }}>
            <h2 style={{ fontSize: '1.5rem', fontWeight: 700, marginBottom: '8px' }}>Add New Jewelry</h2>
            <p style={{ color: '#64748b', marginBottom: '32px' }}>Upload multiple photos to generate better AI embeddings.</p>
            
            {success && (
              <div style={{ padding: '12px 16px', backgroundColor: '#dcfce7', color: '#166534', borderRadius: '8px', marginBottom: '24px', display: 'flex', alignItems: 'center', gap: '8px', fontWeight: 500 }}>
                <CheckCircle2 size={20} /> Views successfully indexed!
              </div>
            )}

            <form onSubmit={handleUpload} style={{ display: 'flex', flexDirection: 'column', gap: '24px' }}>
              <div>
                <label style={{ display: 'block', marginBottom: '8px', fontWeight: 600, fontSize: '0.875rem' }}>Jewelry Name</label>
                <input 
                  type="text" 
                  value={name} 
                  onChange={(e) => setName(e.target.value)} 
                  placeholder="e.g. Diamond Ring 18k White Gold"
                  style={{ width: '100%', padding: '12px', borderRadius: '8px', border: '1px solid #cbd5e1', fontSize: '1rem', boxSizing: 'border-box' }}
                  required
                />
              </div>

              <div>
                <label style={{ display: 'block', marginBottom: '8px', fontWeight: 600, fontSize: '0.875rem' }}>Jewelry Photos (Select Multiple)</label>
                <div 
                  style={{ border: '2px dashed #cbd5e1', padding: '40px', textAlign: 'center', borderRadius: '12px', cursor: 'pointer', transition: 'border-color 0.2s' }} 
                  onMouseOver={(e) => (e.currentTarget.style.borderColor = '#2563eb')}
                  onMouseOut={(e) => (e.currentTarget.style.borderColor = '#cbd5e1')}
                  onClick={() => document.getElementById('file-input')?.click()}
                >
                  <Upload size={40} style={{ color: '#94a3b8', marginBottom: '12px' }} />
                  <p style={{ margin: 0, fontWeight: 500 }}>{files ? `${files.length} photos selected` : 'Click or drag photos here'}</p>
                  <p style={{ margin: '4px 0 0 0', fontSize: '0.75rem', color: '#94a3b8' }}>Upload different angles for better AI recognition</p>
                  <input 
                    id="file-input"
                    type="file" 
                    accept="image/*" 
                    multiple
                    onChange={(e) => setFiles(e.target.files)} 
                    style={{ display: 'none' }}
                  />
                </div>
              </div>

              <button 
                type="submit" 
                disabled={loading || !files || !name}
                style={{ 
                  padding: '14px', 
                  backgroundColor: (loading || !files || !name) ? '#cbd5e1' : '#2563eb', 
                  color: 'white', 
                  border: 'none', 
                  borderRadius: '8px', 
                  cursor: (loading || !files || !name) ? 'not-allowed' : 'pointer',
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'center',
                  gap: '10px',
                  fontSize: '1rem',
                  fontWeight: 600,
                  transition: 'background 0.2s'
                }}
              >
                {loading ? <Loader2 className="animate-spin" size={20} /> : <PlusCircle size={20} />}
                {loading ? 'Generating Embeddings...' : 'Index All Views'}
              </button>
              
              <button 
                type="button"
                onClick={() => setView('list')}
                style={{ color: '#64748b', border: 'none', background: 'none', fontSize: '0.875rem', cursor: 'pointer' }}
              >
                Cancel and return to dashboard
              </button>
            </form>
          </div>
        )}
      </main>
    </div>
  )
}

export default App
