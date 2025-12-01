import type React from 'react';
import { clsx } from 'clsx';

interface CardProps extends React.HTMLAttributes<HTMLDivElement> {
  className?: string;
  children: React.ReactNode;
}

export const Card: React.FC<CardProps> = ({ className, children, ...props }) => (
  <div className={clsx('card', className)} {...props}>
    {children}
  </div>
);

export const CardHeader: React.FC<CardProps> = ({ className, children, ...props }) => (
  <div className={clsx('card-header', className)} {...props}>
    {children}
  </div>
);

export const CardBody: React.FC<CardProps> = ({ className, children, ...props }) => (
  <div className={clsx('card-body', className)} {...props}>
    {children}
  </div>
);

export const CardFooter: React.FC<CardProps> = ({ className, children, ...props }) => (
  <div className={clsx('card-footer', className)} {...props}>
    {children}
  </div>
);
