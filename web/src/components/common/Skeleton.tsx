import type { SkeletonProps } from '../../types/components';
import './Skeleton.scss';

export function Skeleton({ className = '' }: SkeletonProps) {
  return <div className={`skeleton ${className}`} />;
}
