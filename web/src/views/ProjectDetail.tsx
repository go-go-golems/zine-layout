import React, { useMemo, useState } from "react";
import { Link, useParams } from "react-router-dom";
import {
  useApplyPresetMutation,
  useGetPresetsQuery,
  useGetImagesQuery,
  type ImageItem,
} from "../api";
import {
  ProjectAssetsPanel,
  type AssetSummary,
} from "../components/bookSpread/ProjectAssetsPanel";
import { ProjectRenderPanel } from "../components/ProjectRenderPanel";
import { ProjectValidationPanel } from "../components/ProjectValidationPanel";
import { Button, Card, CardBody, CardHeader } from "../components/ui";

// Helper function to convert ImageItem to AssetSummary (same as BookSpreadDesigner)
const toAssetSummary = (projectId: string, image: ImageItem): AssetSummary => ({
  id: image.id,
  name: image.name,
  width: image.width,
  height: image.height,
  src: `/api/projects/${projectId}/images/${image.id}`,
  uploadedPath: `/projects/${projectId}/images/${image.id}`,
});

export const ProjectDetail: React.FC = () => {
  const { id = "" } = useParams();
  const { data: presets } = useGetPresetsQuery();
  const [applyPreset, { isLoading: isApplying }] = useApplyPresetMutation();
  const [sel, setSel] = useState("");
  const [selectedAssetId, setSelectedAssetId] = useState<string | null>(null);

  // Load images using the new SQLite-backed query
  const imagesQuery = useGetImagesQuery({ id }, { skip: !id });

  // Convert images to assets format for ProjectAssetsPanel
  const assets: AssetSummary[] = useMemo(() => {
    if (!id || !imagesQuery.data) {
      return [];
    }
    const order = imagesQuery.data.order?.length
      ? imagesQuery.data.order
      : imagesQuery.data.images.map((img) => img.id);
    const map = new Map(
      imagesQuery.data.images.map((img) => [img.id, img] as const)
    );
    return order
      .map((id) => map.get(id))
      .filter((img): img is ImageItem => Boolean(img))
      .map((img) => toAssetSummary(id, img));
  }, [imagesQuery.data, id]);

  const handleAssetSelect = (asset: AssetSummary) => {
    setSelectedAssetId(asset.id);
    // Could potentially dispatch to global state here if needed for future features
    console.log("Selected asset:", asset);
  };

  return (
    <div className="min-h-screen">
      {/* Breadcrumb */}
      <div className="mb-6">
        <nav className="flex items-center space-x-2 text-sm text-gray-600">
          <Link to="/projects" className="hover:text-gray-900">
            Projects
          </Link>
          <span>/</span>
          <span className="text-gray-900 font-medium">Project {id}</span>
        </nav>
      </div>

      {/* Project Header */}
      <div className="flex items-center justify-between mb-8">
        <div>
          <h1 className="text-3xl font-bold text-gray-900">Project {id}</h1>
          <p className="text-gray-600 mt-1">
            Configure layout, manage images, and generate your zine
          </p>
        </div>
        <div className="flex space-x-3">
          <Link to={`/projects/${id}/yaml`}>
            <Button variant="secondary">YAML Editor</Button>
          </Link>
        </div>
      </div>

      {/* Three Column Layout */}
      <div className="grid grid-cols-1 lg:grid-cols-12 gap-6 min-h-0">
        {/* Left Sidebar - Project Settings */}
        <div className="lg:col-span-3">
          <div className="space-y-6 sticky top-6">
            {/* Project Settings */}
            <Card>
              <CardHeader>
                <h3 className="text-lg font-semibold text-gray-900">
                  Project Settings
                </h3>
              </CardHeader>
              <CardBody className="space-y-4">
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">
                    Project Name
                  </label>
                  <input
                    type="text"
                    value={`Project ${id}`}
                    className="input"
                    readOnly
                  />
                </div>

                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">
                    PPI (Pixels per inch)
                  </label>
                  <input type="number" defaultValue="300" className="input" />
                </div>
              </CardBody>
            </Card>

            {/* Presets */}
            <Card>
              <CardHeader>
                <h3 className="text-lg font-semibold text-gray-900">
                  Apply Preset
                </h3>
              </CardHeader>
              <CardBody className="space-y-4">
                <div>
                  <select
                    value={sel}
                    onChange={(e) => setSel(e.target.value)}
                    className="input"
                  >
                    <option value="">Select a preset...</option>
                    {presets?.presets?.map((p) => (
                      <option key={p.id} value={p.id}>
                        {p.name}
                      </option>
                    ))}
                  </select>
                </div>
                <Button
                  disabled={!sel || isApplying}
                  isLoading={isApplying}
                  onClick={() =>
                    applyPreset({ id, presetId: sel })
                      .unwrap()
                      .then(() => setSel(""))
                  }
                  className="w-full"
                >
                  Apply Preset
                </Button>
              </CardBody>
            </Card>
          </div>
        </div>

        {/* Center - Main Content Area */}
        <div className="lg:col-span-6">
          <div className="space-y-6">
            {/* Grid Canvas Placeholder */}
            <Card>
              <CardHeader>
                <h3 className="text-lg font-semibold text-gray-900">
                  Layout Canvas
                </h3>
              </CardHeader>
              <CardBody>
                <div className="bg-gray-50 border-2 border-dashed border-gray-300 rounded-lg h-96 flex items-center justify-center">
                  <div className="text-center">
                    <svg
                      className="w-12 h-12 text-gray-400 mx-auto mb-4"
                      fill="none"
                      stroke="currentColor"
                      viewBox="0 0 24 24"
                    >
                      <path
                        strokeLinecap="round"
                        strokeLinejoin="round"
                        strokeWidth={2}
                        d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10"
                      />
                    </svg>
                    <p className="text-gray-600 mb-2">
                      Visual grid editor coming soon
                    </p>
                    <p className="text-sm text-gray-500">
                      Use the YAML editor to configure layouts for now
                    </p>
                  </div>
                </div>
              </CardBody>
            </Card>

            {/* Project Assets */}
            <Card>
              <CardHeader>
                <h3 className="text-lg font-semibold text-gray-900">Images</h3>
              </CardHeader>
              <CardBody>
                <ProjectAssetsPanel
                  projectId={id || null}
                  assets={assets}
                  selectedAssetId={selectedAssetId}
                  onSelectAsset={handleAssetSelect}
                />
              </CardBody>
            </Card>
          </div>
        </div>

        {/* Right Sidebar - Tools & Actions */}
        <div className="lg:col-span-3">
          <div className="space-y-6 sticky top-6">
            {/* Validation Panel */}
            <Card>
              <CardHeader>
                <h3 className="text-lg font-semibold text-gray-900">
                  Validation
                </h3>
              </CardHeader>
              <CardBody>
                <ProjectValidationPanel id={id} />
              </CardBody>
            </Card>

            {/* Render Panel */}
            <Card>
              <CardHeader>
                <h3 className="text-lg font-semibold text-gray-900">
                  Render & Export
                </h3>
              </CardHeader>
              <CardBody>
                <ProjectRenderPanel id={id} />
              </CardBody>
            </Card>
          </div>
        </div>
      </div>
    </div>
  );
};
