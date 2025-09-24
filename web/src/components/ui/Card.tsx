import type React from 'react';
import { clsx } from 'clsx';

interface CardProps {
  className?: string;
  children: React.ReactNode;
}

export const Card: React.FC<CardProps> = ({ className, children }) => (
  <div className={clsx('card', className)}>
    {children}
  </div>
);

export const CardHeader: React.FC<CardProps> = ({ className, children }) => (
  <div className={clsx('card-header', className)}>
    {children}
  </div>
);

export const CardBody: React.FC<CardProps> = ({ className, children }) => (
  <div className={clsx('card-body', className)}>
    {children}
  </div>
);

export const CardFooter: React.FC<CardProps> = ({ className, children }) => (
  <div className={clsx('card-footer', className)}>
    {children}
  </div>
);
