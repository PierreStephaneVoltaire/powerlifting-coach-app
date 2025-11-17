import React from 'react';
import { PlateCount, Unit, KG_PLATE_COLORS, LB_PLATE_COLOR } from './plateCalculatorTypes';

interface BarVisualizationProps {
  plates: PlateCount[];
  unit: Unit;
  barWeight: number;
}

export const BarVisualization: React.FC<BarVisualizationProps> = ({ plates, unit, barWeight }) => {
  const sortedPlates = [...plates].sort((a, b) => b.weight - a.weight);

  const renderPlates = (key: string) => {
    return sortedPlates.map((plate, idx) =>
      Array.from({ length: plate.count }).map((_, i) => {
        const color = unit === 'kg'
          ? KG_PLATE_COLORS[plate.weight.toString()] || '#95A5A6'
          : LB_PLATE_COLOR;
        const height = Math.min(20 + (plate.weight / (unit === 'kg' ? 25 : 45)) * 40, 60);

        return (
          <div
            key={`${key}-${idx}-${i}`}
            className="flex flex-col items-center"
          >
            <div
              className="rounded flex items-center justify-center text-white font-bold text-xs shadow-md border-2 border-gray-700"
              style={{
                backgroundColor: color,
                width: '24px',
                height: `${height}px`,
                minHeight: '32px',
              }}
            >
              {plate.weight}
            </div>
          </div>
        );
      })
    ).flat();
  };

  return (
    <div className="mt-6">
      <h3 className="text-lg font-semibold text-gray-900 mb-4 text-center">
        Bar Visualization
      </h3>
      <div className="flex items-center justify-center gap-2 overflow-x-auto pb-4">
        <div className="w-8 h-12 bg-gray-400 rounded-l-md flex-shrink-0" />

        <div className="flex gap-1">
          {renderPlates('left')}
        </div>

        <div className="h-4 bg-gray-500 rounded flex-shrink-0 flex items-center justify-center px-4 min-w-[80px]">
          <span className="text-xs font-semibold text-white">{barWeight}{unit}</span>
        </div>

        <div className="flex gap-1 flex-row-reverse">
          {renderPlates('right')}
        </div>

        <div className="w-8 h-12 bg-gray-400 rounded-r-md flex-shrink-0" />
      </div>
    </div>
  );
};
