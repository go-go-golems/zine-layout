import { createAsyncThunk, createSelector, createSlice, type PayloadAction } from "@reduxjs/toolkit";
import type {
  ImageLayoutComputation,
  ImageLayoutRequest,
  ImageLayoutTemplate,
} from "../api";
import { api } from "../api";
import type { RootState } from "../store";

export const PAPER_SIZES = {
  Letter: { width: 8.5, height: 11 },
  "8x10": { width: 8, height: 10 },
  "5x7": { width: 5, height: 7 },
  "4x6": { width: 4, height: 6 },
  A4: { width: 8.27, height: 11.69 },
  "Square 8x8": { width: 8, height: 8 },
  Custom: { width: 8, height: 10 },
} as const;

export const ASPECT_RATIOS = {
  None: null,
  "1:1 (Square)": 1,
  "2:3 (Portrait)": 2 / 3,
  "3:2 (Landscape)": 3 / 2,
  "4:5 (Portrait)": 4 / 5,
  "16:9 (Widescreen)": 16 / 9,
  "9:16 (Story)": 9 / 16,
} as const;

export type PaperSizeKey = keyof typeof PAPER_SIZES;
export type AspectRatioKey = keyof typeof ASPECT_RATIOS;

type Orientation = "portrait" | "landscape";
type FillMode = "contain" | "cover";

export interface FrameState {
  paperSize: PaperSizeKey;
  paperWidth: number;
  paperHeight: number;
  dpi: number;
  orientation: Orientation;
  uniformMargins: boolean;
  marginAll: number;
  margins: { top: number; right: number; bottom: number; left: number };
  fillMode: FillMode;
  aspectRatio: AspectRatioKey;
}

export interface CropState {
  strategy: "auto" | "anchor" | "focus" | "manual";
  ratio: number | null;
  zoom: number;
  panX: number;
  panY: number;
  anchorPreset: string;
}

export interface PresentationState {
  userScale: number;
  offsetX: number;
  offsetY: number;
  clampToCanvas: boolean;
}

export interface PreviewState {
  assetId: string;
  result: ImageLayoutComputation | null;
  error: string | null;
  isLoading: boolean;
  isRenderLoading: boolean;
  renderUrl: string | null;
  renderError: string | null;
  compareOpen: boolean;
}

type EditorMode = "idle" | "create" | "edit";

export interface MetaState {
  mode: EditorMode;
  templateName: string;
  templateDescription: string;
  isGlobal: boolean;
  editingTemplateId: string | null;
}

export interface ImageLayoutsEditorState {
  meta: MetaState;
  frame: FrameState;
  crop: CropState;
  presentation: PresentationState;
  preview: PreviewState;
}

const initialFrameState = (): FrameState => ({
  paperSize: "8x10",
  paperWidth: 8,
  paperHeight: 10,
  dpi: 300,
  orientation: "portrait",
  uniformMargins: true,
  marginAll: 0.5,
  margins: { top: 0.5, right: 0.5, bottom: 0.5, left: 0.5 },
  fillMode: "cover",
  aspectRatio: "None",
});

const initialCropState = (): CropState => ({
  strategy: "auto",
  ratio: null,
  zoom: 1,
  panX: 0,
  panY: 0,
  anchorPreset: "middle-center",
});

const initialPresentationState = (): PresentationState => ({
  userScale: 1,
  offsetX: 0,
  offsetY: 0,
  clampToCanvas: false,
});

const initialPreviewState = (): PreviewState => ({
  assetId: "",
  result: null,
  error: null,
  isLoading: false,
  isRenderLoading: false,
  renderUrl: null,
  renderError: null,
  compareOpen: false,
});

const initialMetaState = (): MetaState => ({
  mode: "idle",
  templateName: "",
  templateDescription: "",
  isGlobal: false,
  editingTemplateId: null,
});

const initialState = (): ImageLayoutsEditorState => ({
  meta: initialMetaState(),
  frame: initialFrameState(),
  crop: initialCropState(),
  presentation: initialPresentationState(),
  preview: initialPreviewState(),
});

