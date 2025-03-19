import { apiClient } from "@/lib/apiClient"
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query"

export function useAuthMe() {
    const { data, isLoading, error, isError } = useQuery({
        queryKey: ["auth/me"],
        queryFn: async () => {
            const res = await apiClient.GET("/auth/me")


            return {
                user: res.data,
                errorMessage: res.error,
                response: res.response
            }
        }
    })


    return {
        loading: isLoading,
        user: data?.user,
        loggedIn: data?.response?.status === 200,
    }
}

export function useLogout() {

    const queryClient = useQueryClient()

    const mutation = useMutation({
        mutationFn: async () => {
            return await apiClient.POST("/auth/logout")
        },
        onSuccess: async () => {
            await queryClient.invalidateQueries({
                queryKey: ["auth/me"]
            })
        }
    })

    return mutation
}