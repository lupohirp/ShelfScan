import React, { useState, useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import { useAuth } from '../store/auth'
import PageShell from '../components/PageShell'
import { getApiUrl } from '../lib/api'
import { ArrowLeft, Plus, Trash2, Camera, Check } from 'lucide-react'

interface StoreSearchResult {
  id: string
  code: string
  name: string
  province: string
  province_name: string
  address: string
  city: string
  agent_name: string
}

export default function CustomizationForm() {
  const navigate = useNavigate()
  const { user } = useAuth()
  const agentName = user ? `${user.firstName} ${user.lastName}` : ''
  const agentId = user?.id || ''

  // Form states
  const [customerCode, setCustomerCode] = useState('')
  const [customerBusinessName, setCustomerBusinessName] = useState('')
  const [customerStoreName, setCustomerStoreName] = useState('')
  const [customerAddress, setCustomerAddress] = useState('')
  const [customerCap, setCustomerCap] = useState('')
  const [customerCity, setCustomerCity] = useState('')
  const [customerEmail, setCustomerEmail] = useState('')
  const [customerPhone, setCustomerPhone] = useState('')
  const [annualSellInEstimate, setAnnualSellInEstimate] = useState('< 2.000')

  // Search stores state
  const [searchQuery, setSearchQuery] = useState('')
  const [searchResults, setSearchResults] = useState<StoreSearchResult[]>([])
  const [showSearchResults, setShowSearchResults] = useState(false)

  // Customizations array (1 is mandatory, up to 3)
  const [customizationCount, setCustomizationCount] = useState(1)

  const [cust1Subject, setCust1Subject] = useState('Nuove collezioni')
  const [cust1SubjectOther, setCust1SubjectOther] = useState('')
  const [cust1Type, setCust1Type] = useState('Affissione')
  const [cust1TypeOther, setCust1TypeOther] = useState('')
  const [cust1Width, setCust1Width] = useState('')
  const [cust1Height, setCust1Height] = useState('')
  const [cust1Material, setCust1Material] = useState('Pannello in forex')
  const [cust1MaterialOther, setCust1MaterialOther] = useState('')

  const [cust2Subject, setCust2Subject] = useState('Nuove collezioni')
  const [cust2SubjectOther, setCust2SubjectOther] = useState('')
  const [cust2Type, setCust2Type] = useState('Affissione')
  const [cust2TypeOther, setCust2TypeOther] = useState('')
  const [cust2Width, setCust2Width] = useState('')
  const [cust2Height, setCust2Height] = useState('')
  const [cust2Material, setCust2Material] = useState('Pannello in forex')
  const [cust2MaterialOther, setCust2MaterialOther] = useState('')

  const [cust3Subject, setCust3Subject] = useState('Nuove collezioni')
  const [cust3SubjectOther, setCust3SubjectOther] = useState('')
  const [cust3Type, setCust3Type] = useState('Affissione')
  const [cust3TypeOther, setCust3TypeOther] = useState('')
  const [cust3Width, setCust3Width] = useState('')
  const [cust3Height, setCust3Height] = useState('')
  const [cust3Material, setCust3Material] = useState('Pannello in forex')
  const [cust3MaterialOther, setCust3MaterialOther] = useState('')

  const [startDate, setStartDate] = useState('')
  const [endDate, setEndDate] = useState('')

  const [printingCost, setPrintingCost] = useState('Stampa a carico del cliente')
  const [assemblyCost, setAssemblyCost] = useState('Montaggio a carico del cliente')

  const [shippingAddress, setShippingAddress] = useState('')
  const [shippingCivic, setShippingCivic] = useState('')
  const [shippingCity, setShippingCity] = useState('')
  const [shippingProvince, setShippingProvince] = useState('')
  const [shippingCap, setShippingCap] = useState('')

  // Shipping addresses autofill checkbox
  const [autofillShipping, setAutofillShipping] = useState(false)

  // Photo
  const [photoFile, setPhotoFile] = useState<File | null>(null)
  const [photoPreview, setPhotoPreview] = useState<string>('')

  // UI state
  const [submitting, setSubmitting] = useState(false)
  const [success, setSuccess] = useState(false)
  const [error, setError] = useState('')

  // Search stores when query changes
  useEffect(() => {
    if (!searchQuery) {
      setSearchResults([])
      return
    }
    const apiBase = getApiUrl()
    const delayDebounce = setTimeout(async () => {
      try {
        const res = await fetch(`${apiBase}/stores?q=${encodeURIComponent(searchQuery)}`)
        if (res.ok) {
          const data = await res.json()
          setSearchResults(data.slice(0, 5))
        }
      } catch (err) {
        console.error('Error fetching stores for autofill:', err)
      }
    }, 250)

    return () => clearTimeout(delayDebounce)
  }, [searchQuery])

  // Autofill shipping address if checked
  useEffect(() => {
    if (autofillShipping) {
      setShippingAddress(customerAddress)
      setShippingCity(customerCity)
      setShippingCap(customerCap)
      // Attempt to extract civic if present in address
      const parts = customerAddress.split(',')
      if (parts.length > 1) {
        setShippingAddress(parts[0].trim())
        setShippingCivic(parts[1].trim())
      }
    }
  }, [autofillShipping, customerAddress, customerCity, customerCap])

  const handleSelectStore = (store: StoreSearchResult) => {
    setCustomerCode(store.code)
    setCustomerBusinessName(store.name)
    setCustomerStoreName(store.name)
    setCustomerAddress(store.address)
    setCustomerCity(store.city)
    // CAP and Phone aren't always present in the stores list, but we set what we have
    setSearchQuery('')
    setShowSearchResults(false)
  }

  const handlePhotoChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0]
    if (file) {
      setPhotoFile(file)
      const reader = new FileReader()
      reader.onloadend = () => {
        setPhotoPreview(reader.result as string)
      }
      reader.readAsDataURL(file)
    }
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setError('')

    // Basic validation
    if (
      !customerCode ||
      !customerBusinessName ||
      !customerStoreName ||
      !customerAddress ||
      !customerCap ||
      !customerCity ||
      !customerEmail ||
      !customerPhone ||
      !cust1Width ||
      !cust1Height ||
      !shippingAddress ||
      !shippingCivic ||
      !shippingCity ||
      !shippingProvince ||
      !shippingCap ||
      !photoFile
    ) {
      setError('Compila tutti i campi obbligatori contrassegnati con * e carica la foto.')
      window.scrollTo({ top: 0, behavior: 'smooth' })
      return
    }

    setSubmitting(true)

    // Build form data
    const formData = new FormData()
    formData.append('agent_id', agentId)
    formData.append('agent_name', agentName)
    formData.append('customer_code', customerCode)
    formData.append('customer_business_name', customerBusinessName)
    formData.append('customer_store_name', customerStoreName)
    formData.append('customer_address', customerAddress)
    formData.append('customer_cap', customerCap)
    formData.append('customer_city', customerCity)
    formData.append('customer_email', customerEmail)
    formData.append('customer_phone', customerPhone)
    formData.append('annual_sell_in_estimate', annualSellInEstimate)

    // Cust 1
    const subject1 = cust1Subject === 'Altro:' ? cust1SubjectOther : cust1Subject
    const type1 = cust1Type === 'Altro:' ? cust1TypeOther : cust1Type
    const mat1 = cust1Material === 'Altro:' ? cust1MaterialOther : cust1Material
    formData.append('cust1_subject', subject1)
    formData.append('cust1_type', type1)
    formData.append('cust1_width_cm', cust1Width)
    formData.append('cust1_height_cm', cust1Height)
    formData.append('cust1_material', mat1)

    // Cust 2
    if (customizationCount >= 2) {
      const subject2 = cust2Subject === 'Altro:' ? cust2SubjectOther : cust2Subject
      const type2 = cust2Type === 'Altro:' ? cust2TypeOther : cust2Type
      const mat2 = cust2Material === 'Altro:' ? cust2MaterialOther : cust2Material
      formData.append('cust2_subject', subject2)
      formData.append('cust2_type', type2)
      if (cust2Width) formData.append('cust2_width_cm', cust2Width)
      if (cust2Height) formData.append('cust2_height_cm', cust2Height)
      formData.append('cust2_material', mat2)
    }

    // Cust 3
    if (customizationCount >= 3) {
      const subject3 = cust3Subject === 'Altro:' ? cust3SubjectOther : cust3Subject
      const type3 = cust3Type === 'Altro:' ? cust3TypeOther : cust3Type
      const mat3 = cust3Material === 'Altro:' ? cust3MaterialOther : cust3Material
      formData.append('cust3_subject', subject3)
      formData.append('cust3_type', type3)
      if (cust3Width) formData.append('cust3_width_cm', cust3Width)
      if (cust3Height) formData.append('cust3_height_cm', cust3Height)
      formData.append('cust3_material', mat3)
    }

    formData.append('start_date', startDate)
    formData.append('end_date', endDate)

    formData.append('printing_cost_responsibility', printingCost)
    formData.append('assembly_cost_responsibility', assemblyCost)

    formData.append('shipping_address', shippingAddress)
    formData.append('shipping_civic', shippingCivic)
    formData.append('shipping_city', shippingCity)
    formData.append('shipping_province', shippingProvince)
    formData.append('shipping_cap', shippingCap)

    formData.append('photo', photoFile)

    try {
      const apiBase = getApiUrl()
      const res = await fetch(`${apiBase}/customizations`, {
        method: 'POST',
        body: formData,
      })

      if (!res.ok) {
        const txt = await res.text()
        throw new Error(txt || 'Errore durante il salvataggio')
      }

      setSuccess(true)
      window.scrollTo({ top: 0, behavior: 'smooth' })
    } catch (err: any) {
      console.error(err)
      setError(err.message || "Errore di connessione col server.")
    } finally {
      setSubmitting(false)
    }
  }

  if (success) {
    return (
      <PageShell>
        <div className="px-8 pt-16 pb-32 flex flex-col items-center justify-center min-h-[70vh] text-center animate-fade-in">
          <div className="w-16 h-16 bg-black flex items-center justify-center text-white mb-6">
            <Check size={28} />
          </div>
          <h1 className="text-2xl font-black uppercase tracking-wider mb-2">Richiesta Inviata</h1>
          <p className="text-gray-500 text-sm max-w-sm mb-10 uppercase tracking-wide leading-relaxed">
            La richiesta di personalizzazione per il cliente {customerStoreName} è stata salvata con successo.
          </p>
          <button
            onClick={() => navigate('/home')}
            className="lj-button w-full max-w-[280px]"
          >
            Torna alla Home
          </button>
        </div>
      </PageShell>
    )
  }

  return (
    <PageShell>
      {/* Header */}
      <div className="px-8 pt-16 pb-8 safe-top border-b border-gray-100 bg-white flex items-center gap-4">
        <button onClick={() => navigate('/home')} className="p-1 -ml-1 text-black">
          <ArrowLeft size={24} />
        </button>
        <div>
          <h1 className="text-[24px] font-black tracking-tight text-black leading-none uppercase">PERSONALIZZAZIONI</h1>
          <p className="text-gray-400 text-[9px] font-black uppercase tracking-[0.2em] mt-1.5">Richiesta allestimento punto vendita</p>
        </div>
      </div>

      <form onSubmit={handleSubmit} className="px-8 py-8 space-y-10 pb-32">
        {error && (
          <div className="p-4 bg-red-50 border border-red-200 text-red-700 text-xs font-bold uppercase tracking-wider">
            {error}
          </div>
        )}

        {/* Read-only Agent Section */}
        <div className="lj-card p-5 bg-gray-50/50">
          <label className="block text-[9px] font-black uppercase tracking-[0.2em] text-gray-400 mb-1">Rappresentante / Agente</label>
          <p className="text-[14px] font-black uppercase tracking-wider text-black">{agentName || 'Non autenticato'}</p>
        </div>

        {/* Store search autocomplete */}
        <div className="space-y-4">
          <h2 className="text-[11px] font-black uppercase tracking-[0.25em] text-black border-b border-black pb-2">Seleziona Cliente esistente (Opzionale)</h2>
          <div className="relative">
            <input
              type="text"
              placeholder="Cerca cliente per nome o codice..."
              value={searchQuery}
              onChange={(e) => {
                setSearchQuery(e.target.value)
                setShowSearchResults(true)
              }}
              onFocus={() => setShowSearchResults(true)}
              className="lj-input"
            />
            {showSearchResults && searchResults.length > 0 && (
              <div className="absolute top-full left-0 right-0 bg-white border border-gray-100 shadow-lg z-50 divide-y divide-gray-50">
                {searchResults.map((store) => (
                  <button
                    key={store.id}
                    type="button"
                    onClick={() => handleSelectStore(store)}
                    className="w-full px-4 py-3 text-left hover:bg-gray-50 flex flex-col transition-colors"
                  >
                    <span className="text-[12px] font-black uppercase tracking-wider">{store.name}</span>
                    <span className="text-[9px] text-gray-400 font-bold uppercase tracking-wider mt-0.5">{store.code} - {store.city}</span>
                  </button>
                ))}
              </div>
            )}
          </div>
        </div>

        {/* Cliente Info Section */}
        <div className="space-y-6">
          <h2 className="text-[11px] font-black uppercase tracking-[0.25em] text-black border-b border-black pb-2">Dati Cliente</h2>
          
          <div className="grid grid-cols-2 gap-4">
            <div>
              <label className="block text-[9px] font-black uppercase tracking-[0.2em] text-black mb-1">Codice Cliente *</label>
              <input
                type="text"
                required
                value={customerCode}
                onChange={(e) => setCustomerCode(e.target.value)}
                className="lj-input"
              />
            </div>
            <div>
              <label className="block text-[9px] font-black uppercase tracking-[0.2em] text-black mb-1">Ragione Sociale *</label>
              <input
                type="text"
                required
                value={customerBusinessName}
                onChange={(e) => setCustomerBusinessName(e.target.value)}
                className="lj-input"
              />
            </div>
          </div>

          <div>
            <label className="block text-[9px] font-black uppercase tracking-[0.2em] text-black mb-1">Insegna / Nome Negozio *</label>
            <input
              type="text"
              required
              value={customerStoreName}
              onChange={(e) => setCustomerStoreName(e.target.value)}
              className="lj-input"
            />
          </div>

          <div>
            <label className="block text-[9px] font-black uppercase tracking-[0.2em] text-black mb-1">Indirizzo Cliente *</label>
            <input
              type="text"
              required
              value={customerAddress}
              onChange={(e) => setCustomerAddress(e.target.value)}
              placeholder="Via, civico"
              className="lj-input"
            />
          </div>

          <div className="grid grid-cols-3 gap-4">
            <div className="col-span-1">
              <label className="block text-[9px] font-black uppercase tracking-[0.2em] text-black mb-1">CAP *</label>
              <input
                type="text"
                required
                value={customerCap}
                onChange={(e) => setCustomerCap(e.target.value)}
                className="lj-input"
              />
            </div>
            <div className="col-span-2">
              <label className="block text-[9px] font-black uppercase tracking-[0.2em] text-black mb-1">Città *</label>
              <input
                type="text"
                required
                value={customerCity}
                onChange={(e) => setCustomerCity(e.target.value)}
                className="lj-input"
              />
            </div>
          </div>

          <div className="grid grid-cols-2 gap-4">
            <div>
              <label className="block text-[9px] font-black uppercase tracking-[0.2em] text-black mb-1">E-mail *</label>
              <input
                type="email"
                required
                value={customerEmail}
                onChange={(e) => setCustomerEmail(e.target.value)}
                className="lj-input"
              />
            </div>
            <div>
              <label className="block text-[9px] font-black uppercase tracking-[0.2em] text-black mb-1">Telefono *</label>
              <input
                type="tel"
                required
                value={customerPhone}
                onChange={(e) => setCustomerPhone(e.target.value)}
                className="lj-input"
              />
            </div>
          </div>

          <div>
            <label className="block text-[9px] font-black uppercase tracking-[0.2em] text-black mb-1.5">Stima Annuale Sell In *</label>
            <select
              value={annualSellInEstimate}
              onChange={(e) => setAnnualSellInEstimate(e.target.value)}
              className="w-full bg-white border border-gray-200 py-3 px-4 text-xs font-bold uppercase tracking-wider rounded-none outline-none focus:border-black"
            >
              <option value="< 2.000">&lt; 2.000</option>
              <option value="2.000 - 4.000">2.000 - 4.000</option>
              <option value="4.000 - 5.000">4.000 - 5.000</option>
              <option value="5.000 - 7.000">5.000 - 7.000</option>
              <option value="7.000 - 10.000">7.000 - 10.000</option>
              <option value="> 10.000">&gt; 10.000</option>
            </select>
          </div>
        </div>

        {/* Customizations Section */}
        <div className="space-y-8">
          <h2 className="text-[11px] font-black uppercase tracking-[0.25em] text-black border-b border-black pb-2">Misure e Dettagli Allestimento</h2>

          {/* Customization 1 */}
          <div className="lj-card p-6 space-y-6">
            <h3 className="text-xs font-black uppercase tracking-wider flex items-center justify-between">
              <span>Personalizzazione 1 *</span>
            </h3>

            {/* Subject */}
            <div>
              <label className="block text-[9px] font-black uppercase tracking-[0.2em] text-black mb-1">Soggetto Personalizzazione *</label>
              <div className="flex gap-4 mb-2">
                {['Nuove collezioni', 'Solo logo Nanan Bijou', 'Altro:'].map((opt) => (
                  <label key={opt} className="flex items-center gap-1.5 text-xs font-bold uppercase tracking-wider cursor-pointer">
                    <input
                      type="radio"
                      name="cust1_subject_opt"
                      value={opt}
                      checked={cust1Subject === opt}
                      onChange={() => setCust1Subject(opt)}
                      className="accent-black"
                    />
                    <span>{opt}</span>
                  </label>
                ))}
              </div>
              {cust1Subject === 'Altro:' && (
                <input
                  type="text"
                  placeholder="Specifica soggetto..."
                  value={cust1SubjectOther}
                  onChange={(e) => setCust1SubjectOther(e.target.value)}
                  className="lj-input mt-2"
                />
              )}
            </div>

            {/* Type */}
            <div>
              <label className="block text-[9px] font-black uppercase tracking-[0.2em] text-black mb-1">Tipologia di Personalizzazione *</label>
              <div className="grid grid-cols-2 gap-2 mb-2">
                {['Affissione', 'Vetrina esterna', 'Vetrina interna', 'Altro:'].map((opt) => (
                  <label key={opt} className="flex items-center gap-1.5 text-xs font-bold uppercase tracking-wider cursor-pointer">
                    <input
                      type="radio"
                      name="cust1_type_opt"
                      value={opt}
                      checked={cust1Type === opt}
                      onChange={() => setCust1Type(opt)}
                      className="accent-black"
                    />
                    <span>{opt}</span>
                  </label>
                ))}
              </div>
              {cust1Type === 'Altro:' && (
                <input
                  type="text"
                  placeholder="Specifica tipologia..."
                  value={cust1TypeOther}
                  onChange={(e) => setCust1TypeOther(e.target.value)}
                  className="lj-input mt-2"
                />
              )}
            </div>

            {/* Dimensions */}
            <div className="grid grid-cols-2 gap-4">
              <div>
                <label className="block text-[9px] font-black uppercase tracking-[0.2em] text-black mb-1">Larghezza in cm *</label>
                <input
                  type="number"
                  required
                  value={cust1Width}
                  onChange={(e) => setCust1Width(e.target.value)}
                  className="lj-input"
                />
              </div>
              <div>
                <label className="block text-[9px] font-black uppercase tracking-[0.2em] text-black mb-1">Altezza in cm *</label>
                <input
                  type="number"
                  required
                  value={cust1Height}
                  onChange={(e) => setCust1Height(e.target.value)}
                  className="lj-input"
                />
              </div>
            </div>

            {/* Material */}
            <div>
              <label className="block text-[9px] font-black uppercase tracking-[0.2em] text-black mb-1">Tipologia di Materiale *</label>
              <div className="space-y-2 mb-2">
                {[
                  'Pannello in forex',
                  'Vetrofania - Pellicola adesiva',
                  'Vetrofania - Pellicola micoforato',
                  'Altro:'
                ].map((opt) => (
                  <label key={opt} className="flex items-center gap-1.5 text-xs font-bold uppercase tracking-wider cursor-pointer">
                    <input
                      type="radio"
                      name="cust1_material_opt"
                      value={opt}
                      checked={cust1Material === opt}
                      onChange={() => setCust1Material(opt)}
                      className="accent-black"
                    />
                    <span>{opt}</span>
                  </label>
                ))}
              </div>
              {cust1Material === 'Altro:' && (
                <input
                  type="text"
                  placeholder="Specifica materiale..."
                  value={cust1MaterialOther}
                  onChange={(e) => setCust1MaterialOther(e.target.value)}
                  className="lj-input mt-2"
                />
              )}
            </div>
          </div>

          {/* Customization 2 */}
          {customizationCount >= 2 && (
            <div className="lj-card p-6 space-y-6 animate-slide-up">
              <div className="flex items-center justify-between border-b border-gray-100 pb-3">
                <h3 className="text-xs font-black uppercase tracking-wider">Personalizzazione 2</h3>
                <button
                  type="button"
                  onClick={() => setCustomizationCount(1)}
                  className="text-danger p-1"
                >
                  <Trash2 size={16} />
                </button>
              </div>

              <div>
                <label className="block text-[9px] font-black uppercase tracking-[0.2em] text-black mb-1">Soggetto Personalizzazione</label>
                <div className="flex gap-4 mb-2">
                  {['Nuove collezioni', 'Solo logo Nanan Bijou', 'Altro:'].map((opt) => (
                    <label key={opt} className="flex items-center gap-1.5 text-xs font-bold uppercase tracking-wider cursor-pointer">
                      <input
                        type="radio"
                        name="cust2_subject_opt"
                        value={opt}
                        checked={cust2Subject === opt}
                        onChange={() => setCust2Subject(opt)}
                        className="accent-black"
                      />
                      <span>{opt}</span>
                    </label>
                  ))}
                </div>
                {cust2Subject === 'Altro:' && (
                  <input
                    type="text"
                    placeholder="Specifica soggetto..."
                    value={cust2SubjectOther}
                    onChange={(e) => setCust2SubjectOther(e.target.value)}
                    className="lj-input mt-2"
                  />
                )}
              </div>

              <div>
                <label className="block text-[9px] font-black uppercase tracking-[0.2em] text-black mb-1">Tipologia di Personalizzazione</label>
                <div className="grid grid-cols-2 gap-2 mb-2">
                  {['Affissione', 'Vetrina esterna', 'Vetrina interna', 'Altro:'].map((opt) => (
                    <label key={opt} className="flex items-center gap-1.5 text-xs font-bold uppercase tracking-wider cursor-pointer">
                      <input
                        type="radio"
                        name="cust2_type_opt"
                        value={opt}
                        checked={cust2Type === opt}
                        onChange={() => setCust2Type(opt)}
                        className="accent-black"
                      />
                      <span>{opt}</span>
                    </label>
                  ))}
                </div>
                {cust2Type === 'Altro:' && (
                  <input
                    type="text"
                    placeholder="Specifica tipologia..."
                    value={cust2TypeOther}
                    onChange={(e) => setCust2TypeOther(e.target.value)}
                    className="lj-input mt-2"
                  />
                )}
              </div>

              <div className="grid grid-cols-2 gap-4">
                <div>
                  <label className="block text-[9px] font-black uppercase tracking-[0.2em] text-black mb-1">Larghezza in cm</label>
                  <input
                    type="number"
                    value={cust2Width}
                    onChange={(e) => setCust2Width(e.target.value)}
                    className="lj-input"
                  />
                </div>
                <div>
                  <label className="block text-[9px] font-black uppercase tracking-[0.2em] text-black mb-1">Altezza in cm</label>
                  <input
                    type="number"
                    value={cust2Height}
                    onChange={(e) => setCust2Height(e.target.value)}
                    className="lj-input"
                  />
                </div>
              </div>

              <div>
                <label className="block text-[9px] font-black uppercase tracking-[0.2em] text-black mb-1">Tipologia di Materiale</label>
                <div className="space-y-2 mb-2">
                  {[
                    'Pannello in forex',
                    'Vetrofania - Pellicola adesiva',
                    'Vetrofania - Pellicola micoforato',
                    'Altro:'
                  ].map((opt) => (
                    <label key={opt} className="flex items-center gap-1.5 text-xs font-bold uppercase tracking-wider cursor-pointer">
                      <input
                        type="radio"
                        name="cust2_material_opt"
                        value={opt}
                        checked={cust2Material === opt}
                        onChange={() => setCust2Material(opt)}
                        className="accent-black"
                      />
                      <span>{opt}</span>
                    </label>
                  ))}
                </div>
                {cust2Material === 'Altro:' && (
                  <input
                    type="text"
                    placeholder="Specifica materiale..."
                    value={cust2MaterialOther}
                    onChange={(e) => setCust2MaterialOther(e.target.value)}
                    className="lj-input mt-2"
                  />
                )}
              </div>
            </div>
          )}

          {/* Customization 3 */}
          {customizationCount >= 3 && (
            <div className="lj-card p-6 space-y-6 animate-slide-up">
              <div className="flex items-center justify-between border-b border-gray-100 pb-3">
                <h3 className="text-xs font-black uppercase tracking-wider">Personalizzazione 3</h3>
                <button
                  type="button"
                  onClick={() => setCustomizationCount(2)}
                  className="text-danger p-1"
                >
                  <Trash2 size={16} />
                </button>
              </div>

              <div>
                <label className="block text-[9px] font-black uppercase tracking-[0.2em] text-black mb-1">Soggetto Personalizzazione</label>
                <div className="flex gap-4 mb-2">
                  {['Nuove collezioni', 'Solo logo Nanan Bijou', 'Altro:'].map((opt) => (
                    <label key={opt} className="flex items-center gap-1.5 text-xs font-bold uppercase tracking-wider cursor-pointer">
                      <input
                        type="radio"
                        name="cust3_subject_opt"
                        value={opt}
                        checked={cust3Subject === opt}
                        onChange={() => setCust3Subject(opt)}
                        className="accent-black"
                      />
                      <span>{opt}</span>
                    </label>
                  ))}
                </div>
                {cust3Subject === 'Altro:' && (
                  <input
                    type="text"
                    placeholder="Specifica soggetto..."
                    value={cust3SubjectOther}
                    onChange={(e) => setCust3SubjectOther(e.target.value)}
                    className="lj-input mt-2"
                  />
                )}
              </div>

              <div>
                <label className="block text-[9px] font-black uppercase tracking-[0.2em] text-black mb-1">Tipologia di Personalizzazione</label>
                <div className="grid grid-cols-2 gap-2 mb-2">
                  {['Affissione', 'Vetrina esterna', 'Vetrina interna', 'Altro:'].map((opt) => (
                    <label key={opt} className="flex items-center gap-1.5 text-xs font-bold uppercase tracking-wider cursor-pointer">
                      <input
                        type="radio"
                        name="cust3_type_opt"
                        value={opt}
                        checked={cust3Type === opt}
                        onChange={() => setCust3Type(opt)}
                        className="accent-black"
                      />
                      <span>{opt}</span>
                    </label>
                  ))}
                </div>
                {cust3Type === 'Altro:' && (
                  <input
                    type="text"
                    placeholder="Specifica tipologia..."
                    value={cust3TypeOther}
                    onChange={(e) => setCust3TypeOther(e.target.value)}
                    className="lj-input mt-2"
                  />
                )}
              </div>

              <div className="grid grid-cols-2 gap-4">
                <div>
                  <label className="block text-[9px] font-black uppercase tracking-[0.2em] text-black mb-1">Larghezza in cm</label>
                  <input
                    type="number"
                    value={cust3Width}
                    onChange={(e) => setCust3Width(e.target.value)}
                    className="lj-input"
                  />
                </div>
                <div>
                  <label className="block text-[9px] font-black uppercase tracking-[0.2em] text-black mb-1">Altezza in cm</label>
                  <input
                    type="number"
                    value={cust3Height}
                    onChange={(e) => setCust3Height(e.target.value)}
                    className="lj-input"
                  />
                </div>
              </div>

              <div>
                <label className="block text-[9px] font-black uppercase tracking-[0.2em] text-black mb-1">Tipologia di Materiale</label>
                <div className="space-y-2 mb-2">
                  {[
                    'Pannello in forex',
                    'Vetrofania - Pellicola adesiva',
                    'Vetrofania - Pellicola micoforato',
                    'Altro:'
                  ].map((opt) => (
                    <label key={opt} className="flex items-center gap-1.5 text-xs font-bold uppercase tracking-wider cursor-pointer">
                      <input
                        type="radio"
                        name="cust3_material_opt"
                        value={opt}
                        checked={cust3Material === opt}
                        onChange={() => setCust3Material(opt)}
                        className="accent-black"
                      />
                      <span>{opt}</span>
                    </label>
                  ))}
                </div>
                {cust3Material === 'Altro:' && (
                  <input
                    type="text"
                    placeholder="Specifica materiale..."
                    value={cust3MaterialOther}
                    onChange={(e) => setCust3MaterialOther(e.target.value)}
                    className="lj-input mt-2"
                  />
                )}
              </div>
            </div>
          )}

          {customizationCount < 3 && (
            <button
              type="button"
              onClick={() => setCustomizationCount((prev) => prev + 1)}
              className="w-full h-12 border border-dashed border-gray-300 flex items-center justify-center gap-2 text-xs font-bold uppercase tracking-wider active:bg-gray-50 transition-colors"
            >
              <Plus size={16} />
              Aggiungi Personalizzazione
            </button>
          )}
        </div>

        {/* Dates Section */}
        <div className="space-y-6">
          <h2 className="text-[11px] font-black uppercase tracking-[0.25em] text-black border-b border-black pb-2">Periodo Esposizione</h2>
          <div className="grid grid-cols-2 gap-4">
            <div>
              <label className="block text-[9px] font-black uppercase tracking-[0.2em] text-black mb-1">Inizio Esposizione</label>
              <input
                type="date"
                value={startDate}
                onChange={(e) => setStartDate(e.target.value)}
                className="lj-input"
              />
            </div>
            <div>
              <label className="block text-[9px] font-black uppercase tracking-[0.2em] text-black mb-1">Fine Esposizione</label>
              <input
                type="date"
                value={endDate}
                onChange={(e) => setEndDate(e.target.value)}
                className="lj-input"
              />
            </div>
          </div>
        </div>

        {/* Printing and Assembly Costs */}
        <div className="space-y-6">
          <h2 className="text-[11px] font-black uppercase tracking-[0.25em] text-black border-b border-black pb-2">Costi Stampa e Montaggio</h2>
          
          <div>
            <label className="block text-[9px] font-black uppercase tracking-[0.2em] text-black mb-1.5">Costi di Stampa *</label>
            <div className="space-y-2">
              {[
                'Stampa a carico del cliente',
                'Stampa a carico di GN Gioielli'
              ].map((opt) => (
                <label key={opt} className="flex items-center gap-1.5 text-xs font-bold uppercase tracking-wider cursor-pointer">
                  <input
                    type="radio"
                    name="printing_cost_opt"
                    value={opt}
                    checked={printingCost === opt}
                    onChange={() => setPrintingCost(opt)}
                    className="accent-black"
                  />
                  <span>{opt}</span>
                </label>
              ))}
            </div>
          </div>

          <div>
            <label className="block text-[9px] font-black uppercase tracking-[0.2em] text-black mb-1.5">Costi Allestimento e Montaggio *</label>
            <div className="space-y-2">
              {[
                'Montaggio a carico del cliente',
                'Montaggio a carico di GN Gioielli'
              ].map((opt) => (
                <label key={opt} className="flex items-center gap-1.5 text-xs font-bold uppercase tracking-wider cursor-pointer">
                  <input
                    type="radio"
                    name="assembly_cost_opt"
                    value={opt}
                    checked={assemblyCost === opt}
                    onChange={() => setAssemblyCost(opt)}
                    className="accent-black"
                  />
                  <span>{opt}</span>
                </label>
              ))}
            </div>
          </div>
        </div>

        {/* Shipping Address */}
        <div className="space-y-6">
          <div className="flex items-center justify-between border-b border-black pb-2">
            <h2 className="text-[11px] font-black uppercase tracking-[0.25em] text-black">Indirizzo di Spedizione</h2>
            <label className="flex items-center gap-1.5 text-[9px] font-black uppercase tracking-wider cursor-pointer">
              <input
                type="checkbox"
                checked={autofillShipping}
                onChange={(e) => setAutofillShipping(e.target.checked)}
                className="accent-black"
              />
              <span>Copia da Dati Cliente</span>
            </label>
          </div>

          <div>
            <label className="block text-[9px] font-black uppercase tracking-[0.2em] text-black mb-1">Indirizzo - Via *</label>
            <input
              type="text"
              required
              value={shippingAddress}
              onChange={(e) => setShippingAddress(e.target.value)}
              className="lj-input"
            />
          </div>

          <div className="grid grid-cols-3 gap-4">
            <div className="col-span-1">
              <label className="block text-[9px] font-black uppercase tracking-[0.2em] text-black mb-1">Civico *</label>
              <input
                type="text"
                required
                value={shippingCivic}
                onChange={(e) => setShippingCivic(e.target.value)}
                className="lj-input"
              />
            </div>
            <div className="col-span-2">
              <label className="block text-[9px] font-black uppercase tracking-[0.2em] text-black mb-1">Città *</label>
              <input
                type="text"
                required
                value={shippingCity}
                onChange={(e) => setShippingCity(e.target.value)}
                className="lj-input"
              />
            </div>
          </div>

          <div className="grid grid-cols-3 gap-4">
            <div className="col-span-1">
              <label className="block text-[9px] font-black uppercase tracking-[0.2em] text-black mb-1">Provincia *</label>
              <input
                type="text"
                required
                maxLength={2}
                placeholder="es. MI"
                value={shippingProvince}
                onChange={(e) => setShippingProvince(e.target.value.toUpperCase())}
                className="lj-input"
              />
            </div>
            <div className="col-span-2">
              <label className="block text-[9px] font-black uppercase tracking-[0.2em] text-black mb-1">CAP *</label>
              <input
                type="text"
                required
                value={shippingCap}
                onChange={(e) => setShippingCap(e.target.value)}
                className="lj-input"
              />
            </div>
          </div>
        </div>

        {/* Photo Upload Section */}
        <div className="space-y-6">
          <h2 className="text-[11px] font-black uppercase tracking-[0.25em] text-black border-b border-black pb-2">Foto Spazi da Personalizzare *</h2>
          
          <div className="flex flex-col items-center justify-center p-8 border-2 border-dashed border-gray-300 bg-gray-50 text-center relative hover:bg-gray-100 transition-colors">
            {photoPreview ? (
              <div className="relative w-full aspect-video">
                <img src={photoPreview} alt="Anteprima spazi" className="w-full h-full object-cover" />
                <label className="absolute bottom-4 right-4 bg-black text-white p-3 shadow-lg cursor-pointer hover:bg-black/90 transition-colors">
                  <Camera size={18} />
                  <input
                    type="file"
                    accept="image/*"
                    capture="environment"
                    onChange={handlePhotoChange}
                    className="hidden"
                  />
                </label>
              </div>
            ) : (
              <label className="flex flex-col items-center gap-3 cursor-pointer w-full py-8">
                <Camera size={32} className="text-gray-400" />
                <div>
                  <span className="text-[11px] font-black uppercase tracking-wider block text-black">Scatta o Carica Foto</span>
                  <span className="text-[9px] text-gray-400 font-bold uppercase tracking-wider mt-1 block">Obbligatorio per procedere</span>
                </div>
                <input
                  type="file"
                  accept="image/*"
                  capture="environment"
                  required
                  onChange={handlePhotoChange}
                  className="hidden"
                />
              </label>
            )}
          </div>
        </div>

        {/* Submit */}
        <div className="pt-6">
          <button
            type="submit"
            disabled={submitting}
            className="w-full lj-button flex items-center justify-center gap-3"
          >
            {submitting ? (
              <>
                <span className="w-4 h-4 border-2 border-white/30 border-t-white rounded-full animate-spin" />
                Salvataggio...
              </>
            ) : (
              'Invia Richiesta'
            )}
          </button>
        </div>
      </form>
    </PageShell>
  )
}
