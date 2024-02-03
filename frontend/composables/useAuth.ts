import { ref } from 'vue';
import type {User} from "~/models/auth";

const user = ref<User>();

export default function useAuth() {
    const login = async (email: string, password: string): Promise<boolean> => {
        try {
            const {data} = await useFetch('/api/auth/login', {
                method: 'POST',
                body: {
                    email: email,
                    password: password,
                }
            });
            if (data.value) {
                return true
            }
        } catch (error) {
            console.error('Error logging in:', error);
        }
        return false
    }

    const register = async (email: string, password: string): Promise<boolean> => {
        try {
            const {data} = await useFetch('/api/auth/register', {
                method: 'POST',
                body: {
                    email: email,
                    password: password,
                }
            });
            if (data.value) {
                return true
            }
        } catch (error) {
            console.error('Error registering:', error);
        }
        return false
    }

    const logout = async () => {
        try {
            await useFetch('/api/auth/logout', {
                method: 'GET',
            });
        } catch (error) {
            console.error('Error logging out:', error);
        }
    }

    // Only used to regain the AccessToken in case it expires
    const getAccessToken = async () => {
        try {
            await $fetch('/api/access-token', {
                method: 'GET',
            });
        } catch (error) {
            console.error('Error getting access token:', error);
        }
    }

    const getProfile = async (serverSide: boolean) => {
        try {
            const {data} = await useFetch('/api/profile', {
                method: 'GET',
                server: serverSide,
            });
            user.value = data.value
        } catch (error) {
            console.error('Error getting profile:', error);
        }
    }

    const getCategories = async (page: number = 1, limit: number = 10) => {
        try {
            const {data} = await useFetch('/api/categories', {
                method: 'GET',
                query: {
                    page,
                    limit,
                }
            });
        } catch (error) {
            console.error('Error listing categories:', error);
        }
    }

    const getCategory = async (id: string) => {
        try {
            const {data} = await useFetch(`/api/categories/${id}`, {
                method: 'GET',
            });
        } catch (error) {
            console.error('Error getting category:', error);
        }
    }

    return {
        user,
        login,
        register,
        logout,
        getAccessToken,
        getProfile,
        getCategories,
        getCategory,
    };
}
