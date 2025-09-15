import type React from 'react';
import { Link } from 'react-router-dom';
import { Card, CardBody, Button } from '../components/ui';

export const Home: React.FC = () => {
  return (
    <div className="max-w-4xl mx-auto">
      {/* Hero Section */}
      <div className="text-center mb-12">
        <div className="w-16 h-16 bg-primary-600 rounded-2xl flex items-center justify-center mx-auto mb-6">
          <span className="text-white font-bold text-2xl">Z</span>
        </div>
        <h1 className="text-4xl font-bold text-gray-900 mb-4">
          Welcome to Zine Layout
        </h1>
        <p className="text-xl text-gray-600 mb-8 max-w-2xl mx-auto">
          Create beautiful multi-page zines from your images with our intuitive web interface. 
          Configure layouts, manage images, and generate print-ready pages.
        </p>
        <div className="flex justify-center space-x-4">
          <Link to="/projects">
            <Button size="lg">Get Started</Button>
          </Link>
          <a 
            href="https://github.com/your-repo/zine-layout" 
            target="_blank" 
            rel="noopener noreferrer"
          >
            <Button variant="secondary" size="lg">Learn More</Button>
          </a>
        </div>
      </div>

      {/* Features Grid */}
      <div className="grid md:grid-cols-2 lg:grid-cols-3 gap-6 mb-12">
        <Card>
          <CardBody>
            <div className="w-12 h-12 bg-primary-100 rounded-lg flex items-center justify-center mb-4">
              <svg className="w-6 h-6 text-primary-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z" />
              </svg>
            </div>
            <h3 className="text-lg font-semibold text-gray-900 mb-2">Image Management</h3>
            <p className="text-gray-600">
              Upload, organize, and manage your images with drag-and-drop functionality and automatic sizing validation.
            </p>
          </CardBody>
        </Card>

        <Card>
          <CardBody>
            <div className="w-12 h-12 bg-primary-100 rounded-lg flex items-center justify-center mb-4">
              <svg className="w-6 h-6 text-primary-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10" />
              </svg>
            </div>
            <h3 className="text-lg font-semibold text-gray-900 mb-2">Layout Presets</h3>
            <p className="text-gray-600">
              Choose from preset layouts like 2-up, 4-up, 8-sheet zine, and 16-sheet zine, or create custom grids.
            </p>
          </CardBody>
        </Card>

        <Card>
          <CardBody>
            <div className="w-12 h-12 bg-primary-100 rounded-lg flex items-center justify-center mb-4">
              <svg className="w-6 h-6 text-primary-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 10v6m0 0l-3-3m3 3l3-3M3 17V7a2 2 0 012-2h6l2 2h6a2 2 0 012 2v8a2 2 0 01-2 2H5a2 2 0 01-2-2z" />
              </svg>
            </div>
            <h3 className="text-lg font-semibold text-gray-900 mb-2">Export Ready</h3>
            <p className="text-gray-600">
              Generate high-quality PNG outputs with customizable PPI, margins, and borders ready for printing.
            </p>
          </CardBody>
        </Card>
      </div>

      {/* Quick Actions */}
      <Card>
        <CardBody>
          <div className="text-center">
            <h2 className="text-2xl font-bold text-gray-900 mb-4">Ready to create your first zine?</h2>
            <p className="text-gray-600 mb-6">
              Start by creating a new project or exploring existing projects.
            </p>
            <div className="flex justify-center space-x-4">
              <Link to="/projects">
                <Button>View Projects</Button>
              </Link>
            </div>
          </div>
        </CardBody>
      </Card>
    </div>
  );
};
