import React from 'react'
import { useEffect, useState } from 'react'

export const Health: React.FC = () => {
  const [ok, setOk] = useState<boolean | null>(null)

  useEffect(() => {
    fetch('/api/health')
      .then((r) => r.json())
      .then((d) => setOk(Boolean(d?.ok)))
      .catch(() => setOk(false))
  }, [])

  return (
    <span style={{ fontSize: 12, color: ok ? 'green' : 'red' }}>
      {ok === null ? '…' : ok ? 'Server OK' : 'Server DOWN'}
    </span>
  )
}