const aspectRatioFromValue = (value: number | null): AspectRatioKey => {
  if (value === null) return "None";
  const match = Object.entries(ASPECT_RATIOS).find(
    ([, ratioValue]) => ratioValue !== null && Math.abs(ratioValue - value) < 0.01,
  );
  return (match?.[0] as AspectRatioKey | undefined) ?? "None";
};

const buildLayoutFromState = (state: ImageLayoutsEditorState): ImageLayoutRequest => {
  const ratio = ASPECT_RATIOS[state.frame.aspectRatio];
  return {
    frame: {
      mode: "page",
      fill: state.frame.fillMode,
      ratio: ratio ?? undefined,
      page: {
        width_in: state.frame.paperWidth,
        height_in: state.frame.paperHeight,
        dpi: state.frame.dpi,
        orientation: state.frame.orientation,
        margins_in: state.frame.margins,
      },
    },
    crop: {
      strategy: state.crop.strategy,
      ratio: state.crop.ratio ?? undefined,
      zoom: state.crop.zoom,
      anchor: state.crop.anchorPreset,
      pan: { x: state.crop.panX, y: state.crop.panY },
      units: "normalized",
      focus: null,
    },
    presentation: {
      user_scale: state.presentation.userScale,
      offset_px: { x: state.presentation.offsetX, y: state.presentation.offsetY },
      clamp_to_canvas: state.presentation.clampToCanvas,
    },
    export: {
      format: "png",
      quality: 90,
      background: "white",
      filename_template: "{name}-{panel}.{ext}",
      out_dir: "./out",
    },
  };
};

export const selectImageLayoutsEditor = (root: RootState) => root.imageLayoutsEditor;

export const selectCurrentLayout = createSelector(
  selectImageLayoutsEditor,
  (editor) => buildLayoutFromState(editor),
);

export const selectPreviewAssetId = (root: RootState) =>
  root.imageLayoutsEditor.preview.assetId;

export const selectPreviewState = (root: RootState) => root.imageLayoutsEditor.preview;

export const previewLayoutThunk = createAsyncThunk<
  ImageLayoutComputation,
  { projectId: string },
  { state: RootState; rejectValue: string }
>("imageLayoutsEditor/preview", async ({ projectId }, thunkAPI) => {
  const state = thunkAPI.getState();
  const editor = state.imageLayoutsEditor;
  const assetId = editor.preview.assetId;

  if (!assetId) {
    return thunkAPI.rejectWithValue("Select an asset to preview");
  }
  try {
    const layout = selectCurrentLayout(state);
    const res = await thunkAPI
      .dispatch(
        api.endpoints.previewLayoutRequest.initiate({
          projectId,
          layout,
          assetId,
        }),
      )
      .unwrap();
    return res;
  } catch (err: any) {
    const msg = err?.data?.error ?? err?.error ?? err?.message ?? "Preview failed";
    return thunkAPI.rejectWithValue(typeof msg === "string" ? msg : "Preview failed");
  }
});

export const renderLayoutThunk = createAsyncThunk<
  string,
  { projectId: string },
  { state: RootState; rejectValue: string }
>("imageLayoutsEditor/render", async ({ projectId }, thunkAPI) => {
  const state = thunkAPI.getState();
  const editor = state.imageLayoutsEditor;
  const assetId = editor.preview.assetId;

  if (!assetId) {
    return thunkAPI.rejectWithValue("Select an asset to render");
  }
  try {
    const layout = selectCurrentLayout(state);
    const blob = await thunkAPI
      .dispatch(
        api.endpoints.renderLayoutRequest.initiate({
          projectId,
          layout,
          assetId,
        }),
      )
      .unwrap();
    return URL.createObjectURL(blob);
  } catch (err: any) {
    const msg = err?.data?.error ?? err?.error ?? err?.message ?? "Render failed";
    return thunkAPI.rejectWithValue(typeof msg === "string" ? msg : "Render failed");
  }
});

