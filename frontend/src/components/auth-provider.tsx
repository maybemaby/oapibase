import { useAuthMe } from "@/hooks/auth";
import { apiClient } from "@/lib/apiClient";
import { createContext, useContext } from "react";

interface AuthContextData {
    userId: number | null;
    loggedIn?: boolean;
}

const authCtx = createContext<AuthContextData>({
    userId: null,
    loggedIn: undefined
})

export const AuthProvider = ({ children }: { children: React.ReactNode }) => {
    const {
        user, loading, loggedIn
    } = useAuthMe();

    return (
        <authCtx.Provider value={{ userId: user?.id ?? null, loggedIn: loading ? undefined : loggedIn }}>
            {children}
        </authCtx.Provider>
    )
}

export function useAuth() {
    const ctx = useContext(authCtx)

    return ctx
}