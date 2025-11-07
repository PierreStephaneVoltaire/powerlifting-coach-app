import React, { useState } from 'react';
import { apiClient } from '@/utils/api';
import { useAuthStore } from '@/store/authStore';

import { generateUUID } from '@/utils/uuid';
interface PlateInventory {
  '25kg': number;
  '20kg': number;
  '15kg': number;
  '10kg': number;
  '5kg': number;
  '2.5kg': number;
  '1.25kg': number;
  '0.5kg': number;
}

interface PlateCount {
  weight: string;
  count: number;
}

const DEFAULT_INVENTORY: PlateInventory = {
  '25kg': 4,
  '20kg': 4,
  '15kg': 2,
  '10kg': 4,
  '5kg': 4,
  '2.5kg': 4,
  '1.25kg': 2,
  '0.5kg': 2,
};

export const PlateCalculator: React.FC = () => {
  const { user } = useAuthStore();
  const [targetWeight, setTargetWeight] = useState<number>(0);
  const [barWeight, setBarWeight] = useState<number>(20);
  const [inventory, setInventory] = useState<PlateInventory>(DEFAULT_INVENTORY);
  const [result, setResult] = useState<PlateCount[]>([]);
  const [error, setError] = useState<string | null>(null);

  const calculatePlates = () => {
    setError(null);

    if (targetWeight < barWeight) {
      setError('Target weight must be greater than bar weight');
      setResult([]);
      return;
    }

    const weightPerSide = (targetWeight - barWeight) / 2;

    if (weightPerSide < 0) {
      setError('Invalid weight calculation');
      setResult([]);
      return;
    }

    const availablePlates = Object.entries(inventory)
      .map(([weight, count]) => ({
        weight: parseFloat(weight.replace('kg', '')),
        count,
      }))
      .sort((a, b) => b.weight - a.weight);

    let remaining = weightPerSide;
    const plates: PlateCount[] = [];

    for (const plate of availablePlates) {
      if (remaining >= plate.weight && plate.count > 0) {
        const neededCount = Math.min(
          Math.floor(remaining / plate.weight),
          plate.count
        );
        if (neededCount > 0) {
          plates.push({
            weight: `${plate.weight}kg`,
            count: neededCount,
          });
          remaining -= neededCount * plate.weight;
        }
      }
    }

    if (remaining > 0.01) {
      setError(`Cannot make exact weight. ${remaining.toFixed(2)}kg remaining per side.`);
    }

    setResult(plates);

    if (user) {
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
          result_plates: plates,
        },
      };

      apiClient.submitEvent(event).catch((err) => {
        console.error('Failed to log plate calc query', err);
      });
    }
  };

  const updateInventory = (weight: keyof PlateInventory, count: number) => {
    setInventory({
      ...inventory,
      [weight]: Math.max(0, count),
    });
  };

  return (
    <div className="max-w-4xl mx-auto p-6">
      <div className="bg-white shadow rounded-lg p-6">
        <h2 className="text-2xl font-bold text-gray-900 mb-6">Plate Calculator</h2>

        <div className="grid grid-cols-1 md:grid-cols-2 gap-6 mb-6">
          <div>
            <h3 className="text-lg font-semibold text-gray-900 mb-4">Target Weight</h3>
            <div className="space-y-4">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-2">
                  Total Weight (kg)
                </label>
                <input
                  type="number"
                  step="0.5"
                  value={targetWeight || ''}
                  onChange={(e) => setTargetWeight(parseFloat(e.target.value) || 0)}
                  className="w-full px-3 py-2 border border-gray-300 rounded-md"
                  placeholder="e.g., 140"
                />
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 mb-2">
                  Bar Weight (kg)
                </label>
                <select
                  value={barWeight}
                  onChange={(e) => setBarWeight(parseFloat(e.target.value))}
                  className="w-full px-3 py-2 border border-gray-300 rounded-md"
                >
                  <option value="20">20kg (Standard Barbell)</option>
                  <option value="15">15kg (Women's Barbell)</option>
                  <option value="10">10kg (Training Bar)</option>
                  <option value="5">5kg (Technique Bar)</option>
                </select>
              </div>

              <button
                onClick={calculatePlates}
                className="w-full px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700"
              >
                Calculate
              </button>
            </div>
          </div>

          <div>
            <h3 className="text-lg font-semibold text-gray-900 mb-4">Plate Inventory</h3>
            <div className="space-y-2">
              {Object.entries(inventory).map(([weight, count]) => (
                <div key={weight} className="flex items-center justify-between">
                  <label className="text-sm font-medium text-gray-700">{weight}</label>
                  <input
                    type="number"
                    min="0"
                    value={count}
                    onChange={(e) =>
                      updateInventory(weight as keyof PlateInventory, parseInt(e.target.value) || 0)
                    }
                    className="w-20 px-2 py-1 border border-gray-300 rounded text-sm"
                  />
                </div>
              ))}
            </div>
          </div>
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
              {result.map((plate, idx) => (
                <div
                  key={idx}
                  className="flex items-center justify-between bg-white rounded-lg p-4 shadow-sm"
                >
                  <span className="text-2xl font-bold text-blue-600">{plate.weight}</span>
                  <span className="text-lg text-gray-700">
                    x <span className="font-semibold">{plate.count}</span>
                  </span>
                </div>
              ))}
            </div>
            <div className="mt-4 pt-4 border-t border-gray-200">
              <div className="flex justify-between items-center">
                <span className="text-sm text-gray-600">Total per side:</span>
                <span className="text-lg font-semibold text-gray-900">
                  {((targetWeight - barWeight) / 2).toFixed(2)}kg
                </span>
              </div>
              <div className="flex justify-between items-center mt-2">
                <span className="text-sm text-gray-600">Total weight:</span>
                <span className="text-xl font-bold text-blue-600">{targetWeight}kg</span>
              </div>
            </div>
          </div>
        )}
      </div>
    </div>
  );
};
