import { createSlice, PayloadAction } from '@reduxjs/toolkit';

// Paper size definitions (in inches)
export const PAPER_SIZES = {
  '4x6': { width: 4, height: 6 },
  '5x7': { width: 5, height: 7 },
  '8x10': { width: 8, height: 10 },
  '11x14': { width: 11, height: 14 },
  '12x12': { width: 12, height: 12 },
  '8.5x11': { width: 8.5, height: 11 },
  'A4': { width: 8.27, height: 11.69 }
} as const;

// Crop ratio definitions
export const CROP_RATIOS = {
  'original': null,
  // Vertical (portrait) ratios
  '2:3': 2 / 3,     // 0.667 - Classic portrait
  '3:4': 3 / 4,     // 0.750 - Standard photo
  '4:5': 4 / 5,     // 0.800 - Instagram portrait
  '5:7': 5 / 7,     // 0.714 - Medium portrait
  // Square
  '1:1': 1,         // 1.000 - Instagram square
  // Horizontal (landscape) ratios  
  '3:2': 3 / 2,     // 1.500 - Classic 35mm
  '7:5': 7 / 5,     // 1.400 - Medium landscape
  '5:3': 5 / 3,     // 1.667 - Wide landscape
  '16:9': 16 / 9,   // 1.778 - Widescreen
  '2:1': 2 / 1,     // 2.000 - Panoramic
} as const;

export interface BookSpreadState {
  image: {
    src: string | null;
    width: number;
    height: number;
    fileSize: number | null;
    fileName: string | null;
    uploadedPath?: string; // Backend uploaded file path
  } | null;
  paperSize: keyof typeof PAPER_SIZES;
  isSpread: boolean;
  margins: {
    top: number;
    right: number;
    bottom: number;
    left: number;
  };
  orientation: 'portrait' | 'landscape';
  cropRatio: keyof typeof CROP_RATIOS;
  cropToFill: boolean;
  imageScale: number;
  imagePosition: {
    x: number;
    y: number;
  };
  dpi: number;
  gutterMargin: number;
}

const initialState: BookSpreadState = {
  image: null,
  paperSize: '8x10',
  isSpread: false,
  margins: { top: 0.5, right: 0.5, bottom: 0.5, left: 0.5 },
  orientation: 'portrait',
  cropRatio: 'original',
  cropToFill: false,
  imageScale: 1,
  imagePosition: { x: 0, y: 0 },
  dpi: 300,
  gutterMargin: 0,
};

const bookSpreadSlice = createSlice({
  name: 'bookSpread',
  initialState,
  reducers: {
    setImage: (state, action: PayloadAction<BookSpreadState['image']>) => {
      state.image = action.payload;
      // Reset image transformations when new image is loaded
      if (action.payload) {
        state.imageScale = 1;
        state.imagePosition = { x: 0, y: 0 };
      }
    },
    setPaperSize: (state, action: PayloadAction<keyof typeof PAPER_SIZES>) => {
      state.paperSize = action.payload;
    },
    setIsSpread: (state, action: PayloadAction<boolean>) => {
      state.isSpread = action.payload;
    },
    setMargins: (state, action: PayloadAction<Partial<BookSpreadState['margins']>>) => {
      state.margins = { ...state.margins, ...action.payload };
    },
    setOrientation: (state, action: PayloadAction<'portrait' | 'landscape'>) => {
      state.orientation = action.payload;
    },
    setCropRatio: (state, action: PayloadAction<keyof typeof CROP_RATIOS>) => {
      state.cropRatio = action.payload;
    },
    setCropToFill: (state, action: PayloadAction<boolean>) => {
      state.cropToFill = action.payload;
    },
    setImageScale: (state, action: PayloadAction<number>) => {
      state.imageScale = action.payload;
    },
    setImagePosition: (state, action: PayloadAction<{ x?: number; y?: number }>) => {
      state.imagePosition = { ...state.imagePosition, ...action.payload };
    },
    setDpi: (state, action: PayloadAction<number>) => {
      state.dpi = action.payload;
    },
    setGutterMargin: (state, action: PayloadAction<number>) => {
      state.gutterMargin = action.payload;
    },
  },
});

export const {
  setImage,
  setPaperSize,
  setIsSpread,
  setMargins,
  setOrientation,
  setCropRatio,
  setCropToFill,
  setImageScale,
  setImagePosition,
  setDpi,
  setGutterMargin,
} = bookSpreadSlice.actions;

export default bookSpreadSlice.reducer;
