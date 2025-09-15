import type React from 'react';
import { useLazyValidateProjectQuery } from '../api';

export const ProjectValidationPanel: React.FC<{ id: string }> = ({ id }) => {
  const [trigger, { data, isFetching, isUninitialized }] = useLazyValidateProjectQuery();

  const onRun = () => {
    trigger({ id });
  };

  return (
    <section>
      <h2>Validation</h2>
      <div style={{ display: 'flex', gap: 8, alignItems: 'center' }}>
        <button type="button" onClick={onRun} disabled={isFetching}>
          Run validation
        </button>
        {!isUninitialized && !isFetching && data && (
          <span style={{ color: data.ok ? 'green' : 'red' }}>
            {data.ok ? 'OK' : 'Issues found'}
          </span>
        )}
      </div>
      {!isUninitialized && data && (
        <div style={{ marginTop: 8 }}>
          {data.issues?.length ? (
            <ul>
              {data.issues.map((iss) => (
                <li key={iss}>{iss}</li>
              ))}
            </ul>
          ) : (
            <p>No issues.</p>
          )}
          {data.details && (
            <pre style={{ background: '#f7f7f7', padding: 8, overflow: 'auto' }}>
              {JSON.stringify(data.details, null, 2)}
            </pre>
          )}
        </div>
      )}
    </section>
  );
};
