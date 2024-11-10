export interface ForecastResponse {
    month: string;
    revenue: number;
    expense: number;
    cashflow: number;
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
