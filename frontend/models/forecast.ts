export interface ForecastResponse {
    data: {
        month: string;
        revenue: number;
        expense: number;
        cashflow: number;
    }
    updatedAt: string;
}

export interface ForecastDetailRevenueExpenseResponse {
    name: string
    amount: number
}

export interface ForecastDetailResponse {
    month: string;
    revenue: ForecastDetailRevenueExpenseResponse[];
    expense: ForecastDetailRevenueExpenseResponse[];
    forecastID: number;
}
