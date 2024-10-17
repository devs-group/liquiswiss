import type {ChartOptions, ChartData} from "chart.js";

export default function useCharts() {
    const setChartData = (months: string[], saldos: number[]) => {
        return {
            labels: months,
            datasets: [
                {
                    label: 'Saldo',
                    data: saldos,
                    fill: true,
                    borderColor: '#10b981',
                    tension: 0.2,
                },
            ]
        } as ChartData;
    };

    const setChartOptions = () => {
        const textColor = '#000000';
        const textColorSecondary = '#000000';
        const surfaceBorder = '#ffffff';

        return {
            maintainAspectRatio: false,
            aspectRatio: 0.6,
            plugins: {
                legend: {
                    labels: {
                        color: textColor
                    }
                }
            },
            scales: {
                x: {
                    ticks: {
                        color: textColorSecondary
                    },
                    grid: {
                        color: surfaceBorder
                    }
                },
                y: {
                    ticks: {
                        color: textColorSecondary,
                    },
                    grid: {
                        color: ({ tick }) => tick.value == 0 ? "#ff0000" : "#eee",
                        lineWidth: ({ tick }) => tick.value == 0 ? 2 : 1
                    },
                    beginAtZero: true,
                }
            }
        } as ChartOptions;
    }


    return {
        setChartData,
        setChartOptions,
    };
}
