import React, { useEffect, useMemo, useRef, useState } from "react";
import {
  useGetImageLayoutTemplatesQuery,
  useCreateImageLayoutTemplateMutation,
  useCreateGlobalImageLayoutTemplateMutation,
  useUpdateImageLayoutTemplateMutation,
  useDeleteImageLayoutTemplateMutation,
  useGetAssetsQuery,
  useGetLaidOutImagesQuery,
  useCreateLaidOutImageMutation,
  useUpdateLaidOutImageMutation,
  useDeleteLaidOutImageMutation,
  useGetImageSequencesQuery,
  usePreviewLayoutRequestMutation,
  type ImageLayoutTemplate,
  type LaidOutImage,
  type ImageLayoutRequest,
  type ImageLayoutFrameSpec,
  type ImageLayoutCropSpec,
  type ImageLayoutPresentationSpec,
  type Asset,
  type ImageLayoutComputation,
} from "../../api";
import { Button, Card, CardBody, CardHeader, Input } from "../../components/ui";
import { SliderInput } from "../../components/SliderInput";
import { AnchorGrid } from "../../components/AnchorGrid";

interface ImageLayoutsTabProps {
  projectId: string;
}

const PAPER_SIZES = {
  Letter: { width: 8.5, height: 11 },
  "8x10": { width: 8, height: 10 },
  "5x7": { width: 5, height: 7 },
  "4x6": { width: 4, height: 6 },
  A4: { width: 8.27, height: 11.69 },
  "Square 8x8": { width: 8, height: 8 },
  Custom: { width: 8, height: 10 },
};

const ASPECT_RATIOS = {
  None: null,
  "1:1 (Square)": 1,
  "2:3 (Portrait)": 2 / 3,
  "3:2 (Landscape)": 3 / 2,
  "4:5 (Portrait)": 4 / 5,
  "16:9 (Widescreen)": 16 / 9,
  "9:16 (Story)": 9 / 16,
};

const defaultFrame = (): ImageLayoutFrameSpec => ({
  mode: "page",
  fill: "cover",
  page: {
    width_in: 8,
    height_in: 10,
    dpi: 300,
    orientation: "portrait",
    margins_in: { top: 0.5, right: 0.5, bottom: 0.5, left: 0.5 },
  },
});

const defaultCrop = (): ImageLayoutCropSpec => ({
  strategy: "auto",
  ratio: null,
  zoom: 1,
  extent: undefined,
  anchor: "center",
  pan: { x: 0, y: 0 },
  units: "normalized",
  focus: null,
});

const defaultPresentation = (): ImageLayoutPresentationSpec => ({
  user_scale: 1,
  offset_px: { x: 0, y: 0 },
  clamp_to_canvas: false,
});

