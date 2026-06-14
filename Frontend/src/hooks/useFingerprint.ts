import { useState, useEffect, useCallback } from 'react'

export interface FingerprintData {
  // Browser info
  ua: string           // User Agent
  al: string           // Accept Language
  ae: string           // Accept Encoding
  
  // Screen info
  sw: number           // Screen Width
  sh: number           // Screen Height
  cd: number           // Color Depth
  pr: number           // Pixel Ratio
  
  // Timezone
  tz: string           // Timezone
  to: number           // Timezone Offset
  
  // Platform
  pf: string           // Platform
  hc: number           // Hardware Concurrency (CPU cores)
  dm: number           // Device Memory (GB)
  tp: number           // Touch Points
  
  // Canvas fingerprint (hash)
  ch: string
  
  // WebGL fingerprint (hash)
  wh: string
  
  // Audio fingerprint (hash)
  ah: string
  
  // Fonts (hash of installed fonts list)
  fh: string
  
  // Plugins (hash of plugins list)
  ph: string
  
  // WebRTC (IP addresses)
  li: string[]
  
  // Additional browser features
  dnt: string          // Do Not Track
  ce: boolean          // Cookie Enabled
  lg: string           // Language
}

export interface UseFingerprintResult {
  fingerprint: string
  fingerprintData: FingerprintData | null
  loading: boolean
  error: Error | null
  generateFingerprint: () => Promise<void>
}

// Hash function for strings
async function hashString(str: string): Promise<string> {
  const encoder = new TextEncoder()
  const data = encoder.encode(str)
  const hashBuffer = await crypto.subtle.digest('SHA-256', data)
  const hashArray = Array.from(new Uint8Array(hashBuffer))
  const hashHex = hashArray.map(b => b.toString(16).padStart(2, '0')).join('')
  return hashHex.substring(0, 16)
}

// Get canvas fingerprint
async function getCanvasFingerprint(): Promise<string> {
  try {
    const canvas = document.createElement('canvas')
    const ctx = canvas.getContext('2d')
    if (!ctx) return ''
    
    canvas.width = 200
    canvas.height = 50
    
    // Draw unique pattern
    ctx.textBaseline = 'top'
    ctx.font = '14px Arial'
    ctx.fillStyle = '#f60'
    ctx.fillRect(0, 0, 100, 50)
    ctx.fillStyle = '#069'
    ctx.fillText('MentalChat 🌸', 2, 15)
    ctx.fillStyle = 'rgba(102, 204, 0, 0.7)'
    ctx.fillText('MentalChat 🌸', 4, 35)
    
    const dataURL = canvas.toDataURL()
    return await hashString(dataURL)
  } catch (error) {
    console.error('Canvas fingerprint error:', error)
    return ''
  }
}

// Get WebGL fingerprint
async function getWebGLFingerprint(): Promise<string> {
  try {
    const canvas = document.createElement('canvas')
    const gl = canvas.getContext('webgl') || canvas.getContext('experimental-webgl')
    if (!gl) return ''
    
    const webglGl = gl as WebGLRenderingContext
    const debugInfo = webglGl.getExtension('WEBGL_debug_renderer_info')
    if (!debugInfo) return ''
    
    const vendor = webglGl.getParameter(debugInfo.UNMASKED_VENDOR_WEBGL)
    const renderer = webglGl.getParameter(debugInfo.UNMASKED_RENDERER_WEBGL)
    
    return await hashString(`${vendor}|${renderer}`)
  } catch (error) {
    console.error('WebGL fingerprint error:', error)
    return ''
  }
}

// Get installed fonts fingerprint
async function getFontsFingerprint(): Promise<string> {
  try {
    const baseFonts = ['monospace', 'sans-serif', 'serif']
    const testFonts = [
      'Arial', 'Verdana', 'Helvetica', 'Times New Roman', 'Times',
      'Courier New', 'Courier', 'Georgia', 'Palatino', 'Garamond',
      'Bookman', 'Comic Sans MS', 'Trebuchet MS', 'Arial Black',
      'Impact', 'Lucida Sans Unicode', 'Geneva', 'Lucida Console',
      'Tahoma', 'Technora', 'Cantarell', 'DejaVu Sans', 'Liberation Sans'
    ]
    
    const canvas = document.createElement('canvas')
    const ctx = canvas.getContext('2d')
    if (!ctx) return ''
    
    canvas.width = 200
    canvas.height = 50
    
    const detectedFonts: string[] = []
    
    for (const testFont of testFonts) {
      let detected = false
      for (const baseFont of baseFonts) {
        ctx.font = `72px "${testFont}", ${baseFont}`
        const width = ctx.measureText('MMMMMMMMMM').width
        ctx.font = `72px ${baseFont}`
        const baseWidth = ctx.measureText('MMMMMMMMMM').width
        
        if (width !== baseWidth) {
          detected = true
          break
        }
      }
      if (detected) {
        detectedFonts.push(testFont)
      }
    }
    
    return await hashString(detectedFonts.join(','))
  } catch (error) {
    console.error('Fonts fingerprint error:', error)
    return ''
  }
}

