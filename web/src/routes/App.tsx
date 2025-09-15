import type React from 'react';
import { BrowserRouter, Link, Route, Routes } from 'react-router-dom';
import { Health } from '../views/Health';
import { Home } from '../views/Home';
import { ProjectDetail } from '../views/ProjectDetail';
import { ProjectYamlPage } from '../views/ProjectYamlPage';
import { Projects } from '../views/Projects';

export const App: React.FC = () => {
  return (
    <BrowserRouter>
      <div style={{ padding: 16, fontFamily: 'system-ui, sans-serif' }}>
        <header style={{ display: 'flex', gap: 12, alignItems: 'center' }}>
          <Link to="/">Zine Layout</Link>
          <Link to="/projects">Projects</Link>
          <div style={{ marginLeft: 'auto' }}>
            <Health />
          </div>
        </header>
        <Routes>
          <Route path="/" element={<Home />} />
          <Route path="/projects" element={<Projects />} />
          <Route path="/projects/:id" element={<ProjectDetail />} />
          <Route path="/projects/:id/yaml" element={<ProjectYamlPage />} />
        </Routes>
      </div>
    </BrowserRouter>
  );
};
