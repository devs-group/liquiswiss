import type {User, UserPasswordFormData, UserProfileFormData} from "~/models/auth";
import type {TransactionFormData, TransactionResponse} from "~/models/transaction";

export default function useAuth() {
    const user = useState<User|null>('user');
    const hasFetchedInitially = useState('hasFetchedInitially', () => false)

    const login = async (email: string, password: string): Promise<boolean> => {
        try {
            const data = await $fetch('/api/auth/login', {
                method: 'POST',
                body: {
                    email: email,
                    password: password,
                }
            });
            return true
        } catch (err) {
            return false
        }
    }

    const register = async (email: string, password: string): Promise<boolean> => {
        try {
            await $fetch('/api/auth/register', {
                method: 'POST',
                body: {
                    email: email,
                    password: password,
                }
            });
            return true
        } catch (err) {
            return false
        }
    }

    const logout = async () => {
        try {
            await $fetch('/api/auth/logout', {
                method: 'GET',
            });
            user.value = null
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

    const useFetchGetProfile = async () => {
        hasFetchedInitially.value = true
        const {data, error} = await useFetch('/api/profile', {
            method: 'GET',
            retry: false,
        });
        if (error.value) {
            console.error(error.value)
            return Promise.reject('Benutzer konnte nicht geladen werden')
        }
        user.value = data.value
    }

    const updateProfile = async (payload: UserProfileFormData) => {
        try {
            user.value = await $fetch<User>(`/api/profile`, {
                method: 'PATCH',
                body: {
                    ...payload,
                },
            });
        } catch (err) {
            return Promise.reject('Fehler beim Aktualisieren des Profils')
        }
    }

    const updatePassword = async (payload: UserPasswordFormData) => {
        try {
            await $fetch(`/api/profile/password`, {
                method: 'POST',
                body: {
                    ...payload,
                },
            });
        } catch (err) {
            return Promise.reject('Fehler beim Ändern des Password')
        }
    }

    const isAuthenticated = computed(() => !!user.value);

    return {
        user,
        hasFetchedInitially,
        isAuthenticated,
        login,
        register,
        logout,
        getAccessToken,
        useFetchGetProfile,
        updateProfile,
        updatePassword,
    };
}
