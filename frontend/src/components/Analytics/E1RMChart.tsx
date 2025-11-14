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

interface E1RMChartProps {
  data: any[];
}

export const E1RMChart: React.FC<E1RMChartProps> = ({ data }) => {
  // Group by exercise and date, taking the max e1RM for each day
  const groupedData = data.reduce((acc: any, curr) => {
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

  const chartData = Object.values(groupedData);

  // Separate by lift type
  const squatData = chartData.filter((item: any) => item.lift_type === 'squat');
  const benchData = chartData.filter((item: any) => item.lift_type === 'bench');
  const deadliftData = chartData.filter((item: any) => item.lift_type === 'deadlift');

  // Combine all data with separate fields
  const combinedData = Array.from(
    new Set([...squatData, ...benchData, ...deadliftData].map((item: any) => item.date))
  ).map(date => {
    const squatEntry = squatData.find((item: any) => item.date === date);
    const benchEntry = benchData.find((item: any) => item.date === date);
    const deadliftEntry = deadliftData.find((item: any) => item.date === date);

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
      <div className="flex items-center justify-center h-64 text-gray-500">
        No e1RM data available for the selected period
      </div>
    );
  }

  return (
    <ResponsiveContainer width="100%" height={400}>
      <LineChart data={combinedData}>
        <CartesianGrid strokeDasharray="3 3" />
        <XAxis
          dataKey="date"
          tick={{ fontSize: 12 }}
          angle={-45}
          textAnchor="end"
          height={80}
        />
        <YAxis
          label={{ value: 'Estimated 1RM (kg)', angle: -90, position: 'insideLeft' }}
          tick={{ fontSize: 12 }}
        />
        <Tooltip
          contentStyle={{ backgroundColor: '#fff', border: '1px solid #ccc' }}
          formatter={(value: number | undefined, name: string) => {
            if (value === undefined) return ['N/A', name];
            return [`${value.toFixed(1)} kg`, name];
          }}
        />
        <Legend />
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
