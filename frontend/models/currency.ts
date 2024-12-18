import type { Strapi_Meta } from '~/interfaces/strapi-interfaces'
import type { PaginationResponse } from '~/models/pagination'

export interface StrapiCurrency {
  id?: number
  attributes: {
    code: string
    description: string
    localeCode: string
    createdAt: string
    updatedAt: string
    publishedAt: string
  }
}

export interface Strapi_RelationResponse_Currency extends Strapi_Meta {
  data: StrapiCurrency
}

export interface Strapi_ListResponse_Currency extends Strapi_Meta {
  data: StrapiCurrency[]
}

export interface Strapi_PostResponse_Currency extends Strapi_Meta {
  data: StrapiCurrency
}

export interface CurrencyResponse {
  id: number
  code: string
  description: string
  localeCode: string
}

export interface ListCurrencyResponse {
  data: CurrencyResponse[]
  pagination: PaginationResponse
}
