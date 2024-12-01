import type {ChartOptions, ChartData} from "chart.js";
import resolveConfig from "tailwindcss/resolveConfig";
import tailwindConfig from '~/tailwind.config';

export default function useCharts() {
    const {isDarkMode} = useDarkMode()

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

    const getChartOptions = () => {
        const config = resolveConfig(tailwindConfig)

        const textColor = isDarkMode.value ? config.theme.colors.white : config.theme.colors.black;
        const textColorSecondary = isDarkMode.value ? config.theme.colors.white : config.theme.colors.black;
        const surfaceBorder = isDarkMode.value ? config.theme.colors.zinc["700"] : config.theme.colors.zinc["100"];

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
                        color: ({ tick }) => tick.value == 0 ? config.theme.colors.red["600"] : isDarkMode.value ? config.theme.colors.zinc["700"] : config.theme.colors.zinc["100"],
                        lineWidth: ({ tick }) => tick.value == 0 ? 3 : 1
                    },
                    beginAtZero: true,
                }
            }
        } as ChartOptions;
    }


    return {
        setChartData,
        getChartOptions,
    };
}
