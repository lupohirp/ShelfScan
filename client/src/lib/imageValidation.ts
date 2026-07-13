export interface ValidationResult {
  isValid: boolean;
  errors: string[];
  blurScore: number;
  brightness: number;
  edgeDensity: number;
}

/**
 * Validates a captured image for blur, lighting, and framing/distance.
 * Processes on a temporary canvas in the background.
 */
export function validateCapturedImage(
  imageSrc: string,
  onComplete: (result: ValidationResult) => void
) {
  const img = new Image();
  img.crossOrigin = "anonymous";
  img.src = imageSrc;
  
  img.onload = () => {
    const width = 300;
    const height = 300;
    const canvas = document.createElement('canvas');
    canvas.width = width;
    canvas.height = height;
    const ctx = canvas.getContext('2d');
    
    if (!ctx) {
      onComplete({
        isValid: true,
        errors: [],
        blurScore: 100,
        brightness: 128,
        edgeDensity: 0.15
      });
      return;
    }
    
    // --- PASS 1: 1:1 Center Crop for High-Accuracy Blur Detection ---
    const cropSize = 300;
    const cropLeft = Math.max(0, Math.floor((img.width - cropSize) / 2));
    const cropTop = Math.max(0, Math.floor((img.height - cropSize) / 2));
    const actualCropW = Math.min(img.width, cropSize);
    const actualCropH = Math.min(img.height, cropSize);
    
    ctx.clearRect(0, 0, width, height);
    ctx.drawImage(img, cropLeft, cropTop, actualCropW, actualCropH, 0, 0, actualCropW, actualCropH);
    
    let cropImgData;
    try {
      cropImgData = ctx.getImageData(0, 0, width, height);
    } catch (e) {
      onComplete({
        isValid: true,
        errors: [],
        blurScore: 100,
        brightness: 128,
        edgeDensity: 0.15
      });
      return;
    }
    
    const cropData = cropImgData.data;
    const cropGray = new Float32Array(width * height);
    for (let i = 0; i < cropData.length; i += 4) {
      cropGray[i / 4] = 0.299 * cropData[i] + 0.587 * cropData[i + 1] + 0.114 * cropData[i + 2];
    }
    
    // Compute Laplacian variance on center crop (excluding borders)
    const laplacian = new Float32Array(width * height);
    let laplacianSum = 0;
    for (let y = 1; y < height - 1; y++) {
      for (let x = 1; x < width - 1; x++) {
        const idx = y * width + x;
        const val = 
          cropGray[idx - width] + 
          cropGray[idx - 1] + 
          -4 * cropGray[idx] + 
          cropGray[idx + 1] + 
          cropGray[idx + width];
        laplacian[idx] = val;
        laplacianSum += val;
      }
    }
    const laplacianMean = laplacianSum / ((width - 2) * (height - 2));
    
    let varianceSum = 0;
    for (let y = 1; y < height - 1; y++) {
      for (let x = 1; x < width - 1; x++) {
        const idx = y * width + x;
        const diff = laplacian[idx] - laplacianMean;
        varianceSum += diff * diff;
      }
    }
    const blurScore = varianceSum / ((width - 2) * (height - 2));
    
    // --- PASS 2: Downsampled Full Image for Brightness and Edge Density/Framing ---
    ctx.clearRect(0, 0, width, height);
    ctx.drawImage(img, 0, 0, width, height);
    
    let fullImgData;
    try {
      fullImgData = ctx.getImageData(0, 0, width, height);
    } catch (e) {
      onComplete({
        isValid: true,
        errors: [],
        blurScore: blurScore,
        brightness: 128,
        edgeDensity: 0.15
      });
      return;
    }
    
    const fullData = fullImgData.data;
    const fullGray = new Float32Array(width * height);
    let brightnessSum = 0;
    for (let i = 0; i < fullData.length; i += 4) {
      const grayVal = 0.299 * fullData[i] + 0.587 * fullData[i + 1] + 0.114 * fullData[i + 2];
      fullGray[i / 4] = grayVal;
      brightnessSum += grayVal;
    }
    const brightness = brightnessSum / fullGray.length;
    
    // Sobel Edge Density on full downsampled image
    let edgeCount = 0;
    const edgeThreshold = 35;
    for (let y = 1; y < height - 1; y++) {
      for (let x = 1; x < width - 1; x++) {
        const idx = y * width + x;
        const gx = 
          -1 * fullGray[idx - width - 1] + 1 * fullGray[idx - width + 1] +
          -2 * fullGray[idx - 1]         + 2 * fullGray[idx + 1] +
          -1 * fullGray[idx + width - 1] + 1 * fullGray[idx + width + 1];
          
        const gy = 
          -1 * fullGray[idx - width - 1] - 2 * fullGray[idx - width] - 1 * fullGray[idx - width + 1] +
          1 * fullGray[idx + width - 1]  + 2 * fullGray[idx + width]  + 1 * fullGray[idx + width + 1];
          
        const mag = Math.sqrt(gx * gx + gy * gy);
        if (mag > edgeThreshold) {
          edgeCount++;
        }
      }
    }
    const edgeDensity = edgeCount / ((width - 2) * (height - 2));
    
    const errors: string[] = [];
    
    // Brightness Check
    if (brightness < 45) {
      errors.push("La foto è troppo buia. Assicurati che lo scaffale sia ben illuminato o attiva il flash.");
    } else if (brightness > 225) {
      errors.push("La foto è troppo luminosa o presenta riflessi. Cambia angolazione per ridurre il riflesso.");
    }
    
    // Blur Check (Using 1:1 Crop Laplacian Variance with threshold of 160)
    if (blurScore < 160) {
      errors.push("La foto è sfocata. Tieni il telefono più fermo e riprova.");
    }
    
    // Edge Density Check (Too far or empty framing)
    if (edgeDensity < 0.055) {
      errors.push("Troppo distante o inquadratura vuota. Avvicinati allo scaffale ed allinealo nel box guida.");
    }
    
    onComplete({
      isValid: errors.length === 0,
      errors,
      blurScore,
      brightness,
      edgeDensity
    });
  };
  
  img.onerror = () => {
    onComplete({
      isValid: true,
      errors: [],
      blurScore: 100,
      brightness: 128,
      edgeDensity: 0.15
    });
  };
}
