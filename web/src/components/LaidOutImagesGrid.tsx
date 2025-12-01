import React from "react";
import type { Asset, ImageLayoutTemplate, LaidOutImage } from "../api";
import { Button } from "./ui/Button";
import { Card, CardBody } from "./ui/Card";

interface LaidOutImagesGridProps {
  projectId: string;
  assets: Asset[];
  templates: ImageLayoutTemplate[];
  laidOutImages: LaidOutImage[];
  selectedLaidOutId: string | null;
  onSelect: (id: string) => void;
  onDelete: (id: string) => void;
  onEdit: (id: string) => void;
}

export const LaidOutImagesGrid: React.FC<LaidOutImagesGridProps> = ({
  projectId,
  assets,
  templates,
  laidOutImages,
  selectedLaidOutId,
  onSelect,
  onDelete,
  onEdit,
}) => {
  return (
    <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
      {laidOutImages.map((image) => {
        const asset = assets.find((a) => a.id === image.asset_id);
        const template = templates.find((t) => t.id === image.template_id);
        return (
          <div
            key={image.id}
            onClick={() => onSelect(image.id)}
          >
            <Card
              className={`cursor-pointer border-2 transition-all ${
                selectedLaidOutId === image.id
                  ? "border-primary-500 shadow-md"
                  : "border-gray-200 hover:border-primary-300"
              }`}
            >
              <CardBody className="space-y-3">
                <div className="aspect-[4/3] bg-gray-100 rounded overflow-hidden flex items-center justify-center">
                  {asset && (
                    <img
                      src={
                        asset.url ?? `/projects/${projectId}/images/${asset.filename}`
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
                      onEdit(image.id);
                    }}
                  >
                    Edit
                  </Button>
                  <Button
                    size="sm"
                    variant="danger"
                    onClick={(e) => {
                      e.stopPropagation();
                      onDelete(image.id);
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
  );
};
