import { PlateCount, PlateInventoryKg, PlateInventoryLb, Unit, BAR_OPTIONS } from './plateCalculatorTypes';

export interface CalculationResult {
  plates: PlateCount[];
  error: string | null;
}

export const calculatePlates = (
  targetWeight: number,
  barWeight: number,
  inventory: PlateInventoryKg | PlateInventoryLb,
  unit: Unit
): CalculationResult => {
  if (targetWeight < barWeight) {
    return {
      plates: [],
      error: 'Target weight must be greater than bar weight',
    };
  }

  const weightPerSide = (targetWeight - barWeight) / 2;

  if (weightPerSide < 0) {
    return {
      plates: [],
      error: 'Invalid weight calculation',
    };
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

  let error: string | null = null;
  if (remaining > 0.01) {
    error = `Cannot make exact weight. ${remaining.toFixed(2)}${unit} remaining per side.`;
  }

  return { plates, error };
};

export const convertBarWeight = (currentBarWeight: number, fromUnit: Unit, toUnit: Unit): number => {
  const currentBar = BAR_OPTIONS.find((bar) => bar[fromUnit] === currentBarWeight);

  if (currentBar) {
    return currentBar[toUnit];
  } else {
    const conversionFactor = toUnit === 'kg' ? 0.453592 : 2.20462;
    return parseFloat((currentBarWeight * conversionFactor).toFixed(2));
  }
};
