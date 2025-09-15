import React from 'react';
import { Link, useParams } from 'react-router-dom';
import { useApplyPresetMutation, useGetPresetsQuery } from '../api';
import { ImageTray } from '../components/ImageTray';
import { ProjectValidationPanel } from '../components/ProjectValidationPanel';
import { ProjectYamlEditor } from '../components/ProjectYamlEditor';

export const ProjectDetail: React.FC = () => {
  const { id = '' } = useParams();
  const { data: presets } = useGetPresetsQuery();
  const [applyPreset, { isLoading: isApplying }] = useApplyPresetMutation();
  const [sel, setSel] = React.useState('');

  return (
    <main>
      <p>
        <Link to="/projects">← Back to Projects</Link>
      </p>
      <h1>Project {id}</h1>
      <section style={{ margin: '12px 0' }}>
        <label>
          Apply preset:
          <select value={sel} onChange={(e) => setSel(e.target.value)} style={{ marginLeft: 8 }}>
            <option value="">Select…</option>
            {presets?.presets?.map((p) => (
              <option key={p.id} value={p.id}>
                {p.name}
              </option>
            ))}
          </select>
        </label>
        <button
          type="button"
          disabled={!sel || isApplying}
          onClick={() =>
            applyPreset({ id, presetId: sel })
              .unwrap()
              .then(() => setSel(''))
          }
          style={{ marginLeft: 8 }}
        >
          Apply
        </button>
      </section>
      <ImageTray id={id} />
      <ProjectValidationPanel id={id} />
      <ProjectYamlEditor id={id} />
    </main>
  );
};
