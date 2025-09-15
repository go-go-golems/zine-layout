import React from 'react';
import { useGetYamlQuery, usePutYamlMutation } from '../api';

export const ProjectYamlEditor: React.FC<{ id: string }> = ({ id }) => {
  const { data: yaml, isLoading, refetch } = useGetYamlQuery({ id });
  const [putYaml, { isLoading: isSaving }] = usePutYamlMutation();
  const [value, setValue] = React.useState('');
  const [dirty, setDirty] = React.useState(false);

  React.useEffect(() => {
    if (!isLoading && yaml !== undefined && !dirty) setValue(yaml);
  }, [yaml, isLoading, dirty]);

  const onSave = async () => {
    await putYaml({ id, yaml: value }).unwrap();
    setDirty(false);
    refetch();
  };

  return (
    <section>
      <h2>YAML</h2>
      {isLoading ? (
        <p>Loading YAML…</p>
      ) : (
        <div>
          <textarea
            value={value}
            onChange={(e) => {
              setValue(e.target.value);
              setDirty(true);
            }}
            rows={16}
            style={{ width: '100%', fontFamily: 'monospace' }}
            placeholder="spec.yaml (optional)"
          />
          <div style={{ marginTop: 8, display: 'flex', gap: 8 }}>
            <button
              type="button"
              onClick={() => {
                setValue(yaml || '');
                setDirty(false);
              }}
              disabled={!dirty}
            >
              Reset
            </button>
            <button type="button" onClick={onSave} disabled={!dirty || isSaving}>
              Save
            </button>
          </div>
        </div>
      )}
    </section>
  );
};
