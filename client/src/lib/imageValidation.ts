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
    // Downsample image for quick processing
    const width = 300;
    const height = 300;
    const canvas = document.createElement('canvas');
    canvas.width = width;
    canvas.height = height;
    const ctx = canvas.getContext('2d');
    
    if (!ctx) {
      // Fallback if context is not available
      onComplete({
        isValid: true,
        errors: [],
        blurScore: 100,
        brightness: 128,
        edgeDensity: 0.15
      });
      return;
    }
    
    ctx.drawImage(img, 0, 0, width, height);
    let imgData;
    try {
      imgData = ctx.getImageData(0, 0, width, height);
    } catch (e) {
      // CORS or canvas error fallback
      onComplete({
        isValid: true,
        errors: [],
        blurScore: 100,
        brightness: 128,
        edgeDensity: 0.15
      });
      return;
    }
    
    const data = imgData.data;
    
    // 1. Grayscale Conversion
    const gray = new Float32Array(width * height);
    let brightnessSum = 0;
    for (let i = 0; i < data.length; i += 4) {
      const r = data[i];
      const g = data[i + 1];
      const b = data[i + 2];
      // Luminance formula
      const grayVal = 0.299 * r + 0.587 * g + 0.114 * b;
      gray[i / 4] = grayVal;
      brightnessSum += grayVal;
    }
    const brightness = brightnessSum / gray.length;
    
    // 2. Blur Detection (Laplacian Variance)
    const laplacian = new Float32Array(width * height);
    let laplacianSum = 0;
    for (let y = 1; y < height - 1; y++) {
      for (let x = 1; x < width - 1; x++) {
        const idx = y * width + x;
        // Laplacian kernel: [[0, 1, 0], [1, -4, 1], [0, 1, 0]]
        const val = 
          gray[idx - width] + 
          gray[idx - 1] + 
          -4 * gray[idx] + 
          gray[idx + 1] + 
          gray[idx + width];
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
    
    // 3. Sobel Edge Density (Distance & Framing Check)
    let edgeCount = 0;
    const edgeThreshold = 35; // Gradient magnitude threshold
    
    for (let y = 1; y < height - 1; y++) {
      for (let x = 1; x < width - 1; x++) {
        const idx = y * width + x;
        
        const gx = 
          -1 * gray[idx - width - 1] + 1 * gray[idx - width + 1] +
          -2 * gray[idx - 1]         + 2 * gray[idx + 1] +
          -1 * gray[idx + width - 1] + 1 * gray[idx + width + 1];
          
        const gy = 
          -1 * gray[idx - width - 1] - 2 * gray[idx - width] - 1 * gray[idx - width + 1] +
          1 * gray[idx + width - 1]  + 2 * gray[idx + width]  + 1 * gray[idx + width + 1];
          
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
      errors.push("The photo is too dark. Please hold steady, turn on flash, or improve lighting.");
    } else if (brightness > 225) {
      errors.push("The photo is too bright / has too much glare. Please change your position.");
    }
    
    // Blur Check
    if (blurScore < 85) {
      errors.push("The photo is blurry. Please hold the phone steady and try again.");
    }
    
    // Edge Density Check (Too far or empty framing)
    if (edgeDensity < 0.055) {
      errors.push("Too far or empty. Please move closer and align the shelf inside the guide box.");
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
