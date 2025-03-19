import { components } from "@/lib/api";
import { apiClient } from "@/lib/apiClient";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";

export function useFeeds() {
    return useQuery({
        queryKey: ["feeds"],
        queryFn: async () => {
            const res = await apiClient.GET("/feeds")

            if (res.error) {
                throw new Error("Failed to fetch feeds")
            }

            return res.data
        }
    })
}

export function useDeleteFeed(feedId: number) {

    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: async () => {
            const res = await apiClient.DELETE("/feeds/{id}", { params: { path: { id: feedId } } })

            if (res.error) {
                throw new Error("Failed to delete feed")
            }


            return res.data
        },
        onMutate() {
            const prevFeeds = queryClient.getQueryData(["feeds"]) as { id: number }[]

            return {
                prevFeeds
            }
        },
        onSuccess(data, variables, context) {
            queryClient.setQueryData(["feeds"], context.prevFeeds.filter(feed => feed.id !== feedId))
        },
    })
}

export function useAddFeed() {

    const client = useQueryClient();

    return useMutation({
        mutationFn: async (data: components["schemas"]["createFeed"]) => {
            const res = await apiClient.POST("/feeds", { body: data })

            if (res.error) {
                throw new Error("Failed to add feed")
            }

            return res.data
        },
        onSuccess(data, variables, context) {
            client.invalidateQueries({
                queryKey: ["feeds"]
            })
        },
    })
}