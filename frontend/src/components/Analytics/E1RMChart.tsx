import React from 'react';
import {
  LineChart,
  Line,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  Legend,
  ResponsiveContainer,
} from 'recharts';
import { format } from 'date-fns';
import { useTheme } from '@/context/ThemeContext';

interface E1RMChartProps {
  data: any[];
}

interface E1RMDataPoint {
  date: string;
  exercise_name: string;
  lift_type: string;
  estimated_1rm: number;
  weight_used: number;
  reps_achieved: number;
}

export const E1RMChart: React.FC<E1RMChartProps> = ({ data }) => {
  const { theme } = useTheme();
  // Group by exercise and date, taking the max e1RM for each day
  const groupedData = data.reduce((acc: Record<string, E1RMDataPoint>, curr) => {
    const dateStr = format(new Date(curr.date), 'MMM dd');
    const key = `${dateStr}_${curr.exercise_name}`;

    if (!acc[key] || acc[key].estimated_1rm < curr.estimated_1rm) {
      acc[key] = {
        date: dateStr,
        exercise_name: curr.exercise_name,
        lift_type: curr.lift_type,
        estimated_1rm: curr.estimated_1rm,
        weight_used: curr.weight_used,
        reps_achieved: curr.reps_achieved,
      };
    }

    return acc;
  }, {});

  const chartData: E1RMDataPoint[] = Object.values(groupedData);

  // Separate by lift type
  const squatData = chartData.filter((item) => item.lift_type === 'squat');
  const benchData = chartData.filter((item) => item.lift_type === 'bench');
  const deadliftData = chartData.filter((item) => item.lift_type === 'deadlift');

  // Combine all data with separate fields
  const combinedData = Array.from(
    new Set([...squatData, ...benchData, ...deadliftData].map((item) => item.date))
  ).map(date => {
    const squatEntry = squatData.find((item) => item.date === date);
    const benchEntry = benchData.find((item) => item.date === date);
    const deadliftEntry = deadliftData.find((item) => item.date === date);

    return {
      date,
      squat: squatEntry?.estimated_1rm,
      bench: benchEntry?.estimated_1rm,
      deadlift: deadliftEntry?.estimated_1rm,
    };
  }).sort((a, b) => {
    const dateA = new Date(a.date);
    const dateB = new Date(b.date);
    return dateA.getTime() - dateB.getTime();
  });

  if (combinedData.length === 0) {
    return (
      <div className="flex items-center justify-center h-64 text-gray-500 dark:text-gray-400">
        No e1RM data available for the selected period
      </div>
    );
  }

  const isDark = theme === 'dark';
  const axisColor = isDark ? '#9ca3af' : '#6b7280';
  const gridColor = isDark ? '#374151' : '#e5e7eb';

  return (
    <ResponsiveContainer width="100%" height={400}>
      <LineChart data={combinedData}>
        <CartesianGrid strokeDasharray="3 3" stroke={gridColor} />
        <XAxis
          dataKey="date"
          tick={{ fontSize: 12, fill: axisColor }}
          angle={-45}
          textAnchor="end"
          height={80}
          stroke={axisColor}
        />
        <YAxis
          label={{ value: 'Estimated 1RM (kg)', angle: -90, position: 'insideLeft', fill: axisColor }}
          tick={{ fontSize: 12, fill: axisColor }}
          stroke={axisColor}
        />
        <Tooltip
          contentStyle={{
            backgroundColor: isDark ? '#1f2937' : '#fff',
            border: `1px solid ${isDark ? '#374151' : '#ccc'}`,
            color: isDark ? '#fff' : '#000',
          }}
          formatter={(value: any) => {
            if (value === undefined || value === null) return 'N/A';
            return `${Number(value).toFixed(1)} kg`;
          }}
        />
        <Legend wrapperStyle={{ color: axisColor }} />
        <Line
          type="monotone"
          dataKey="squat"
          stroke="#ef4444"
          strokeWidth={2}
          dot={{ r: 4 }}
          connectNulls
          name="Squat e1RM"
        />
        <Line
          type="monotone"
          dataKey="bench"
          stroke="#3b82f6"
          strokeWidth={2}
          dot={{ r: 4 }}
          connectNulls
          name="Bench e1RM"
        />
        <Line
          type="monotone"
          dataKey="deadlift"
          stroke="#10b981"
          strokeWidth={2}
          dot={{ r: 4 }}
          connectNulls
          name="Deadlift e1RM"
        />
      </LineChart>
    </ResponsiveContainer>
  );
};
