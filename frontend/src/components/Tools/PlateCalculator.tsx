import React, { useState } from 'react';
import { apiClient } from '@/utils/api';
import { useAuthStore } from '@/store/authStore';
import { generateUUID } from '@/utils/uuid';
import {
  Unit,
  PlateInventoryKg,
  PlateInventoryLb,
  PlateCount,
  BAR_OPTIONS,
  DEFAULT_INVENTORY_KG,
  DEFAULT_INVENTORY_LB,
  KG_PLATE_COLORS,
  LB_PLATE_COLOR,
} from './plateCalculatorTypes';
import { calculatePlates, convertBarWeight } from './plateCalculatorUtils';
import { BarVisualization } from './BarVisualization';
import { PlateInventoryEditor } from './PlateInventoryEditor';

export const PlateCalculator: React.FC = () => {
  const { user } = useAuthStore();
  const [unit, setUnit] = useState<Unit>('lb');
  const [targetWeight, setTargetWeight] = useState<number>(0);
  const [barWeight, setBarWeight] = useState<number>(45);
  const [inventoryKg, setInventoryKg] = useState<PlateInventoryKg>(DEFAULT_INVENTORY_KG);
  const [inventoryLb, setInventoryLb] = useState<PlateInventoryLb>(DEFAULT_INVENTORY_LB);
  const [result, setResult] = useState<PlateCount[]>([]);
  const [error, setError] = useState<string | null>(null);

  const inventory = unit === 'kg' ? inventoryKg : inventoryLb;

  const handleCalculate = () => {
    const { plates, error: calcError } = calculatePlates(targetWeight, barWeight, inventory, unit);
    setResult(plates);
    setError(calcError);

    if (user && plates.length > 0) {
      const event = {
        schema_version: '1.0.0',
        event_type: 'tools.platecalc.query',
        client_generated_id: generateUUID(),
        user_id: user.id,
        timestamp: new Date().toISOString(),
        source_service: 'frontend',
        data: {
          target_weight: targetWeight,
          bar_weight: barWeight,
          unit: unit,
          result_plates: plates,
        },
      };

      apiClient.submitEvent(event).catch((err) => {
        console.error('Failed to log plate calc query', err);
      });
    }
  };

  const updateInventory = (weight: string, count: number) => {
    if (unit === 'kg') {
      setInventoryKg({
        ...inventoryKg,
        [weight]: Math.max(0, count),
      });
    } else {
      setInventoryLb({
        ...inventoryLb,
        [weight]: Math.max(0, count),
      });
    }
  };

  const switchUnit = (newUnit: Unit) => {
    setUnit(newUnit);
    setResult([]);
    setError(null);
    setBarWeight(convertBarWeight(barWeight, unit, newUnit));
  };

  return (
    <div className="max-w-4xl mx-auto p-6">
      <div className="bg-white shadow rounded-lg p-6">
        <div className="flex justify-between items-center mb-6">
          <h2 className="text-2xl font-bold text-gray-900">Plate Calculator</h2>

          <div className="flex bg-gray-200 rounded-lg p-1">
            <button
              onClick={() => switchUnit('lb')}
              className={`px-4 py-2 rounded-md font-semibold transition-colors ${
                unit === 'lb'
                  ? 'bg-blue-600 text-white'
                  : 'text-gray-700 hover:bg-gray-300'
              }`}
            >
              LB
            </button>
            <button
              onClick={() => switchUnit('kg')}
              className={`px-4 py-2 rounded-md font-semibold transition-colors ${
                unit === 'kg'
                  ? 'bg-blue-600 text-white'
                  : 'text-gray-700 hover:bg-gray-300'
              }`}
            >
              KG
            </button>
          </div>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-2 gap-6 mb-6">
          <div>
            <h3 className="text-lg font-semibold text-gray-900 mb-4">Target Weight</h3>
            <div className="space-y-4">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-2">
                  Total Weight ({unit})
                </label>
                <input
                  type="number"
                  step={unit === 'kg' ? '0.5' : '2.5'}
                  value={targetWeight || ''}
                  onChange={(e) => setTargetWeight(parseFloat(e.target.value) || 0)}
                  className="w-full px-3 py-2 border border-gray-300 rounded-md"
                  placeholder={unit === 'kg' ? 'e.g., 140' : 'e.g., 315'}
                />
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 mb-2">
                  Bar Weight ({unit})
                </label>
                <select
                  value={barWeight}
                  onChange={(e) => setBarWeight(parseFloat(e.target.value))}
                  className="w-full px-3 py-2 border border-gray-300 rounded-md"
                >
                  {BAR_OPTIONS.map((bar) => (
                    <option key={bar.label} value={bar[unit]}>
                      {bar[unit]}
                      {unit} ({bar.label})
                    </option>
                  ))}
                </select>
              </div>

              <button
                onClick={handleCalculate}
                className="w-full px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700"
              >
                Calculate
              </button>
            </div>
          </div>

          <PlateInventoryEditor
            unit={unit}
            inventory={inventory}
            onUpdateInventory={updateInventory}
          />
        </div>

        {error && (
          <div className="mb-6 p-3 bg-yellow-50 border border-yellow-200 rounded-md">
            <p className="text-sm text-yellow-800">{error}</p>
          </div>
        )}

        {result.length > 0 && (
          <div className="bg-gray-50 rounded-lg p-6">
            <h3 className="text-lg font-semibold text-gray-900 mb-4">
              Plates Per Side
            </h3>
            <div className="space-y-3">
              {result.map((plate, idx) => {
                const color = unit === 'kg'
                  ? KG_PLATE_COLORS[plate.weight.toString()] || '#95A5A6'
                  : LB_PLATE_COLOR;

                return (
                  <div
                    key={idx}
                    className="flex items-center justify-between bg-white rounded-lg p-4 shadow-sm"
                  >
                    <div className="flex items-center gap-3">
                      <div
                        className="w-8 h-8 rounded border-2 border-gray-700"
                        style={{ backgroundColor: color }}
                      />
                      <span className="text-2xl font-bold text-blue-600">
                        {plate.weight}{unit}
                      </span>
                    </div>
                    <span className="text-lg text-gray-700">
                      x <span className="font-semibold">{plate.count}</span>
                    </span>
                  </div>
                );
              })}
            </div>

            <BarVisualization plates={result} unit={unit} barWeight={barWeight} />

            <div className="mt-6 pt-4 border-t border-gray-200">
              <div className="flex justify-between items-center">
                <span className="text-sm text-gray-600">Total per side:</span>
                <span className="text-lg font-semibold text-gray-900">
                  {((targetWeight - barWeight) / 2).toFixed(2)}{unit}
                </span>
              </div>
              <div className="flex justify-between items-center mt-2">
                <span className="text-sm text-gray-600">Total weight:</span>
                <span className="text-xl font-bold text-blue-600">{targetWeight}{unit}</span>
              </div>
            </div>
          </div>
        )}
      </div>

      <div className="bg-white shadow rounded-lg p-6 mt-6">
        <h2 className="text-2xl font-bold text-gray-900">Strength Check Me</h2>
        <p className="text-gray-600 mt-2">Coming soon...</p>
      </div>
    </div>
  );
};