const imageLayoutsEditorSlice = createSlice({
  name: "imageLayoutsEditor",
  initialState: initialState(),
  reducers: {
    resetForm(state) {
      const preservedAssetId = state.preview.assetId;
      const preservedCompare = state.preview.compareOpen;
      return {
        ...initialState(),
        preview: {
          ...initialPreviewState(),
          assetId: preservedAssetId,
          compareOpen: preservedCompare,
        },
        meta: {
          ...initialMetaState(),
          mode: "idle",
        },
      };
    },
    startCreate(state) {
      const preservedAssetId = state.preview.assetId;
      return {
        ...initialState(),
        meta: {
          ...initialMetaState(),
          mode: "create",
        },
        preview: {
          ...initialPreviewState(),
          assetId: preservedAssetId,
        },
      };
    },
    startEdit(state, action: PayloadAction<ImageLayoutTemplate>) {
      const preservedAssetId = state.preview.assetId;
      const preservedCompare = state.preview.compareOpen;
      const next = initialState();
      next.meta.mode = "edit";
      next.meta.templateName = action.payload.name ?? "";
      next.meta.templateDescription = action.payload.description ?? "";
      next.meta.isGlobal = action.payload.scope === "global";
      next.meta.editingTemplateId = action.payload.id;
      next.preview.assetId = preservedAssetId;
      next.preview.compareOpen = preservedCompare;

      const maybeLayout = action.payload.settings as any;
      const frame = maybeLayout?.frame ?? null;
      const crop = maybeLayout?.crop ?? null;
      const presentation = maybeLayout?.presentation ?? null;

      if (frame && crop && presentation) {
        if (frame.mode === "page" && frame.page) {
          next.frame.paperWidth = frame.page.width_in ?? next.frame.paperWidth;
          next.frame.paperHeight = frame.page.height_in ?? next.frame.paperHeight;
          next.frame.dpi = frame.page.dpi ?? next.frame.dpi;
          next.frame.orientation = frame.page.orientation ?? next.frame.orientation;
          const m = frame.page.margins_in ?? next.frame.margins;
          next.frame.margins = {
            top: m.top ?? next.frame.margins.top,
            right: m.right ?? next.frame.margins.right,
            bottom: m.bottom ?? next.frame.margins.bottom,
            left: m.left ?? next.frame.margins.left,
          };
          const allEqual =
            next.frame.margins.top === next.frame.margins.right &&
            next.frame.margins.top === next.frame.margins.bottom &&
            next.frame.margins.top === next.frame.margins.left;
          next.frame.uniformMargins = allEqual;
          if (allEqual) {
            next.frame.marginAll = next.frame.margins.top;
          }
        }
        next.frame.fillMode = frame.fill ?? next.frame.fillMode;
        next.frame.aspectRatio = aspectRatioFromValue(frame.ratio ?? null);

        next.crop.strategy = crop.strategy ?? next.crop.strategy;
        next.crop.ratio = crop.ratio ?? null;
        next.crop.zoom = crop.zoom ?? next.crop.zoom;
        next.crop.panX = crop.pan?.x ?? next.crop.panX;
        next.crop.panY = crop.pan?.y ?? next.crop.panY;
        next.crop.anchorPreset = crop.anchor ?? next.crop.anchorPreset;

        next.presentation.userScale = presentation.user_scale ?? next.presentation.userScale;
        next.presentation.offsetX = presentation.offset_px?.x ?? next.presentation.offsetX;
        next.presentation.offsetY = presentation.offset_px?.y ?? next.presentation.offsetY;
        next.presentation.clampToCanvas =
          presentation.clamp_to_canvas ?? next.presentation.clampToCanvas;

        return next;
      }

      const legacy = action.payload.settings as any;
      next.frame.paperWidth = legacy.paper_width_in ?? next.frame.paperWidth;
      next.frame.paperHeight = legacy.paper_height_in ?? next.frame.paperHeight;
      next.frame.dpi = legacy.dpi ?? next.frame.dpi;
      next.frame.orientation = legacy.orientation ?? next.frame.orientation;
      next.frame.margins = {
        top: legacy.margin_top_in ?? next.frame.margins.top,
        right: legacy.margin_right_in ?? next.frame.margins.right,
        bottom: legacy.margin_bottom_in ?? next.frame.margins.bottom,
        left: legacy.margin_left_in ?? next.frame.margins.left,
      };
      const allEqual =
        next.frame.margins.top === next.frame.margins.right &&
        next.frame.margins.top === next.frame.margins.bottom &&
        next.frame.margins.top === next.frame.margins.left;
      next.frame.uniformMargins = allEqual;
      if (allEqual) {
        next.frame.marginAll = next.frame.margins.top;
      }
      next.frame.fillMode = legacy.crop_to_fill ? "cover" : "contain";
      next.frame.aspectRatio = aspectRatioFromValue(legacy.crop_ratio ?? null);

      next.crop.strategy = "auto";
      next.crop.ratio = legacy.crop_ratio ?? null;
      next.crop.zoom = 1;
      next.crop.panX = legacy.position_x ?? next.crop.panX;
      next.crop.panY = legacy.position_y ?? next.crop.panY;
      next.crop.anchorPreset = legacy.anchor_preset ?? next.crop.anchorPreset;

      next.presentation.userScale = legacy.user_scale ?? next.presentation.userScale;
      next.presentation.offsetX = 0;
      next.presentation.offsetY = 0;
      next.presentation.clampToCanvas = false;
      return next;
    },
    setTemplateName(state, action: PayloadAction<string>) {
      state.meta.templateName = action.payload;
    },
    setTemplateDescription(state, action: PayloadAction<string>) {
      state.meta.templateDescription = action.payload;
    },
    setIsGlobal(state, action: PayloadAction<boolean>) {
      state.meta.isGlobal = action.payload;
    },
    setPaperSize(state, action: PayloadAction<PaperSizeKey>) {
      state.frame.paperSize = action.payload;
      if (action.payload !== "Custom") {
        state.frame.paperWidth = PAPER_SIZES[action.payload].width;
        state.frame.paperHeight = PAPER_SIZES[action.payload].height;
      }
    },
    setPaperWidth(state, action: PayloadAction<number>) {
      state.frame.paperWidth = action.payload;
    },
    setPaperHeight(state, action: PayloadAction<number>) {
      state.frame.paperHeight = action.payload;
    },
    setDpi(state, action: PayloadAction<number>) {
      state.frame.dpi = action.payload;
    },
    setOrientation(state, action: PayloadAction<Orientation>) {
      state.frame.orientation = action.payload;
    },
    setUniformMargins(state, action: PayloadAction<boolean>) {
      state.frame.uniformMargins = action.payload;
      if (action.payload) {
        const current = state.frame.margins.top;
        state.frame.marginAll = current;
      }
    },
    setMarginAll(state, action: PayloadAction<number>) {
      state.frame.marginAll = action.payload;
      state.frame.margins = {
        top: action.payload,
        right: action.payload,
        bottom: action.payload,
        left: action.payload,
      };
    },
    setMarginTop(state, action: PayloadAction<number>) {
      state.frame.margins.top = action.payload;
    },
    setMarginRight(state, action: PayloadAction<number>) {
      state.frame.margins.right = action.payload;
    },
    setMarginBottom(state, action: PayloadAction<number>) {
      state.frame.margins.bottom = action.payload;
    },
    setMarginLeft(state, action: PayloadAction<number>) {
      state.frame.margins.left = action.payload;
    },
    setFillMode(state, action: PayloadAction<FillMode>) {
      state.frame.fillMode = action.payload;
    },
    setAspectRatio(state, action: PayloadAction<AspectRatioKey>) {
      state.frame.aspectRatio = action.payload;
    },
    setCropStrategy(state, action: PayloadAction<CropState["strategy"]>) {
      state.crop.strategy = action.payload;
    },
    setCropRatio(state, action: PayloadAction<number | null>) {
      state.crop.ratio = action.payload;
    },
    setZoom(state, action: PayloadAction<number>) {
      state.crop.zoom = action.payload;
    },
    setPanX(state, action: PayloadAction<number>) {
      state.crop.panX = action.payload;
    },
    setPanY(state, action: PayloadAction<number>) {
      state.crop.panY = action.payload;
    },
    setAnchorPreset(state, action: PayloadAction<string>) {
      state.crop.anchorPreset = action.payload;
    },
    setUserScale(state, action: PayloadAction<number>) {
      state.presentation.userScale = action.payload;
    },
    setOffsetX(state, action: PayloadAction<number>) {
      state.presentation.offsetX = action.payload;
    },
    setOffsetY(state, action: PayloadAction<number>) {
      state.presentation.offsetY = action.payload;
    },
    setClampToCanvas(state, action: PayloadAction<boolean>) {
      state.presentation.clampToCanvas = action.payload;
    },
    setPreviewAssetId(state, action: PayloadAction<string>) {
      state.preview.assetId = action.payload;
    },
    setCompareOpen(state, action: PayloadAction<boolean>) {
      state.preview.compareOpen = action.payload;
    },
    setPreviewError(state, action: PayloadAction<string | null>) {
      state.preview.error = action.payload;
      state.preview.result = action.payload ? null : state.preview.result;
    },
    clearRender(state) {
      if (state.preview.renderUrl) {
        URL.revokeObjectURL(state.preview.renderUrl);
      }
      state.preview.renderUrl = null;
      state.preview.renderError = null;
    },
  },
  extraReducers: (builder) => {
    builder
      .addCase(previewLayoutThunk.pending, (state) => {
        state.preview.isLoading = true;
        state.preview.error = null;
      })
      .addCase(previewLayoutThunk.fulfilled, (state, action) => {
        state.preview.isLoading = false;
        state.preview.result = action.payload;
        state.preview.error = null;
      })
      .addCase(previewLayoutThunk.rejected, (state, action) => {
        state.preview.isLoading = false;
        state.preview.result = null;
        state.preview.error =
          (typeof action.payload === "string" && action.payload) ||
          action.error.message ||
          "Preview failed";
      })
      .addCase(renderLayoutThunk.pending, (state) => {
        state.preview.isRenderLoading = true;
        state.preview.renderError = null;
      })
      .addCase(renderLayoutThunk.fulfilled, (state, action) => {
        if (state.preview.renderUrl) {
          URL.revokeObjectURL(state.preview.renderUrl);
        }
        state.preview.isRenderLoading = false;
        state.preview.renderUrl = action.payload;
      })
      .addCase(renderLayoutThunk.rejected, (state, action) => {
        state.preview.isRenderLoading = false;
        state.preview.renderError =
          (typeof action.payload === "string" && action.payload) ||
          action.error.message ||
          "Render failed";
      });
  },
});

export const {
  resetForm,
  startCreate,
  startEdit,
  setTemplateName,
  setTemplateDescription,
  setIsGlobal,
  setPaperSize,
  setPaperWidth,
  setPaperHeight,
  setDpi,
  setOrientation,
  setUniformMargins,
  setMarginAll,
  setMarginTop,
  setMarginRight,
  setMarginBottom,
  setMarginLeft,
  setFillMode,
  setAspectRatio,
  setCropStrategy,
  setCropRatio,
  setZoom,
  setPanX,
  setPanY,
  setAnchorPreset,
  setUserScale,
  setOffsetX,
  setOffsetY,
  setClampToCanvas,
  setPreviewAssetId,
  setCompareOpen,
  setPreviewError,
  clearRender,
} = imageLayoutsEditorSlice.actions;

export const imageLayoutsEditorReducer = imageLayoutsEditorSlice.reducer;
