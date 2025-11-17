export type Unit = 'kg' | 'lb';

export interface PlateInventoryKg {
  '25': number;
  '20': number;
  '15': number;
  '10': number;
  '5': number;
  '2.5': number;
  '1.25': number;
  '0.5': number;
}

export interface PlateInventoryLb {
  '45': number;
  '35': number;
  '25': number;
  '10': number;
  '5': number;
  '2.5': number;
}

export interface PlateCount {
  weight: number;
  count: number;
}

export const BAR_OPTIONS = [
  { lb: 45, kg: 20, label: "Men's Bar" },
  { lb: 35, kg: 15, label: "Women's Bar" },
  { lb: 55, kg: 25, label: 'Hex Bar' },
  { lb: 33, kg: 15, label: '15kg Training Bar' },
  { lb: 15, kg: 7, label: 'Technique Bar' },
  { lb: 0, kg: 0, label: 'Machine - Leg Press/Hack Squat' },
];

export const DEFAULT_INVENTORY_KG: PlateInventoryKg = {
  '25': 4,
  '20': 4,
  '15': 2,
  '10': 4,
  '5': 4,
  '2.5': 4,
  '1.25': 2,
  '0.5': 2,
};

export const DEFAULT_INVENTORY_LB: PlateInventoryLb = {
  '45': 4,
  '35': 2,
  '25': 4,
  '10': 4,
  '5': 4,
  '2.5': 4,
};

export const KG_PLATE_COLORS: Record<string, string> = {
  '25': '#E74C3C',
  '20': '#3498DB',
  '15': '#F1C40F',
  '10': '#27AE60',
  '5': '#ECF0F1',
  '2.5': '#E74C3C',
  '1.25': '#3498DB',
  '0.5': '#F1C40F',
};

export const LB_PLATE_COLOR = '#2C3E50';
