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
  useDeleteLaidOutImageMutation,
  useGetImageSequencesQuery,
  type ImageLayoutTemplate,
} from "../../api";
import { Button, Card, CardBody, CardHeader, Input } from "../../components/ui";
import { ImageLayoutCropControls } from "../../components/ImageLayoutCropControls";
import { ImageLayoutCompareModal } from "../../components/ImageLayoutCompareModal";
import { ImageLayoutFrameForm } from "../../components/ImageLayoutFrameForm";
import { ImageLayoutPreviewPanel } from "../../components/ImageLayoutPreviewPanel";
import { LaidOutImagesGrid } from "../../components/LaidOutImagesGrid";
import { useAppDispatch, useAppSelector } from "../../hooks/redux";
import {
  clearRender,
  previewLayoutThunk,
  renderLayoutThunk,
  resetForm,
  selectCurrentLayout,
  selectImageLayoutsEditor,
  selectPreviewState,
  setAnchorPreset,
  setAspectRatio,
  setClampToCanvas,
  setCompareOpen,
  setCropStrategy,
  setDpi,
  setFillMode,
  setIsGlobal,
  setMarginAll,
  setMarginBottom,
  setMarginLeft,
  setMarginRight,
  setMarginTop,
  setOffsetX,
  setOffsetY,
  setOrientation,
  setPanX,
  setPanY,
  setPaperHeight,
  setPaperSize,
  setPaperWidth,
  setPreviewAssetId,
  setPreviewError,
  setTemplateDescription,
  setTemplateName,
  setUniformMargins,
  setUserScale,
  setZoom,
  startCreate,
  startEdit,
} from "../../state/imageLayoutsEditorSlice";

interface ImageLayoutsTabProps {
  projectId: string;
}

