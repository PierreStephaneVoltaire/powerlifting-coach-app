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

interface VolumeChartProps {
  data: any[];
}

export const VolumeChart: React.FC<VolumeChartProps> = ({ data }) => {
  const { theme } = useTheme();
  // Aggregate data by date
  const aggregatedData = data.reduce((acc: any[], curr) => {
    const dateStr = format(new Date(curr.date), 'MMM dd');
    const existing = acc.find(item => item.date === dateStr);

    if (existing) {
      existing.volume += curr.total_volume;
      existing.sets += curr.total_sets;
      existing.reps += curr.total_reps;
    } else {
      acc.push({
        date: dateStr,
        volume: curr.total_volume,
        sets: curr.total_sets,
        reps: curr.total_reps,
      });
    }

    return acc;
  }, []);

  // Sort by date
  aggregatedData.sort((a, b) => {
    const dateA = new Date(a.date);
    const dateB = new Date(b.date);
    return dateA.getTime() - dateB.getTime();
  });

  if (aggregatedData.length === 0) {
    return (
      <div className="flex items-center justify-center h-64 text-gray-500 dark:text-gray-400">
        No volume data available for the selected period
      </div>
    );
  }

  const isDark = theme === 'dark';
  const axisColor = isDark ? '#9ca3af' : '#6b7280';
  const gridColor = isDark ? '#374151' : '#e5e7eb';

  return (
    <ResponsiveContainer width="100%" height={400}>
      <LineChart data={aggregatedData}>
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
          label={{ value: 'Volume (kg)', angle: -90, position: 'insideLeft', fill: axisColor }}
          tick={{ fontSize: 12, fill: axisColor }}
          stroke={axisColor}
        />
        <Tooltip
          contentStyle={{
            backgroundColor: isDark ? '#1f2937' : '#fff',
            border: `1px solid ${isDark ? '#374151' : '#ccc'}`,
            color: isDark ? '#fff' : '#000',
          }}
          formatter={(value: number) => [`${value.toFixed(0)} kg`, 'Volume']}
        />
        <Legend wrapperStyle={{ color: axisColor }} />
        <Line
          type="monotone"
          dataKey="volume"
          stroke="#3b82f6"
          strokeWidth={2}
          dot={{ r: 4 }}
          activeDot={{ r: 6 }}
          name="Total Volume"
        />
      </LineChart>
    </ResponsiveContainer>
  );
};