// Get plugins fingerprint
async function getPluginsFingerprint(): Promise<string> {
  try {
    const plugins: string[] = []
    for (let i = 0; i < navigator.plugins.length; i++) {
      const plugin = navigator.plugins[i]
      plugins.push(`${plugin.name}|${plugin.description}|${plugin.filename}`)
    }
    return plugins.length > 0 ? await hashString(plugins.join(',')) : 'no_plugins'
  } catch (error) {
    console.error('Plugins fingerprint error:', error)
    return ''
  }
}

// Get local IPs via WebRTC
async function getLocalIPs(): Promise<string[]> {
  const ips: string[] = []
  
  try {
    const pc = new RTCPeerConnection({ iceServers: [] })
    
    pc.createDataChannel('')
    const offer = await pc.createOffer()
    await pc.setLocalDescription(offer)
    
    await new Promise<void>((resolve) => {
      pc.onicecandidate = (event) => {
        if (!event.candidate) {
          resolve()
          return
        }
        
        const candidate = event.candidate.candidate
        if (candidate) {
          const ipMatch = /([0-9]{1,3}(\.[0-9]{1,3}){3})/.exec(candidate)
          if (ipMatch && ipMatch[1] && !ips.includes(ipMatch[1])) {
            ips.push(ipMatch[1])
          }
        }
      }
      
      setTimeout(resolve, 1000)
    })
    
    pc.close()
  } catch (error) {
    console.error('WebRTC IP discovery error:', error)
  }
  
  return ips
}

export function useFingerprint(): UseFingerprintResult {
  const [fingerprint, setFingerprint] = useState('')
  const [fingerprintData, setFingerprintData] = useState<FingerprintData | null>(null)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<Error | null>(null)

  const generateFingerprint = useCallback(async () => {
    setLoading(true)
    setError(null)

    try {
      // Collect basic data
      const data: FingerprintData = {
        ua: navigator.userAgent,
        al: navigator.languages.join(','),
        ae: 'gzip, deflate, br',
        sw: screen.width,
        sh: screen.height,
        cd: screen.colorDepth,
        pr: window.devicePixelRatio,
        tz: Intl.DateTimeFormat().resolvedOptions().timeZone,
        to: new Date().getTimezoneOffset(),
        pf: navigator.platform,
        hc: navigator.hardwareConcurrency || 0,
        dm: (navigator as any).deviceMemory || 0,
        tp: navigator.maxTouchPoints || 0,
        ch: '',
        wh: '',
        ah: '',
        fh: '',
        ph: '',
        li: [],
        dnt: navigator.doNotTrack || 'unspecified',
        ce: navigator.cookieEnabled,
        lg: navigator.language,
      }

      // Get canvas fingerprint
      data.ch = await getCanvasFingerprint()
      
      // Get WebGL fingerprint
      data.wh = await getWebGLFingerprint()
      
      // Get fonts fingerprint
      data.fh = await getFontsFingerprint()
      
      // Get plugins fingerprint
      try {
        data.ph = await getPluginsFingerprint()
      } catch (e) {
        data.ph = 'unavailable'
      }
      
      // Get local IPs
      data.li = await getLocalIPs()

      // Generate combined fingerprint
      const components = [
        data.ua,
        data.al,
        `${data.sw}x${data.sh}`,
        `${data.cd}-${data.pr}-${data.dm}`,
        data.tz,
        `${data.to}-${data.hc}-${data.tp}`,
        data.pf,
        data.ch,
        data.wh,
        data.fh,
        data.ph,
        data.li.join(','),
        data.dnt,
        String(data.ce),
        data.lg,
      ].join('|')

      const fp = await hashString(components)
      
      setFingerprint(fp)
      setFingerprintData(data)
      
      console.log('Fingerprint generated:', fp.substring(0, 16) + '...')
    } catch (err) {
      console.error('Fingerprint generation error:', err)
      setError(err instanceof Error ? err : new Error('Failed to generate fingerprint'))
    } finally {
      setLoading(false)
    }
  }, [])

  // Auto-generate on mount
  useEffect(() => {
    generateFingerprint()
  }, [generateFingerprint])

  return {
    fingerprint,
    fingerprintData,
    loading,
    error,
    generateFingerprint,
  }
}

export default useFingerprint
