import createFetchClient from "openapi-fetch"
import { paths } from "./api"

export const apiClient = createFetchClient<paths>({
    baseUrl: "http://localhost:8000",
    credentials: "include",
})

