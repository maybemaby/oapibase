import React, { useMemo } from "react";
import ReactDOM from "react-dom/client";
import { RouterProvider, createRouter } from "@tanstack/react-router";
import { routeTree } from "./routeTree.gen";
import { useAuthMe } from "./hooks/auth";
import { AuthProvider } from "./components/auth-provider";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";

// Set up a Router instance
const router = createRouter({
  routeTree,
  defaultPreload: "intent",
  context: {
    auth: null,
  },
});

// Register things for typesafety
declare module "@tanstack/react-router" {
  interface Register {
    router: typeof router;
  }
}

const queryClient = new QueryClient();


function InnerApp() {
  const auth = useAuthMe();

  const authCtx = useMemo(() => {
    if (auth === null || !auth.loggedIn || auth.loading || !auth.user?.id) {
      return null
    }

    return {
      userId: auth.user.id.toString()
    }
  }, [auth])

  return <RouterProvider router={router} context={{ auth: authCtx }}></RouterProvider >;
}

function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <AuthProvider>
        <InnerApp />
      </AuthProvider>
    </QueryClientProvider>
  )
}

const rootElement = document.getElementById("app")!;

if (!rootElement.innerHTML) {
  const root = ReactDOM.createRoot(rootElement);
  root.render(<App />);
}
