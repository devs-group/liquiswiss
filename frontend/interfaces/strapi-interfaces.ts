export interface Strapi_ListResponse {
    data: any[]
    meta: {
        pagination: {
            page: number;
            pageSize: number;
            pageCount: number;
            total: number;
        }
    }
}
