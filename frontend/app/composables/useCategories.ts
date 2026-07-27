import type { CategoryFormData, CategoryResponse, ListCategoryResponse } from '~/models/category'

export default function useCategories() {
  const { categories } = useGlobalData()

  const listCategories = async () => {
    try {
      const data = await $fetch<ListCategoryResponse>('/api/categories', {
        method: 'GET',
        query: {
          page: 1,
          limit: 100,
        },
      })
      categories.value = data.data ?? []
    }
    catch {
      return Promise.reject('Kategorien konnten nicht geladen werden')
    }
  }

  const getCategory = async (categoryID: number) => {
    try {
      return await $fetch<CategoryResponse>(`/api/categories/${categoryID}`, {
        method: 'GET',
      })
    }
    catch {
      return Promise.reject(`Kategorie mit ID "${categoryID}" konnte nicht geladen werden`)
    }
  }

  const createCategory = async (payload: CategoryFormData) => {
    try {
      const category = await $fetch<CategoryResponse>('/api/categories', {
        method: 'POST',
        body: payload,
      })
      await listCategories()
      return category
    }
    catch {
      return Promise.reject('Kategorie konnte nicht angelegt werden')
    }
  }

  const updateCategory = async (payload: CategoryFormData) => {
    try {
      const category = await $fetch<CategoryResponse>(`/api/categories/${payload.id}`, {
        method: 'PATCH',
        body: { name: payload.name },
      })
      await listCategories()
      return category
    }
    catch {
      return Promise.reject('Kategorie konnte nicht bearbeitet werden')
    }
  }

  // Rejects with { status, message } so callers can offer the reassign flow
  // on 409 (category still used by transactions)
  const deleteCategory = async (categoryID: number) => {
    try {
      await $fetch(`/api/categories/${categoryID}`, {
        method: 'DELETE',
      })
      await listCategories()
    }
    catch (err: unknown) {
      const fetchError = err as { status?: number, data?: { error?: string } }
      return Promise.reject({
        status: fetchError.status,
        message: fetchError.data?.error ?? 'Kategorie konnte nicht gelöscht werden',
      })
    }
  }

  // Moves all transactions from one category to another (prelude to delete)
  const reassignCategoryTransactions = async (fromCategoryID: number, toCategoryID: number) => {
    try {
      const result = await $fetch<{ affected: number }>(`/api/categories/${fromCategoryID}/reassign`, {
        method: 'POST',
        body: { targetId: toCategoryID },
      })
      return result.affected
    }
    catch {
      return Promise.reject('Transaktionen konnten nicht umgehängt werden')
    }
  }

  return {
    categories,
    listCategories,
    getCategory,
    createCategory,
    updateCategory,
    deleteCategory,
    reassignCategoryTransactions,
  }
}
