export interface User {
    id: number;
    name: string;
    email: string;
}

export interface Register {
    email: string;
    password: string;
}

export interface Login {
    email: string;
    password: string;
}

export interface UserProfileFormData {
    id: number;
    name: string;
    email: string;
}

export interface UserPasswordFormData {
    password: string;
    passwordRepeat: string;
}
