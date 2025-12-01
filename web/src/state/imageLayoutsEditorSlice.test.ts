import { describe, expect, it } from "vitest";
import {
  imageLayoutsEditorReducer,
  selectCurrentLayout,
  startCreate,
  startEdit,
  setPreviewAssetId,
} from "./imageLayoutsEditorSlice";
import type { ImageLayoutTemplate } from "../api";

describe("imageLayoutsEditorSlice", () => {
  const asRootState = (editor: ReturnType<typeof imageLayoutsEditorReducer>) =>
    ({ imageLayoutsEditor: editor } as unknown as import("../store").RootState);

  it("builds a layout from defaults", () => {
    const state = imageLayoutsEditorReducer(undefined, { type: "init" });
    const layout = selectCurrentLayout(asRootState(state));

    expect(layout.frame.page?.width_in).toBeCloseTo(8);
    expect(layout.frame.page?.height_in).toBeCloseTo(10);
    expect(layout.crop.strategy).toBe("auto");
  });

  it("loads template data in new format", () => {
    const template: ImageLayoutTemplate = {
      id: "tpl-new",
      name: "New format",
      description: "test",
      scope: "project",
      created_at: "",
      updated_at: "",
      settings: {
        frame: {
          mode: "page",
          fill: "contain",
          ratio: 1,
          page: {
            width_in: 5,
            height_in: 7,
            dpi: 150,
            orientation: "landscape",
            margins_in: { top: 0.25, right: 0.5, bottom: 0.75, left: 0.5 },
          },
        },
        crop: {
          strategy: "anchor",
          ratio: 1,
          zoom: 1.2,
          anchor: "center",
          pan: { x: 0.1, y: -0.05 },
          units: "normalized",
          focus: null,
        },
        export: {
          format: "png",
          quality: 90,
          background: "white",
          filename_template: "{name}-{panel}.{ext}",
          out_dir: "./out",
        },
      },
    };

    const state = imageLayoutsEditorReducer(undefined, startEdit(template));
    expect(state.meta.mode).toBe("edit");
    expect(state.frame.paperWidth).toBeCloseTo(5);
    expect(state.frame.margins.bottom).toBeCloseTo(0.75);
    expect(state.frame.aspectRatio).toBe("1:1 (Square)");
    expect(state.crop.anchorPreset).toBe("center");
  });

  it("preserves preview asset when starting a new create session", () => {
    const withAsset = imageLayoutsEditorReducer(
      undefined,
      setPreviewAssetId("asset-123"),
    );
    const next = imageLayoutsEditorReducer(withAsset, startCreate());

    expect(next.meta.mode).toBe("create");
    expect(next.preview.assetId).toBe("asset-123");
    expect(next.preview.result).toBeNull();
  });
});