export const ImageLayoutsTab: React.FC<ImageLayoutsTabProps> = ({
  projectId,
}) => {
  const templatesQuery = useGetImageLayoutTemplatesQuery({ projectId });
  const assetsQuery = useGetAssetsQuery({ projectId });
  const laidOutImagesQuery = useGetLaidOutImagesQuery({ projectId });
  const sequencesQuery = useGetImageSequencesQuery({ projectId });

  const [createTemplate] = useCreateImageLayoutTemplateMutation();
  const [createGlobalTemplate] = useCreateGlobalImageLayoutTemplateMutation();
  const [updateTemplate] = useUpdateImageLayoutTemplateMutation();
  const [deleteTemplate] = useDeleteImageLayoutTemplateMutation();

  const [createLaidOutImage] = useCreateLaidOutImageMutation();
  const [updateLaidOutImage] = useUpdateLaidOutImageMutation();
  const [deleteLaidOutImage] = useDeleteLaidOutImageMutation();

  // Template Editor State
  const [isCreating, setIsCreating] = useState(false);
  const [editingTemplate, setEditingTemplate] =
    useState<ImageLayoutTemplate | null>(null);
  const [previewAssetId, setPreviewAssetId] = useState<string>("");

  // Template Form State
  const [templateName, setTemplateName] = useState("");
  const [templateDescription, setTemplateDescription] = useState("");
  const [isGlobal, setIsGlobal] = useState(false);
  const [paperSize, setPaperSize] = useState<keyof typeof PAPER_SIZES>("8x10");
  const [paperWidth, setPaperWidth] = useState(8);
  const [paperHeight, setPaperHeight] = useState(10);
  const [dpi, setDpi] = useState(300);
  const [orientation, setOrientation] = useState<"portrait" | "landscape">(
    "portrait",
  );
  const [uniformMargins, setUniformMargins] = useState(true);
  const [marginAll, setMarginAll] = useState(0.5);
  const [marginTop, setMarginTop] = useState(0.5);
  const [marginRight, setMarginRight] = useState(0.5);
  const [marginBottom, setMarginBottom] = useState(0.5);
  const [marginLeft, setMarginLeft] = useState(0.5);
  const [fillMode, setFillMode] = useState<"contain" | "cover">("cover");
  const [aspectRatio, setAspectRatio] =
    useState<keyof typeof ASPECT_RATIOS>("None");
  const [cropStrategy, setCropStrategy] = useState<
    "auto" | "anchor" | "focus" | "manual"
  >("auto");
  const [cropRatio, setCropRatio] = useState<number | null>(null);
  const [zoom, setZoom] = useState(1);
  const [panX, setPanX] = useState(0);
  const [panY, setPanY] = useState(0);
  const [anchorPreset, setAnchorPreset] = useState("middle-center");
  const [userScale, setUserScale] = useState(1);
  const [offsetX, setOffsetX] = useState(0);
  const [offsetY, setOffsetY] = useState(0);

  // Laid-Out Images State
  const [selectedLaidOutId, setSelectedLaidOutId] = useState<string | null>(
    null,
  );
  const [batchAssetSource, setBatchAssetSource] = useState<
    "assets" | "sequence"
  >("assets");
  const [batchSequenceId, setBatchSequenceId] = useState("");
  const [batchTemplateId, setBatchTemplateId] = useState("");
  const [isCreatingLaidOut, setIsCreatingLaidOut] = useState(false);
  const [createAssetId, setCreateAssetId] = useState("");
  const [createTemplateId, setCreateTemplateId] = useState("");

  const templates = useMemo(
    () => templatesQuery.data ?? [],
    [templatesQuery.data],
  );
  const assets = useMemo(() => assetsQuery.data ?? [], [assetsQuery.data]);
  const laidOutImages = useMemo(
    () => laidOutImagesQuery.data ?? [],
    [laidOutImagesQuery.data],
  );
  const sequences = useMemo(
    () => sequencesQuery.data ?? [],
    [sequencesQuery.data],
  );

  const previewAsset = useMemo(
    () => assets.find((a) => a.id === previewAssetId),
    [assets, previewAssetId],
  );

  // Derived layout request from current form state (for preview + payload)
  const currentLayout = useMemo(() => {
    const margins = uniformMargins
      ? { top: marginAll, right: marginAll, bottom: marginAll, left: marginAll }
      : {
          top: marginTop,
          right: marginRight,
          bottom: marginBottom,
          left: marginLeft,
        };
    return {
      frame: {
        mode: "page",
        fill: fillMode,
        ratio: ASPECT_RATIOS[aspectRatio] ?? undefined,
        page: {
          width_in: paperWidth,
          height_in: paperHeight,
          dpi,
          orientation,
          margins_in: margins,
        },
      },
      crop: {
        strategy: cropStrategy,
        ratio: cropRatio ?? undefined,
        zoom,
        anchor: anchorPreset,
        pan: { x: panX, y: panY },
        units: "normalized",
        focus: null,
      },
      presentation: {
        user_scale: userScale,
        offset_px: { x: offsetX, y: offsetY },
        clamp_to_canvas: false,
      },
      export: {
        format: "png",
        quality: 90,
        background: "white",
        filename_template: "{name}-{panel}.{ext}",
        out_dir: "./out",
      },
    } satisfies ImageLayoutRequest;
  }, [
    uniformMargins,
    marginAll,
    marginTop,
    marginRight,
    marginBottom,
    marginLeft,
    fillMode,
    aspectRatio,
    paperWidth,
    paperHeight,
    dpi,
    orientation,
    cropStrategy,
    cropRatio,
    zoom,
    anchorPreset,
    panX,
    panY,
    userScale,
    offsetX,
    offsetY,
  ]);

  const [triggerPreviewLayout, { isLoading: isPreviewLoading }] =
    usePreviewLayoutRequestMutation();
  const [previewResult, setPreviewResult] =
    useState<ImageLayoutComputation | null>(null);
  const [previewError, setPreviewError] = useState<string | null>(null);
  const previewDebounceRef = useRef<number | null>(null);

  const previewCanvas = previewResult?.result?.canvas_rect;
  const previewTarget = previewResult?.result?.target_rect;
  const previewSource = previewResult?.result?.source_rect;
  const previewCanvasScale = useMemo(() => {
    if (!previewCanvas) return 1;
    const safeW = Math.max(previewCanvas.w, 1);
    const safeH = Math.max(previewCanvas.h, 1);
    return Math.min(420 / safeW, 420 / safeH);
  }, [previewCanvas]);

  const previewImagePlacement = useMemo(() => {
    if (!previewAsset || !previewSource || !previewTarget) {
      return null;
    }
    const layoutScale = previewResult?.result?.scale ?? 1;
    const scale = layoutScale * previewCanvasScale;
    return {
      imgWidth: previewAsset.width * scale,
      imgHeight: previewAsset.height * scale,
      offsetX: -previewSource.x * scale,
      offsetY: -previewSource.y * scale,
      targetWidth: previewTarget.w * previewCanvasScale,
      targetHeight: previewTarget.h * previewCanvasScale,
      targetLeft: previewTarget.x * previewCanvasScale,
      targetTop: previewTarget.y * previewCanvasScale,
    };
  }, [previewAsset, previewSource, previewTarget, previewResult, previewCanvasScale]);

  useEffect(() => {
    if (previewDebounceRef.current) {
      window.clearTimeout(previewDebounceRef.current);
    }

    if (!isCreating && !editingTemplate) {
      setPreviewResult(null);
      setPreviewError(null);
      return;
    }

    if (!previewAsset) {
      setPreviewResult(null);
      setPreviewError("Select an asset to preview");
      return;
    }
    if (previewAsset.width <= 0 || previewAsset.height <= 0) {
      setPreviewResult(null);
      setPreviewError("Preview asset is missing dimensions");
      return;
    }
    if (paperWidth <= 0 || paperHeight <= 0 || dpi <= 0) {
      setPreviewResult(null);
      setPreviewError("Page dimensions and DPI must be positive");
      return;
    }

    previewDebounceRef.current = window.setTimeout(() => {
      triggerPreviewLayout({
        projectId,
        layout: currentLayout,
        assetId: previewAsset.id,
      })
        .unwrap()
        .then((res) => {
          setPreviewResult(res);
          setPreviewError(null);
        })
        .catch((err) => {
          const msg =
            (err as any)?.data?.error ??
            (err as any)?.error ??
            "Preview failed";
          setPreviewResult(null);
          setPreviewError(typeof msg === "string" ? msg : "Preview failed");
        });
    }, 250);

    return () => {
      if (previewDebounceRef.current) {
        window.clearTimeout(previewDebounceRef.current);
      }
    };
  }, [
    currentLayout,
    dpi,
    editingTemplate,
    isCreating,
    paperHeight,
    paperWidth,
    previewAsset,
    projectId,
    triggerPreviewLayout,
  ]);

  const resetForm = () => {
    setTemplateName("");
    setTemplateDescription("");
    setIsGlobal(false);
    setPaperSize("8x10");
    setPaperWidth(8);
    setPaperHeight(10);
    setDpi(300);
    setOrientation("portrait");
    setUniformMargins(true);
    setMarginAll(0.5);
    setMarginTop(0.5);
    setMarginRight(0.5);
    setMarginBottom(0.5);
    setMarginLeft(0.5);
    setFillMode("cover");
    setAspectRatio("None");
    setCropStrategy("auto");
    setCropRatio(null);
    setZoom(1);
    setPanX(0);
    setPanY(0);
    setAnchorPreset("middle-center");
    setUserScale(1);
    setOffsetX(0);
    setOffsetY(0);
  };

  const loadTemplateIntoForm = (template: ImageLayoutTemplate) => {
    setEditingTemplate(template);
    setTemplateName(template.name);
    setTemplateDescription(template.description ?? "");
    setIsGlobal(template.scope === "global");

    // Handle both legacy and new shapes
    const maybeLayout = template.settings as any;
    const frame = (maybeLayout.frame as any) ?? null;
    const crop = (maybeLayout.crop as any) ?? null;
    const presentation = (maybeLayout.presentation as any) ?? null;

    if (frame && crop && presentation) {
      if (frame.mode === "page" && frame.page) {
        setPaperWidth(frame.page.width_in ?? 8);
        setPaperHeight(frame.page.height_in ?? 10);
        setDpi(frame.page.dpi ?? 300);
        setOrientation(frame.page.orientation ?? "portrait");
        const m = frame.page.margins_in ?? {
          top: 0.5,
          right: 0.5,
          bottom: 0.5,
          left: 0.5,
        };
        setMarginTop(m.top ?? 0.5);
        setMarginRight(m.right ?? 0.5);
        setMarginBottom(m.bottom ?? 0.5);
        setMarginLeft(m.left ?? 0.5);
        const allEqual =
          m.top === m.right && m.top === m.bottom && m.top === m.left;
        setUniformMargins(allEqual);
        if (allEqual) setMarginAll(m.top ?? 0.5);
      }
      setFillMode(frame.fill ?? "cover");
      if (frame.ratio) {
        const ratioEntry = Object.entries(ASPECT_RATIOS).find(
          ([_, val]) =>
            val !== null && Math.abs(val - (frame.ratio ?? 0)) < 0.01,
        );
        setAspectRatio(
          ratioEntry ? (ratioEntry[0] as keyof typeof ASPECT_RATIOS) : "None",
        );
      } else {
        setAspectRatio("None");
      }

      setCropStrategy(crop.strategy ?? "auto");
      setCropRatio(crop.ratio ?? null);
      setZoom(crop.zoom ?? 1);
      setPanX(crop.pan?.x ?? 0);
      setPanY(crop.pan?.y ?? 0);
      setAnchorPreset(crop.anchor ?? "center");

      setUserScale(presentation.user_scale ?? 1);
      setOffsetX(presentation.offset_px?.x ?? 0);
      setOffsetY(presentation.offset_px?.y ?? 0);
      return;
    }

    // Legacy fallback: ImageLayoutViewportSettings
    const settings = template.settings as any;
    setPaperWidth(settings.paper_width_in ?? 8);
    setPaperHeight(settings.paper_height_in ?? 10);
    setDpi(settings.dpi ?? 300);
    setOrientation(settings.orientation ?? "portrait");
    setMarginTop(settings.margin_top_in ?? 0.5);
    setMarginRight(settings.margin_right_in ?? 0.5);
    setMarginBottom(settings.margin_bottom_in ?? 0.5);
    setMarginLeft(settings.margin_left_in ?? 0.5);
    setFillMode(settings.crop_to_fill ? "cover" : "contain");
    setUserScale(settings.user_scale ?? 1);
    setPanX(settings.position_x ?? 0);
    setPanY(settings.position_y ?? 0);
    setAnchorPreset(settings.anchor_preset ?? "center");

    const allEqual =
      settings.margin_top_in === settings.margin_right_in &&
      settings.margin_top_in === settings.margin_bottom_in &&
      settings.margin_top_in === settings.margin_left_in;
    setUniformMargins(allEqual);
    if (allEqual) setMarginAll(settings.margin_top_in ?? 0.5);

    if (settings.crop_ratio) {
      const ratioEntry = Object.entries(ASPECT_RATIOS).find(
        ([_, val]) =>
          val !== null && Math.abs(val - (settings.crop_ratio ?? 0)) < 0.01,
      );
      setAspectRatio(
        ratioEntry ? (ratioEntry[0] as keyof typeof ASPECT_RATIOS) : "None",
      );
    } else {
      setAspectRatio("None");
    }
    setCropStrategy("auto");
    setCropRatio(settings.crop_ratio ?? null);
    setZoom(1);
    setOffsetX(0);
    setOffsetY(0);
  };

  const buildLayoutFromForm = (): ImageLayoutRequest => currentLayout;

  const handleCreateTemplate = async (e: React.FormEvent) => {
    e.preventDefault();
    const layout = buildLayoutFromForm();

    if (isGlobal) {
      await createGlobalTemplate({
        name: templateName || "Untitled Template",
        description: templateDescription || undefined,
        settings: layout,
      }).unwrap();
    } else {
      await createTemplate({
        projectId,
        name: templateName || "Untitled Template",
        description: templateDescription || undefined,
        settings: layout,
      }).unwrap();
    }

    setIsCreating(false);
    resetForm();
  };

  const handleUpdateTemplate = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!editingTemplate) return;
    const layout = buildLayoutFromForm();
    await updateTemplate({
      templateId: editingTemplate.id,
      name: templateName || undefined,
      description: templateDescription || undefined,
      settings: layout,
    }).unwrap();
    setEditingTemplate(null);
    resetForm();
  };

  const handleDeleteTemplate = async (template: ImageLayoutTemplate) => {
    if (!window.confirm(`Delete template "${template.name}"?`)) return;
    await deleteTemplate({
      templateId: template.id,
      scopeKey: template.project_id ?? "global",
    }).unwrap();
    if (editingTemplate?.id === template.id) {
      setEditingTemplate(null);
      resetForm();
    }
  };

  const handlePaperSizeChange = (size: keyof typeof PAPER_SIZES) => {
    setPaperSize(size);
    if (size !== "Custom") {
      setPaperWidth(PAPER_SIZES[size].width);
      setPaperHeight(PAPER_SIZES[size].height);
    }
  };

  const handleUniformMarginsToggle = (checked: boolean) => {
    setUniformMargins(checked);
    if (checked) {
      setMarginAll(marginTop);
    }
  };

  return (
    <div className="space-y-8">
      {/* ═══ SECTION 1: Template Library ═══ */}
      <div>
        <div className="flex items-center justify-between mb-6">
          <div>
            <h2 className="text-2xl font-bold text-gray-900">
              Layout Templates
            </h2>
            <p className="text-sm text-gray-500 mt-1">
              Create reusable layout presets with visual controls
            </p>
          </div>
          <Button
            onClick={() => {
              if (isCreating || editingTemplate) {
                setIsCreating(false);
                setEditingTemplate(null);
                resetForm();
              } else {
                setIsCreating(true);
                resetForm();
                // Select first asset for preview if available
                if (assets.length > 0 && !previewAssetId) {
                  setPreviewAssetId(assets[0]!.id);
                }
              }
            }}
            variant={isCreating || editingTemplate ? "secondary" : "primary"}
          >
            {isCreating || editingTemplate ? "Cancel" : "+ Create Template"}
          </Button>
        </div>

        {/* Template Grid */}
        {!isCreating && !editingTemplate && (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
            {templates.map((template) => (
              <Card key={template.id}>
                <CardBody className="space-y-3">
                  <div className="flex items-start justify-between">
                    <div>
                      <p className="text-xs uppercase tracking-wide text-gray-400 mb-1">
                        {template.scope === "global"
                          ? "🌐 Global"
                          : "📁 Project"}
                      </p>
                      <h3 className="text-lg font-semibold text-gray-900">
                        {template.name}
                      </h3>
                    </div>
                  </div>
                  {template.description && (
                    <p className="text-sm text-gray-600">
                      {template.description}
                    </p>
                  )}
                  <div className="text-xs text-gray-500">
                    {(() => {
                      const settings = template.settings as any;
                      const page = settings.frame?.page;
                      const w = page?.width_in ?? settings.paper_width_in ?? 8;
                      const h =
                        page?.height_in ?? settings.paper_height_in ?? 10;
                      const dpi = page?.dpi ?? settings.dpi ?? 300;
                      return `${w} × ${h}\" • ${dpi} DPI`;
                    })()}
                  </div>
                  <div className="flex space-x-2 pt-2">
                    <Button
                      size="sm"
                      variant="secondary"
                      onClick={() => loadTemplateIntoForm(template)}
                    >
                      Edit
                    </Button>
                    <Button
                      size="sm"
                      variant="danger"
                      onClick={() => handleDeleteTemplate(template)}
                    >
                      Delete
                    </Button>
                  </div>
                </CardBody>
              </Card>
            ))}
            {templates.length === 0 && (
              <div className="col-span-full text-center py-12 text-gray-500">
                <p className="text-sm">
                  No templates yet. Create one to get started.
                </p>
              </div>
            )}
          </div>
        )}

        {/* Template Editor */}
        {(isCreating || editingTemplate) && (
          <Card>
            <CardHeader>
              <h3 className="text-xl font-semibold text-gray-900">
                {editingTemplate ? "Edit Template" : "Create Template"}
              </h3>
            </CardHeader>
            <CardBody>
              <form
                onSubmit={
                  editingTemplate ? handleUpdateTemplate : handleCreateTemplate
                }
              >
                <div className="grid lg:grid-cols-2 gap-8">
                  {/* Left: Form Controls */}
                  <div className="space-y-6">
                    {/* Basic Info */}
                    <div className="space-y-4">
                      <Input
                        label="Template Name"
                        value={templateName}
                        onChange={(e) => setTemplateName(e.target.value)}
                        placeholder="e.g., 8×10 Portrait Crop"
                        required
                      />
                      <Input
                        label="Description (optional)"
                        value={templateDescription}
                        onChange={(e) => setTemplateDescription(e.target.value)}
                        placeholder="e.g., Full bleed portrait for book pages"
                      />

                      <div>
                        <label className="flex items-center space-x-2 text-sm font-medium text-gray-700">
                          <input
                            type="checkbox"
                            checked={isGlobal}
                            onChange={(e) => setIsGlobal(e.target.checked)}
                            className="rounded border-gray-300 text-primary-600 focus:ring-primary-500"
                          />
                          <span>
                            Make this a global template (available to all
                            projects)
                          </span>
                        </label>
                      </div>
                    </div>

                    <hr className="border-gray-200" />

                    {/* Page Setup */}
                    <div className="space-y-4">
                      <h4 className="text-sm font-semibold text-gray-700 uppercase tracking-wide">
                        Page Setup
                      </h4>

                      <div>
                        <label className="block text-sm font-medium text-gray-700 mb-2">
                          Paper Size
                        </label>
                        <select
                          value={paperSize}
                          onChange={(e) =>
                            handlePaperSizeChange(
                              e.target.value as keyof typeof PAPER_SIZES,
                            )
                          }
                          className="w-full border border-gray-300 rounded-md px-3 py-2 focus:border-primary-500 focus:ring-1 focus:ring-primary-500"
                        >
                          {Object.keys(PAPER_SIZES).map((size) => (
                            <option key={size} value={size}>
                              {size}
                            </option>
                          ))}
                        </select>
                      </div>

                      {paperSize === "Custom" && (
                        <div className="grid grid-cols-2 gap-3">
                          <Input
                            label="Width (inches)"
                            type="number"
                            value={paperWidth}
                            onChange={(e) =>
                              setPaperWidth(parseFloat(e.target.value) || 8)
                            }
                            step="0.1"
                            min="1"
                          />
                          <Input
                            label="Height (inches)"
                            type="number"
                            value={paperHeight}
                            onChange={(e) =>
                              setPaperHeight(parseFloat(e.target.value) || 10)
                            }
                            step="0.1"
                            min="1"
                          />
                        </div>
                      )}

                      <SliderInput
                        label="DPI"
                        value={dpi}
                        onChange={setDpi}
                        min={72}
                        max={600}
                        step={1}
                      />

                      <div>
                        <label className="block text-sm font-medium text-gray-700 mb-2">
                          Orientation
                        </label>
                        <div className="flex space-x-3">
                          <button
                            type="button"
                            onClick={() => setOrientation("portrait")}
                            className={`flex-1 py-2 px-4 rounded-md border-2 transition-all ${
                              orientation === "portrait"
                                ? "border-primary-600 bg-primary-50 text-primary-900"
                                : "border-gray-300 bg-white text-gray-700 hover:border-primary-400"
                            }`}
                          >
                            Portrait ⬜
                          </button>
                          <button
                            type="button"
                            onClick={() => setOrientation("landscape")}
                            className={`flex-1 py-2 px-4 rounded-md border-2 transition-all ${
                              orientation === "landscape"
                                ? "border-primary-600 bg-primary-50 text-primary-900"
                                : "border-gray-300 bg-white text-gray-700 hover:border-primary-400"
                            }`}
                          >
                            Landscape ▭
                          </button>
                        </div>
                      </div>
                    </div>

                    <hr className="border-gray-200" />

                    {/* Margins */}
                    <div className="space-y-4">
                      <h4 className="text-sm font-semibold text-gray-700 uppercase tracking-wide">
                        Margins
                      </h4>

                      <div>
                        <label className="flex items-center space-x-2 text-sm font-medium text-gray-700">
                          <input
                            type="checkbox"
                            checked={uniformMargins}
                            onChange={(e) =>
                              handleUniformMarginsToggle(e.target.checked)
                            }
                            className="rounded border-gray-300 text-primary-600 focus:ring-primary-500"
                          />
                          <span>Uniform margins</span>
                        </label>
                      </div>

                      {uniformMargins ? (
                        <SliderInput
                          label="All Margins"
                          value={marginAll}
                          onChange={(val) => {
                            setMarginAll(val);
                            setMarginTop(val);
                            setMarginRight(val);
                            setMarginBottom(val);
                            setMarginLeft(val);
                          }}
                          min={0}
                          max={2}
                          step={0.05}
                          unit="in"
                        />
                      ) : (
                        <div className="grid grid-cols-2 gap-3">
                          <SliderInput
                            label="Top"
                            value={marginTop}
                            onChange={setMarginTop}
                            min={0}
                            max={2}
                            step={0.05}
                            unit="in"
                          />
                          <SliderInput
                            label="Right"
                            value={marginRight}
                            onChange={setMarginRight}
                            min={0}
                            max={2}
                            step={0.05}
                            unit="in"
                          />
                          <SliderInput
                            label="Bottom"
                            value={marginBottom}
                            onChange={setMarginBottom}
                            min={0}
                            max={2}
                            step={0.05}
                            unit="in"
                          />
                          <SliderInput
                            label="Left"
                            value={marginLeft}
                            onChange={setMarginLeft}
                            min={0}
                            max={2}
                            step={0.05}
                            unit="in"
                          />
                        </div>
                      )}
                    </div>

                    <hr className="border-gray-200" />

                    {/* Crop Settings */}
                    <div className="space-y-4">
                      <h4 className="text-sm font-semibold text-gray-700 uppercase tracking-wide">
                        Frame & Crop
                      </h4>

                      <div>
                        <label className="block text-sm font-medium text-gray-700 mb-2">
                          Fill Mode
                        </label>
                        <div className="space-y-2">
                          <label className="flex items-center space-x-2">
                            <input
                              type="radio"
                              checked={fillMode === "cover"}
                              onChange={() => setFillMode("cover")}
                              className="text-primary-600 focus:ring-primary-500"
                            />
                            <span className="text-sm text-gray-700">
                              Fill (cover) - Crop to fill entire canvas
                            </span>
                          </label>
                          <label className="flex items-center space-x-2">
                            <input
                              type="radio"
                              checked={fillMode === "contain"}
                              onChange={() => setFillMode("contain")}
                              className="text-primary-600 focus:ring-primary-500"
                            />
                            <span className="text-sm text-gray-700">
                              Fit (contain) - Show entire image with
                              letterboxing
                            </span>
                          </label>
                        </div>
                      </div>

                      <div>
                        <label className="block text-sm font-medium text-gray-700 mb-2">
                          Aspect Ratio
                        </label>
                        <select
                          value={aspectRatio}
                          onChange={(e) =>
                            setAspectRatio(
                              e.target.value as keyof typeof ASPECT_RATIOS,
                            )
                          }
                          className="w-full border border-gray-300 rounded-md px-3 py-2 focus:border-primary-500 focus:ring-1 focus:ring-primary-500"
                        >
                          {Object.keys(ASPECT_RATIOS).map((ratio) => (
                            <option key={ratio} value={ratio}>
                              {ratio}
                            </option>
                          ))}
                        </select>
                      </div>

                      <div>
                        <label className="block text-sm font-medium text-gray-700 mb-2">
                          Crop Strategy
                        </label>
                        <select
                          value={cropStrategy}
                          onChange={(e) =>
                            setCropStrategy(
                              e.target.value as
                                | "auto"
                                | "anchor"
                                | "focus"
                                | "manual",
                            )
                          }
                          className="w-full border border-gray-300 rounded-md px-3 py-2 focus:border-primary-500 focus:ring-1 focus:ring-primary-500"
                        >
                          <option value="auto">Auto center</option>
                          <option value="anchor">Anchor preset</option>
                          <option value="manual">Manual pan</option>
                        </select>
                      </div>
                    </div>

                    <hr className="border-gray-200" />

                    {/* Crop Position & Presentation */}
                    <div className="space-y-4">
                      <h4 className="text-sm font-semibold text-gray-700 uppercase tracking-wide">
                        Position & Presentation
                      </h4>

                      {cropStrategy === "anchor" && (
                        <AnchorGrid
                          value={anchorPreset}
                          onChange={setAnchorPreset}
                        />
                      )}

                      {cropStrategy === "manual" && (
                        <div className="grid grid-cols-2 gap-3">
                          <SliderInput
                            label="Pan X"
                            value={panX}
                            onChange={setPanX}
                            min={-1}
                            max={1}
                            step={0.01}
                            unit="norm"
                          />
                          <SliderInput
                            label="Pan Y"
                            value={panY}
                            onChange={setPanY}
                            min={-1}
                            max={1}
                            step={0.01}
                            unit="norm"
                          />
                        </div>
                      )}

                      <SliderInput
                        label="User Scale"
                        value={userScale}
                        onChange={setUserScale}
                        min={0.5}
                        max={2}
                        step={0.05}
                        unit="×"
                      />

                      <div className="grid grid-cols-2 gap-3">
                        <Input
                          label="Offset X (px)"
                          type="number"
                          value={offsetX}
                          onChange={(e) =>
                            setOffsetX(parseFloat(e.target.value) || 0)
                          }
                        />
                        <Input
                          label="Offset Y (px)"
                          type="number"
                          value={offsetY}
                          onChange={(e) =>
                            setOffsetY(parseFloat(e.target.value) || 0)
                          }
                        />
                      </div>
                    </div>

                    {/* Submit Buttons */}
                    <div className="flex justify-end space-x-3 pt-4">
                      <Button
                        type="button"
                        variant="secondary"
                        onClick={() => {
                          setIsCreating(false);
                          setEditingTemplate(null);
                          resetForm();
                        }}
                      >
                        Cancel
                      </Button>
                      <Button type="submit">
                        {editingTemplate ? "Save Changes" : "Create Template"}
                      </Button>
                    </div>
                  </div>

                  {/* Right: Live Preview */}
                  <div className="space-y-4">
                    <div>
                      <label className="block text-sm font-medium text-gray-700 mb-2">
                        Preview Asset
                      </label>
                      <select
                        value={previewAssetId}
                        onChange={(e) => setPreviewAssetId(e.target.value)}
                        className="w-full border border-gray-300 rounded-md px-3 py-2 focus:border-primary-500 focus:ring-1 focus:ring-primary-500"
                      >
                        <option value="">Select an asset to preview...</option>
                        {assets.map((asset) => (
                          <option key={asset.id} value={asset.id}>
                            {asset.filename} ({asset.width}×{asset.height})
                          </option>
                        ))}
                      </select>
                    </div>

                    {/* Preview Panel */}
                    <div className="border-2 border-gray-300 rounded-lg p-6 bg-gray-50 min-h-[500px] flex flex-col">
                      {previewAsset ? (
                        <div className="w-full space-y-3">
                          <div className="flex items-center justify-between text-xs text-gray-500">
                            <span>Backend preview updates as you edit.</span>
                            {isPreviewLoading && (
                            <span className="text-primary-600 animate-pulse">
                                Computing...
                              </span>
                            )}
                          </div>

                          {previewError && (
                            <div className="rounded-md border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-700">
                              {previewError}
                            </div>
                          )}

                          <div className="flex flex-1 items-center justify-center">
                            {previewResult &&
                            previewCanvas &&
                            previewTarget &&
                            previewImagePlacement ? (
                              <div
                                className="relative bg-white border-2 border-gray-400 shadow-lg"
                                style={{
                                  width: `${previewCanvas.w * previewCanvasScale}px`,
                                  height: `${previewCanvas.h * previewCanvasScale}px`,
                                }}
                              >
                                <div className="absolute inset-0 bg-gray-50" />
                                <div
                                  className="absolute border-2 border-primary-500/80 bg-primary-100/60 overflow-hidden shadow-inner"
                                  style={{
                                    left: `${previewImagePlacement.targetLeft}px`,
                                    top: `${previewImagePlacement.targetTop}px`,
                                    width: `${previewImagePlacement.targetWidth}px`,
                                    height: `${previewImagePlacement.targetHeight}px`,
                                  }}
                                >
                                  <img
                                    src={
                                      previewAsset.url ??
                                      `/projects/${projectId}/images/${previewAsset.filename}`
                                    }
                                    alt={previewAsset.filename}
                                    className="pointer-events-none select-none"
                                    style={{
                                      position: "absolute",
                                      width: `${previewImagePlacement.imgWidth}px`,
                                      height: `${previewImagePlacement.imgHeight}px`,
                                      left: `${previewImagePlacement.offsetX}px`,
                                      top: `${previewImagePlacement.offsetY}px`,
                                    }}
                                  />
                                  <div className="absolute inset-0 border border-white/70 pointer-events-none" />
                                </div>
                                <div className="absolute inset-0 border border-dashed border-gray-300 pointer-events-none" />
                              </div>
                            ) : (
                              <div className="text-center text-gray-500 text-sm">
                                {isPreviewLoading
                                  ? "Computing preview..."
                                  : "Adjust settings to compute preview geometry."}
                              </div>
                            )}
                          </div>

                          <div className="space-y-1 text-sm text-gray-600 text-center">
                            <p className="font-medium">{previewAsset.filename}</p>
                            {previewResult && previewResult.result && (
                              <div className="grid grid-cols-2 gap-1 text-xs text-gray-600">
                                <span>
                                  Canvas:{" "}
                                  {Math.round(previewResult.result.canvas_rect.w)} ×{" "}
                                  {Math.round(previewResult.result.canvas_rect.h)} px
                                </span>
                                <span>
                                  Target:{" "}
                                  {Math.round(previewResult.result.target_rect.w)} ×{" "}
                                  {Math.round(previewResult.result.target_rect.h)} px
                                </span>
                                <span>
                                  Source:{" "}
                                  {Math.round(previewResult.result.source_rect.w)} ×{" "}
                                  {Math.round(previewResult.result.source_rect.h)} px
                                </span>
                                <span>
                                  Scale: {previewResult.result.scale.toFixed(3)}× (
                                  {previewResult.result.mode})
                                </span>
                              </div>
                            )}
                          </div>
                        </div>
                      ) : (
                        <div className="flex flex-1 flex-col items-center justify-center text-center text-gray-400">
                          <div className="text-6xl mb-4">📐</div>
                          <p className="text-sm">
                            Select an asset to preview the template
                          </p>
                        </div>
                      )}
                    </div>
                  </div>
                </div>
              </form>
            </CardBody>
          </Card>
        )}
      </div>

      {/* ═══ SECTION 2: Laid-Out Images ═══ */}
      <div>
        <div className="flex items-center justify-between mb-6">
          <div>
            <h2 className="text-2xl font-bold text-gray-900">
              Laid-Out Images ({laidOutImages.length})
            </h2>
            <p className="text-sm text-gray-500 mt-1">
              Apply templates to assets to create production-ready layouts
            </p>
          </div>
          <Button
            onClick={() => {
              setIsCreatingLaidOut(!isCreatingLaidOut);
              setCreateAssetId("");
              setCreateTemplateId("");
            }}
            variant={isCreatingLaidOut ? "secondary" : "primary"}
          >
            {isCreatingLaidOut ? "Cancel" : "+ Create Laid-Out Image"}
          </Button>
        </div>

        {/* Create Form */}
        {isCreatingLaidOut && (
          <Card className="mb-6">
            <CardHeader>
              <h3 className="text-lg font-semibold text-gray-900">
                Create New Laid-Out Image
              </h3>
            </CardHeader>
            <CardBody>
              <form
                onSubmit={async (e) => {
                  e.preventDefault();
                  if (!createAssetId || !createTemplateId) {
                    alert("Select both asset and template");
                    return;
                  }
                  await createLaidOutImage({
                    projectId,
                    assetId: createAssetId,
                    templateId: createTemplateId,
                  }).unwrap();
                  setIsCreatingLaidOut(false);
                  setCreateAssetId("");
                  setCreateTemplateId("");
                }}
                className="grid md:grid-cols-3 gap-4"
              >
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-2">
                    Asset
                  </label>
                  <select
                    value={createAssetId}
                    onChange={(e) => setCreateAssetId(e.target.value)}
                    className="w-full border border-gray-300 rounded-md px-3 py-2 focus:border-primary-500 focus:ring-1 focus:ring-primary-500"
                    required
                  >
                    <option value="">Select asset...</option>
                    {assets.map((asset) => (
                      <option key={asset.id} value={asset.id}>
                        {asset.filename}
                      </option>
                    ))}
                  </select>
                </div>

                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-2">
                    Template
                  </label>
                  <select
                    value={createTemplateId}
                    onChange={(e) => setCreateTemplateId(e.target.value)}
                    className="w-full border border-gray-300 rounded-md px-3 py-2 focus:border-primary-500 focus:ring-1 focus:ring-primary-500"
                    required
                  >
                    <option value="">Select template...</option>
                    {templates.map((tpl) => (
                      <option key={tpl.id} value={tpl.id}>
                        {tpl.name} ({tpl.scope})
                      </option>
                    ))}
                  </select>
                </div>

                <div className="flex items-end">
                  <Button type="submit" className="w-full">
                    Create
                  </Button>
                </div>
              </form>
            </CardBody>
          </Card>
        )}

        {/* Batch Apply Section */}
        <Card className="mb-6">
          <CardHeader>
            <h3 className="text-lg font-semibold text-gray-900">
              Batch Apply Template
            </h3>
            <p className="text-sm text-gray-500 mt-1">
              Apply a template to multiple assets at once
            </p>
          </CardHeader>
          <CardBody>
            <form
              onSubmit={async (e) => {
                e.preventDefault();
                if (!batchTemplateId) {
                  alert("Select a template");
                  return;
                }
                // TODO: Implement batch apply with progress indicator
                alert("Batch apply not yet implemented - coming soon!");
              }}
              className="grid md:grid-cols-4 gap-4 items-end"
            >
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-2">
                  Source
                </label>
                <select
                  value={batchAssetSource}
                  onChange={(e) =>
                    setBatchAssetSource(e.target.value as "assets" | "sequence")
                  }
                  className="w-full border border-gray-300 rounded-md px-3 py-2 focus:border-primary-500 focus:ring-1 focus:ring-primary-500"
                >
                  <option value="assets">All Assets</option>
                  <option value="sequence">From Sequence</option>
                </select>
              </div>

              {batchAssetSource === "sequence" && (
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-2">
                    Sequence
                  </label>
                  <select
                    value={batchSequenceId}
                    onChange={(e) => setBatchSequenceId(e.target.value)}
                    className="w-full border border-gray-300 rounded-md px-3 py-2 focus:border-primary-500 focus:ring-1 focus:ring-primary-500"
                  >
                    <option value="">Select sequence...</option>
                    {sequences.map((seq) => (
                      <option key={seq.id} value={seq.id}>
                        {seq.name}
                      </option>
                    ))}
                  </select>
                </div>
              )}

              <div>
                <label className="block text-sm font-medium text-gray-700 mb-2">
                  Template
                </label>
                <select
                  value={batchTemplateId}
                  onChange={(e) => setBatchTemplateId(e.target.value)}
                  className="w-full border border-gray-300 rounded-md px-3 py-2 focus:border-primary-500 focus:ring-1 focus:ring-primary-500"
                >
                  <option value="">Select template...</option>
                  {templates.map((tpl) => (
                    <option key={tpl.id} value={tpl.id}>
                      {tpl.name} ({tpl.scope})
                    </option>
                  ))}
                </select>
              </div>

              <Button type="submit" disabled={!batchTemplateId}>
                Apply to All
              </Button>
            </form>
          </CardBody>
        </Card>

        {/* Laid-Out Images Grid */}
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
          {laidOutImages.map((image) => {
            const asset = assets.find((a) => a.id === image.asset_id);
            const template = templates.find((t) => t.id === image.template_id);
            return (
              <div
                key={image.id}
                onClick={() => setSelectedLaidOutId(image.id)}
              >
                <Card
                  className={`cursor-pointer border-2 transition-all ${
                    selectedLaidOutId === image.id
                      ? "border-primary-500 shadow-md"
                      : "border-gray-200 hover:border-primary-300"
                  }`}
                >
                  <CardBody className="space-y-3">
                    {/* Thumbnail Preview */}
                    <div className="aspect-[4/3] bg-gray-100 rounded overflow-hidden flex items-center justify-center">
                      {asset && (
                        <img
                          src={
                            asset.url ??
                            `/projects/${projectId}/images/${asset.filename}`
                          }
                          alt={asset.filename}
                          className="max-w-full max-h-full object-contain"
                        />
                      )}
                    </div>

                    <div>
                      <p className="text-sm font-semibold text-gray-900 truncate">
                        {asset?.filename ?? image.asset_id}
                      </p>
                      <p className="text-xs text-gray-500 truncate">
                        Template: {template?.name ?? image.template_id}
                      </p>
                    </div>

                    <div className="flex space-x-2">
                      <Button
                        size="sm"
                        variant="secondary"
                        onClick={(e) => {
                          e.stopPropagation();
                          // TODO: Open edit drawer
                          alert("Edit drawer not yet implemented");
                        }}
                      >
                        Edit
                      </Button>
                      <Button
                        size="sm"
                        variant="danger"
                        onClick={(e) => {
                          e.stopPropagation();
                          if (window.confirm("Delete this laid-out image?")) {
                            deleteLaidOutImage({
                              id: image.id,
                              projectId,
                            }).unwrap();
                          }
                        }}
                      >
                        Delete
                      </Button>
                    </div>
                  </CardBody>
                </Card>
              </div>
            );
          })}
          {laidOutImages.length === 0 && (
            <div className="col-span-full text-center py-12 text-gray-500">
              <p className="text-sm">No laid-out images yet.</p>
              <p className="text-xs mt-2">
                Create a template above, then apply it to your assets.
              </p>
            </div>
          )}
        </div>
      </div>
    </div>
  );
};
