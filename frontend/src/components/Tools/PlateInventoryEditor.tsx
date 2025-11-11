import React from 'react';
import { PlateInventoryKg, PlateInventoryLb, Unit, KG_PLATE_COLORS, LB_PLATE_COLOR } from './plateCalculatorTypes';

interface PlateInventoryEditorProps {
  unit: Unit;
  inventory: PlateInventoryKg | PlateInventoryLb;
  onUpdateInventory: (weight: string, count: number) => void;
}

export const PlateInventoryEditor: React.FC<PlateInventoryEditorProps> = ({
  unit,
  inventory,
  onUpdateInventory,
}) => {
  return (
    <div>
      <h3 className="text-lg font-semibold text-gray-900 mb-4">
        Plate Inventory ({unit === 'kg' ? 'Competition Colors' : 'Standard Plates'})
      </h3>
      {unit === 'kg' && (
        <p className="text-sm text-gray-600 mb-4">
          Note: Olympic clips are usually 2.5kg each.
        </p>
      )}
      <div className="space-y-2">
        {Object.entries(inventory).map(([weight, count]) => {
          const color = unit === 'kg' ? KG_PLATE_COLORS[weight] : LB_PLATE_COLOR;
          return (
            <div key={weight} className="flex items-center justify-between">
              <div className="flex items-center gap-2">
                <div
                  className="w-6 h-6 rounded border-2 border-gray-700"
                  style={{ backgroundColor: color }}
                />
                <label className="text-sm font-medium text-gray-700">
                  {weight}{unit}
                </label>
              </div>
              <input
                type="number"
                min="0"
                value={count}
                onChange={(e) =>
                  onUpdateInventory(weight, parseInt(e.target.value) || 0)
                }
                className="w-20 px-2 py-1 border border-gray-300 rounded text-sm"
              />
            </div>
          );
        })}
      </div>
    </div>
  );
};
