import React from 'react';
import { Link, useParams } from 'react-router-dom';
import {
  useGetPresetYamlQuery,
  useGetPresetsQuery,
  usePutYamlMutation,
  useRenderProjectMutation,
} from '../api';
import { ProjectYamlEditor } from '../components/ProjectYamlEditor';

export const ProjectYamlPage: React.FC = () => {
  const { id = '' } = useParams();
  const { data: presets } = useGetPresetsQuery();
  const [sel, setSel] = React.useState('');
  const [putYaml] = usePutYamlMutation();
  const { data: presetYaml } = useGetPresetYamlQuery({ id: sel }, { skip: !sel });
  const [editorKey, setEditorKey] = React.useState(0);
  const [renderProject, { isLoading: isRendering }] = useRenderProjectMutation();
  const [dims, setDims] = React.useState('1200px,1600px');
  const [bw, setBw] = React.useState(false);
  const [preview, setPreview] = React.useState<{ renderId: string; files: string[] } | null>(null);

  const onLoadPreset = async () => {
    if (!presetYaml) return;
    await putYaml({ id, yaml: presetYaml }).unwrap();
    setEditorKey((k) => k + 1); // remount editor to reload YAML
  };

  const onPreview = async () => {
    const res = await renderProject({
      id,
      test: true,
      test_bw: bw,
      test_dimensions: dims,
    }).unwrap();
    setPreview(res);
  };

  return (
    <main>
      <p>
        <Link to={`/projects/${id}`}>← Back to Project</Link>
      </p>
      <h1>YAML Editor</h1>
      <p style={{ color: '#666', marginTop: -8 }}>
        Edit the project spec.yaml directly. Use presets to start from common templates, and preview
        with readable test images.
      </p>

      <section
        style={{
          margin: '12px 0',
          display: 'flex',
          gap: 8,
          alignItems: 'center',
          flexWrap: 'wrap',
        }}
      >
        <label>
          Preset
          <select value={sel} onChange={(e) => setSel(e.target.value)} style={{ marginLeft: 8 }}>
            <option value="">Select…</option>
            {presets?.presets?.map((p) => (
              <option key={p.id} value={p.id}>
                {p.name}
              </option>
            ))}
          </select>
        </label>
        <button type="button" disabled={!sel} onClick={onLoadPreset}>
          Load into editor
        </button>
        <span style={{ marginLeft: 16 }} />
        <label>
          <input type="checkbox" checked={bw} onChange={(e) => setBw(e.target.checked)} /> BW
        </label>
        <label>
          Test size
          <input
            value={dims}
            onChange={(e) => setDims(e.target.value)}
            placeholder="WIDTH,HEIGHT"
            style={{ marginLeft: 8, width: 180 }}
          />
        </label>
        <button type="button" disabled={isRendering} onClick={onPreview}>
          Preview (test images)
        </button>
      </section>

      <div style={{ display: 'grid', gridTemplateColumns: '1fr', gap: 16 }}>
        <div style={{ minHeight: 200 }}>
          <ProjectYamlEditor key={editorKey} id={id} />
        </div>
        <section>
          <h2>Preview</h2>
          {preview?.files?.length ? (
            <div style={{ display: 'flex', gap: 12, flexWrap: 'wrap' }}>
              {preview.files.map((f) => (
                <img
                  key={f}
                  src={`/api/projects/${id}/renders/${preview.renderId}/files/${encodeURIComponent(f)}`}
                  alt={f}
                  style={{
                    maxWidth: '100%',
                    height: 'auto',
                    border: '1px solid #ddd',
                    background: '#fff',
                  }}
                />
              ))}
            </div>
          ) : (
            <p style={{ color: '#666' }}>
              Click “Preview (test images)” to generate readable preview pages.
            </p>
          )}
        </section>
      </div>
    </main>
  );
};
