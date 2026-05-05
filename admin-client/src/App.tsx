import { useState } from 'react'
import { Upload, Package, CheckCircle2 } from 'lucide-react'

function App() {
  const [file, setFile] = useState<File | null>(null)
  const [name, setName] = useState('')
  const [uploading, setUploading] = useState(false)
  const [success, setSuccess] = useState(false)

  const handleUpload = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!file || !name) return

    setUploading(true)
    const formData = new FormData()
    formData.append('image', file)
    formData.append('name', name)

    try {
      const response = await fetch('http://localhost:8080/upload', {
        method: 'POST',
        body: formData,
      })
      if (response.ok) {
        setSuccess(true)
        setName('')
        setFile(null)
        setTimeout(() => setSuccess(false), 3000)
      }
    } catch (err) {
      console.error('Upload failed:', err)
    } finally {
      setUploading(false)
    }
  }

  return (
    <div style={{ maxWidth: '600px', margin: '40px auto', padding: '20px', fontFamily: 'sans-serif' }}>
      <h1>ShelfScan Inventory Admin</h1>
      
      {success && (
        <div style={{ padding: '10px', backgroundColor: '#dcfce7', color: '#166534', borderRadius: '8px', marginBottom: '20px', display: 'flex', alignItems: 'center', gap: '8px' }}>
          <CheckCircle2 size={20} />
          Jewelry item uploaded and indexed!
        </div>
      )}

      <form onSubmit={handleUpload} style={{ display: 'flex', flexDirection: 'column', gap: '20px' }}>
        <div>
          <label style={{ display: 'block', marginBottom: '8px' }}>Jewelry Name</label>
          <input 
            type="text" 
            value={name} 
            onChange={(e) => setName(e.target.value)} 
            placeholder="e.g. Diamond Ring 18k"
            style={{ width: '100%', padding: '8px', borderRadius: '4px', border: '1px solid #ccc' }}
            required
          />
        </div>

        <div style={{ border: '2px dashed #ccc', padding: '40px', textAlign: 'center', borderRadius: '8px', cursor: 'pointer' }} onClick={() => document.getElementById('file-input')?.click()}>
          <Upload size={40} style={{ color: '#666', marginBottom: '10px' }} />
          <p>{file ? file.name : 'Click to select jewelry photo'}</p>
          <input 
            id="file-input"
            type="file" 
            accept="image/*" 
            onChange={(e) => setFile(e.target.files?.[0] || null)} 
            style={{ display: 'none' }}
          />
        </div>

        <button 
          type="submit" 
          disabled={uploading || !file || !name}
          style={{ 
            padding: '12px', 
            backgroundColor: (uploading || !file || !name) ? '#ccc' : '#2563eb', 
            color: 'white', 
            border: 'none', 
            borderRadius: '4px', 
            cursor: 'pointer',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            gap: '8px'
          }}
        >
          {uploading ? 'Processing...' : (
            <>
              <Package size={20} />
              Add to Inventory
            </>
          )}
        </button>
      </form>
    </div>
  )
}

export default App
