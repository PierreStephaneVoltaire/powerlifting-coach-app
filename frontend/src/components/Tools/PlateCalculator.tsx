import React, { useState } from 'react';
import { apiClient } from '@/utils/api';
import { useAuthStore } from '@/store/authStore';
import { generateUUID } from '@/utils/uuid';

type Unit = 'kg' | 'lb';

interface PlateInventoryKg {
  '25': number;
  '20': number;
  '15': number;
  '10': number;
  '5': number;
  '2.5': number;
  '1.25': number;
  '0.5': number;
}

interface PlateInventoryLb {
  '45': number;
  '35': number;
  '25': number;
  '10': number;
  '5': number;
  '2.5': number;
}

interface PlateCount {
  weight: number;
  count: number;
}

const DEFAULT_INVENTORY_KG: PlateInventoryKg = {
  '25': 4,
  '20': 4,
  '15': 2,
  '10': 4,
  '5': 4,
  '2.5': 4,
  '1.25': 2,
  '0.5': 2,
};

const DEFAULT_INVENTORY_LB: PlateInventoryLb = {
  '45': 4,
  '35': 2,
  '25': 4,
  '10': 4,
  '5': 4,
  '2.5': 4,
};

// Competition plate colors (IPF standard)
const KG_PLATE_COLORS: Record<string, string> = {
  '25': '#E74C3C',    // Red
  '20': '#3498DB',    // Blue
  '15': '#F1C40F',    // Yellow
  '10': '#27AE60',    // Green
  '5': '#ECF0F1',     // White
  '2.5': '#E74C3C',   // Red
  '1.25': '#3498DB',  // Blue
  '0.5': '#F1C40F',   // Yellow
};

// Standard gym plates (black/gray)
const LB_PLATE_COLOR = '#2C3E50'; // Dark gray/black

export const PlateCalculator: React.FC = () => {
  const { user } = useAuthStore();
  const [unit, setUnit] = useState<Unit>('lb');
  const [targetWeight, setTargetWeight] = useState<number>(0);
  const [barWeight, setBarWeight] = useState<number>(45); // Default to 45lb bar
  const [inventoryKg, setInventoryKg] = useState<PlateInventoryKg>(DEFAULT_INVENTORY_KG);
  const [inventoryLb, setInventoryLb] = useState<PlateInventoryLb>(DEFAULT_INVENTORY_LB);
  const [result, setResult] = useState<PlateCount[]>([]);
  const [error, setError] = useState<string | null>(null);

  const inventory = unit === 'kg' ? inventoryKg : inventoryLb;

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
        weight: parseFloat(weight),
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
            weight: plate.weight,
            count: neededCount,
          });
          remaining -= neededCount * plate.weight;
        }
      }
    }

    if (remaining > 0.01) {
      setError(`Cannot make exact weight. ${remaining.toFixed(2)}${unit} remaining per side.`);
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
    // Update bar weight to equivalent
    if (newUnit === 'kg') {
      if (barWeight === 45) setBarWeight(20);
      else if (barWeight === 35) setBarWeight(15);
      else if (barWeight === 33) setBarWeight(15);
    } else {
      if (barWeight === 20) setBarWeight(45);
      else if (barWeight === 15) setBarWeight(35);
    }
  };

  const PlateVisualization: React.FC<{ plates: PlateCount[] }> = ({ plates }) => {
    return (
      <div className="mt-6">
        <h3 className="text-lg font-semibold text-gray-900 mb-4 text-center">
          Bar Visualization
        </h3>
        <div className="flex items-center justify-center gap-2 overflow-x-auto pb-4">
          {/* Left sleeve */}
          <div className="w-8 h-12 bg-gray-400 rounded-l-md flex-shrink-0" />

          {/* Plates on left side */}
          <div className="flex gap-1">
            {plates.map((plate, idx) => (
              Array.from({ length: plate.count }).map((_, i) => {
                const color = unit === 'kg'
                  ? KG_PLATE_COLORS[plate.weight.toString()] || '#95A5A6'
                  : LB_PLATE_COLOR;
                const height = Math.min(20 + (plate.weight / (unit === 'kg' ? 25 : 45)) * 40, 60);

                return (
                  <div
                    key={`left-${idx}-${i}`}
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
            )).flat()}
          </div>

          {/* Bar */}
          <div className="h-4 bg-gray-500 rounded flex-shrink-0 flex items-center justify-center px-4 min-w-[80px]">
            <span className="text-xs font-semibold text-white">{barWeight}{unit}</span>
          </div>

          {/* Plates on right side */}
          <div className="flex gap-1 flex-row-reverse">
            {plates.map((plate, idx) => (
              Array.from({ length: plate.count }).map((_, i) => {
                const color = unit === 'kg'
                  ? KG_PLATE_COLORS[plate.weight.toString()] || '#95A5A6'
                  : LB_PLATE_COLOR;
                const height = Math.min(20 + (plate.weight / (unit === 'kg' ? 25 : 45)) * 40, 60);

                return (
                  <div
                    key={`right-${idx}-${i}`}
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
            )).flat()}
          </div>

          {/* Right sleeve */}
          <div className="w-8 h-12 bg-gray-400 rounded-r-md flex-shrink-0" />
        </div>
      </div>
    );
  };

  return (
    <div className="max-w-4xl mx-auto p-6">
      <div className="bg-white shadow rounded-lg p-6">
        <div className="flex justify-between items-center mb-6">
          <h2 className="text-2xl font-bold text-gray-900">Plate Calculator</h2>

          {/* Unit Toggle */}
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
                  {unit === 'kg' ? (
                    <>
                      <option value="20">20kg (Men's Bar)</option>
                      <option value="15">15kg (Women's Bar)</option>
                      <option value="10">10kg (Training Bar)</option>
                      <option value="5">5kg (Technique Bar)</option>
                    </>
                  ) : (
                    <>
                      <option value="45">45lb (Men's Bar)</option>
                      <option value="35">35lb (Women's Bar)</option>
                      <option value="33">33lb (15kg Training Bar)</option>
                      <option value="15">15lb (Technique Bar)</option>
                    </>
                  )}
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
            <h3 className="text-lg font-semibold text-gray-900 mb-4">
              Plate Inventory ({unit === 'kg' ? 'Competition Colors' : 'Standard Plates'})
            </h3>
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
                        updateInventory(weight, parseInt(e.target.value) || 0)
                      }
                      className="w-20 px-2 py-1 border border-gray-300 rounded text-sm"
                    />
                  </div>
                );
              })}
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

            <PlateVisualization plates={result} />

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
    </div>
  );
};