export const ImageLayoutsTab: React.FC<ImageLayoutsTabProps> = ({ projectId }) => {
  const dispatch = useAppDispatch();
  const editor = useAppSelector(selectImageLayoutsEditor);
  const previewState = useAppSelector(selectPreviewState);
  const currentLayout = useAppSelector(selectCurrentLayout);

  const templatesQuery = useGetImageLayoutTemplatesQuery({ projectId });
  const assetsQuery = useGetAssetsQuery({ projectId });
  const laidOutImagesQuery = useGetLaidOutImagesQuery({ projectId });
  const sequencesQuery = useGetImageSequencesQuery({ projectId });

  const [createTemplate] = useCreateImageLayoutTemplateMutation();
  const [createGlobalTemplate] = useCreateGlobalImageLayoutTemplateMutation();
  const [updateTemplate] = useUpdateImageLayoutTemplateMutation();
  const [deleteTemplate] = useDeleteImageLayoutTemplateMutation();
  const [createLaidOutImage] = useCreateLaidOutImageMutation();
  const [deleteLaidOutImage] = useDeleteLaidOutImageMutation();

  const [selectedLaidOutId, setSelectedLaidOutId] = useState<string | null>(null);
  const [batchAssetSource, setBatchAssetSource] = useState<"assets" | "sequence">("assets");
  const [batchSequenceId, setBatchSequenceId] = useState("");
  const [batchTemplateId, setBatchTemplateId] = useState("");
  const [isCreatingLaidOut, setIsCreatingLaidOut] = useState(false);
  const [createAssetId, setCreateAssetId] = useState("");
  const [createTemplateId, setCreateTemplateId] = useState("");

  const templates = useMemo(() => templatesQuery.data ?? [], [templatesQuery.data]);
  const assets = useMemo(() => assetsQuery.data ?? [], [assetsQuery.data]);
  const laidOutImages = useMemo(
    () => laidOutImagesQuery.data ?? [],
    [laidOutImagesQuery.data],
  );
  const sequences = useMemo(() => sequencesQuery.data ?? [], [sequencesQuery.data]);

  const previewAsset = useMemo(
    () => assets.find((asset) => asset.id === previewState.assetId),
    [assets, previewState.assetId],
  );

  const previewDebounceRef = useRef<number | null>(null);

  const previewCanvas = previewState.result?.result?.canvas_rect;
  const previewTarget = previewState.result?.result?.target_rect;
  const previewSource = previewState.result?.result?.source_rect;

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
    const layoutScale = previewState.result?.result?.scale ?? 1;
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
  }, [previewAsset, previewSource, previewTarget, previewState, previewCanvasScale]);

  useEffect(() => {
    if (previewDebounceRef.current) {
      window.clearTimeout(previewDebounceRef.current);
    }

    if (editor.meta.mode === "idle") {
      dispatch(setPreviewError(null));
      return;
    }

    if (!previewAsset) {
      dispatch(setPreviewError("Select an asset to preview"));
      return;
    }
    if (previewAsset.width <= 0 || previewAsset.height <= 0) {
      dispatch(setPreviewError("Preview asset is missing dimensions"));
      return;
    }
    if (editor.frame.paperWidth <= 0 || editor.frame.paperHeight <= 0 || editor.frame.dpi <= 0) {
      dispatch(setPreviewError("Page dimensions and DPI must be positive"));
      return;
    }

    dispatch(setPreviewError(null));
    previewDebounceRef.current = window.setTimeout(() => {
      dispatch(previewLayoutThunk({ projectId }));
    }, 250);

    return () => {
      if (previewDebounceRef.current) {
        window.clearTimeout(previewDebounceRef.current);
      }
    };
  }, [
    dispatch,
    editor.meta.mode,
    editor.frame.paperWidth,
    editor.frame.paperHeight,
    editor.frame.dpi,
    previewAsset,
    projectId,
    currentLayout,
  ]);

  useEffect(() => {
    dispatch(clearRender());
  }, [dispatch, currentLayout, previewState.assetId]);

  useEffect(() => {
    return () => {
      if (previewState.renderUrl) {
        URL.revokeObjectURL(previewState.renderUrl);
      }
    };
  }, [previewState.renderUrl]);

  const handleToggleEditor = () => {
    if (editor.meta.mode !== "idle") {
      dispatch(resetForm());
      return;
    }
    dispatch(startCreate());
    if (assets.length > 0 && !previewState.assetId) {
      dispatch(setPreviewAssetId(assets[0]!.id));
    }
  };

  const handleCreateTemplate = async (e: React.FormEvent) => {
    e.preventDefault();
    const layout = currentLayout;

    if (editor.meta.isGlobal) {
      await createGlobalTemplate({
        name: editor.meta.templateName || "Untitled Template",
        description: editor.meta.templateDescription || undefined,
        settings: layout,
      }).unwrap();
    } else {
      await createTemplate({
        projectId,
        name: editor.meta.templateName || "Untitled Template",
        description: editor.meta.templateDescription || undefined,
        settings: layout,
      }).unwrap();
    }

    dispatch(resetForm());
  };

  const handleUpdateTemplate = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!editor.meta.editingTemplateId) return;
    const layout = currentLayout;
    await updateTemplate({
      templateId: editor.meta.editingTemplateId,
      name: editor.meta.templateName || undefined,
      description: editor.meta.templateDescription || undefined,
      settings: layout,
    }).unwrap();
    dispatch(resetForm());
  };

  const handleDeleteTemplate = async (template: ImageLayoutTemplate) => {
    if (!window.confirm(`Delete template "${template.name}"?`)) return;
    await deleteTemplate({
      templateId: template.id,
      scopeKey: template.project_id ?? "global",
    }).unwrap();
    if (editor.meta.editingTemplateId === template.id) {
      dispatch(resetForm());
    }
  };

  const handleRender = () => {
    dispatch(renderLayoutThunk({ projectId }));
  };

  const handleLoadTemplate = (template: ImageLayoutTemplate) => {
    dispatch(startEdit(template));
    if (assets.length > 0 && !previewState.assetId) {
      dispatch(setPreviewAssetId(assets[0]!.id));
    }
  };

  return (
    <>
      <div className="space-y-8">
        {/* ═══ SECTION 1: Template Library ═══ */}
        <div>
          <div className="flex items-center justify-between mb-6">
            <div>
              <h2 className="text-2xl font-bold text-gray-900">Layout Templates</h2>
              <p className="text-sm text-gray-500 mt-1">
                Create reusable layout presets with visual controls
              </p>
            </div>
            <Button
              onClick={handleToggleEditor}
              variant={editor.meta.mode !== "idle" ? "secondary" : "primary"}
            >
              {editor.meta.mode !== "idle" ? "Cancel" : "+ Create Template"}
            </Button>
          </div>

          {/* Template Grid */}
          {editor.meta.mode === "idle" && (
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
              {templates.map((template) => (
                <Card key={template.id}>
                  <CardBody className="space-y-3">
                    <div className="flex items-start justify-between">
                      <div>
                        <p className="text-xs uppercase tracking-wide text-gray-400 mb-1">
                          {template.scope === "global" ? "🌐 Global" : "📁 Project"}
                        </p>
                        <h3 className="text-lg font-semibold text-gray-900">
                          {template.name}
                        </h3>
                      </div>
                    </div>
                    {template.description && (
                      <p className="text-sm text-gray-600">{template.description}</p>
                    )}
                    <div className="text-xs text-gray-500">
                      {(() => {
                        const settings = template.settings as any;
                        const page = settings.frame?.page;
                        const w = page?.width_in ?? settings.paper_width_in ?? 8;
                        const h = page?.height_in ?? settings.paper_height_in ?? 10;
                        const dpi = page?.dpi ?? settings.dpi ?? 300;
                        return `${w} × ${h}" • ${dpi} DPI`;
                      })()}
                    </div>
                    <div className="flex space-x-2 pt-2">
                      <Button size="sm" variant="secondary" onClick={() => handleLoadTemplate(template)}>
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
                  <p className="text-sm">No templates yet. Create one to get started.</p>
                </div>
              )}
            </div>
          )}

          {/* Template Editor */}
          {editor.meta.mode !== "idle" && (
            <Card>
              <CardHeader>
                <h3 className="text-xl font-semibold text-gray-900">
                  {editor.meta.mode === "edit" ? "Edit Template" : "Create Template"}
                </h3>
              </CardHeader>
              <CardBody>
                <form
                  onSubmit={editor.meta.mode === "edit" ? handleUpdateTemplate : handleCreateTemplate}
                >
                  <div className="grid lg:grid-cols-2 gap-8">
                    {/* Left: Form Controls */}
                    <div className="space-y-6">
                      {/* Basic Info */}
                      <div className="space-y-4">
                        <Input
                          label="Template Name"
                          value={editor.meta.templateName}
                          onChange={(e) => dispatch(setTemplateName(e.target.value))}
                          placeholder="e.g., 8×10 Portrait Crop"
                          required
                        />
                        <Input
                          label="Description (optional)"
                          value={editor.meta.templateDescription}
                          onChange={(e) => dispatch(setTemplateDescription(e.target.value))}
                          placeholder="e.g., Full bleed portrait for book pages"
                        />

                        <div>
                          <label className="flex items-center space-x-2 text-sm font-medium text-gray-700">
                            <input
                              type="checkbox"
                              checked={editor.meta.isGlobal}
                              onChange={(e) => dispatch(setIsGlobal(e.target.checked))}
                              className="rounded border-gray-300 text-primary-600 focus:ring-primary-500"
                            />
                            <span>Make this a global template (available to all projects)</span>
                          </label>
                        </div>
                      </div>

                      <hr className="border-gray-200" />

                      <ImageLayoutFrameForm
                        frame={editor.frame}
                        onPaperSizeChange={(size) => dispatch(setPaperSize(size))}
                        onPaperWidthChange={(val) => dispatch(setPaperWidth(val))}
                        onPaperHeightChange={(val) => dispatch(setPaperHeight(val))}
                        onDpiChange={(val) => dispatch(setDpi(val))}
                        onOrientationChange={(val) => dispatch(setOrientation(val))}
                        onUniformMarginsChange={(checked) => {
                          dispatch(setUniformMargins(checked));
                          if (checked) {
                            dispatch(setMarginAll(editor.frame.margins.top));
                          }
                        }}
                        onMarginAllChange={(val) => dispatch(setMarginAll(val))}
                        onMarginTopChange={(val) => dispatch(setMarginTop(val))}
                        onMarginRightChange={(val) => dispatch(setMarginRight(val))}
                        onMarginBottomChange={(val) => dispatch(setMarginBottom(val))}
                        onMarginLeftChange={(val) => dispatch(setMarginLeft(val))}
                        onFillModeChange={(val) => dispatch(setFillMode(val))}
                        onAspectRatioChange={(val) => dispatch(setAspectRatio(val))}
                      />

                      <hr className="border-gray-200" />

                      <ImageLayoutCropControls
                        cropStrategy={editor.crop.strategy}
                        setCropStrategy={(val) => dispatch(setCropStrategy(val))}
                        anchorPreset={editor.crop.anchorPreset}
                        setAnchorPreset={(val) => dispatch(setAnchorPreset(val))}
                        panX={editor.crop.panX}
                        setPanX={(val) => dispatch(setPanX(val))}
                        panY={editor.crop.panY}
                        setPanY={(val) => dispatch(setPanY(val))}
                        zoom={editor.crop.zoom}
                        setZoom={(val) => dispatch(setZoom(val))}
                        userScale={editor.presentation.userScale}
                        setUserScale={(val) => dispatch(setUserScale(val))}
                        offsetX={editor.presentation.offsetX}
                        setOffsetX={(val) => dispatch(setOffsetX(val))}
                        offsetY={editor.presentation.offsetY}
                        setOffsetY={(val) => dispatch(setOffsetY(val))}
                        clampToCanvas={editor.presentation.clampToCanvas}
                        setClampToCanvas={(val) => dispatch(setClampToCanvas(val))}
                      />

                      {/* Submit Buttons */}
                      <div className="flex justify-end space-x-3 pt-4">
                        <Button type="button" variant="secondary" onClick={() => dispatch(resetForm())}>
                          Cancel
                        </Button>
                        <Button type="submit">
                          {editor.meta.mode === "edit" ? "Save Changes" : "Create Template"}
                        </Button>
                      </div>
                    </div>

                    {/* Right: Live Preview */}
                    <ImageLayoutPreviewPanel
                      projectId={projectId}
                      assets={assets}
                      previewAssetId={previewState.assetId}
                      onPreviewAssetChange={(id) => dispatch(setPreviewAssetId(id))}
                      previewState={previewState}
                      previewResult={previewState.result ?? null}
                      previewCanvasScale={previewCanvasScale}
                      previewPlacement={previewImagePlacement}
                      onRender={handleRender}
                      onOpenCompare={() => dispatch(setCompareOpen(true))}
                    />
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
                <h3 className="text-lg font-semibold text-gray-900">Create New Laid-Out Image</h3>
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
                    <label className="block text-sm font-medium text-gray-700 mb-2">Asset</label>
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
                    <label className="block text-sm font-medium text-gray-700 mb-2">Template</label>
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
              <h3 className="text-lg font-semibold text-gray-900">Batch Apply Template</h3>
              <p className="text-sm text-gray-500 mt-1">
                Apply a template to multiple assets at once
              </p>
            </CardHeader>
            <CardBody>
              <form
                onSubmit={(e) => {
                  e.preventDefault();
                  if (!batchTemplateId) {
                    alert("Select a template");
                    return;
                  }
                  alert("Batch apply not yet implemented - coming soon!");
                }}
                className="grid md:grid-cols-4 gap-4 items-end"
              >
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-2">Source</label>
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
                    <label className="block text-sm font-medium text-gray-700 mb-2">Sequence</label>
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
                  <label className="block text-sm font-medium text-gray-700 mb-2">Template</label>
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
          <LaidOutImagesGrid
            projectId={projectId}
            assets={assets}
            templates={templates}
            laidOutImages={laidOutImages}
            selectedLaidOutId={selectedLaidOutId}
            onSelect={(id) => setSelectedLaidOutId(id)}
            onDelete={(id) => {
              if (window.confirm("Delete this laid-out image?")) {
                void deleteLaidOutImage({ id, projectId }).unwrap();
              }
            }}
            onEdit={(id) => {
              setSelectedLaidOutId(id);
              alert("Edit drawer not yet implemented");
            }}
          />
        </div>
      </div>
      <ImageLayoutCompareModal
        open={previewState.compareOpen}
        onClose={() => dispatch(setCompareOpen(false))}
        previewResult={previewState.result?.result}
        previewAssetUrl={
          previewAsset
            ? previewAsset.url ?? `/projects/${projectId}/images/${previewAsset.filename}`
            : undefined
        }
        previewPlacement={
          previewImagePlacement
            ? {
                targetLeft: previewImagePlacement.targetLeft,
                targetTop: previewImagePlacement.targetTop,
                targetWidth: previewImagePlacement.targetWidth,
                targetHeight: previewImagePlacement.targetHeight,
                imgWidth: previewImagePlacement.imgWidth,
                imgHeight: previewImagePlacement.imgHeight,
                offsetX: previewImagePlacement.offsetX,
                offsetY: previewImagePlacement.offsetY,
              }
            : null
        }
        previewCanvas={previewCanvas ? { w: previewCanvas.w, h: previewCanvas.h } : null}
        previewScale={previewCanvasScale}
        renderUrl={previewState.renderUrl}
      />
    </>
  );
};
